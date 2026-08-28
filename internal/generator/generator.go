// Package generator は catchup-news からのネタ収集 → Grok 解説生成 →
// （dry-run でなければ）X 投稿までの一連の処理をまとめる。
// cmd/post（CLI）と cmd/web（HTTP API）の両方から呼び出される。
package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"catchup-x-post/internal/copywriter"
	"catchup-x-post/internal/grok"
	"catchup-x-post/internal/history"
	"catchup-x-post/internal/newsclient"
	"catchup-x-post/internal/picker"
	"catchup-x-post/internal/topics"
	"catchup-x-post/internal/xclient"
)

const (
	maxNewsAttempts = 5
	// MaxArticleCount は 1 回の実行で生成できる記事数の上限。
	MaxArticleCount = 20
	timeLayout      = "200601021504"
)

// Options は Run 1回分の実行パラメータ。
type Options struct {
	// DryRun が true の場合、記事は output/ に保存するのみで X へは投稿しない。
	DryRun bool
	// Count は生成する記事数（1〜MaxArticleCount）。
	Count int
	// Topic は catchup-news の検索トピック（空なら catchup-news 側で自動選定）。
	Topic string
}

// Article は生成・保存済みの記事1件。
type Article struct {
	ID    string
	Title string
	Path  string
}

// Result は Run の実行結果。エラー時でもそれまでに生成できた Articles を保持する。
type Result struct {
	Articles []Article
}

type logFunc func(format string, args ...any)

// Run はネタ収集から記事保存（・任意で X 投稿）までを実行する。
func Run(opts Options) (*Result, error) {
	grokKey := os.Getenv("GROK_API_KEY")
	if grokKey == "" {
		return nil, fmt.Errorf("GROK_API_KEY is not set")
	}

	model := os.Getenv("GROK_MODEL")
	if model == "" {
		model = "grok-4"
	}

	baseURL := os.Getenv("CATCHUP_NEWS_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	logsDir := os.Getenv("LOGS_DIR")
	if logsDir == "" {
		logsDir = "./logs"
	}
	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "./output"
	}
	historyDir := os.Getenv("HISTORY_DIR")
	if historyDir == "" {
		historyDir = "./history"
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	ts := time.Now().Format(timeLayout)
	logPath := filepath.Join(logsDir, ts+".log")

	store, err := history.NewStore(historyDir)
	if err != nil {
		return nil, err
	}
	topicStore, err := topics.NewStore(logsDir)
	if err != nil {
		return nil, err
	}

	logf := func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		log.Print(msg)
		appendLog(logPath, msg)
	}

	topic := strings.TrimSpace(opts.Topic)
	logf("start dry_run=%v count=%d topic=%q history=%d topics_history=%d", opts.DryRun, opts.Count, topic, store.Count(), topicStore.Count())

	client := newsclient.NewClient(baseURL)
	if topic != "" {
		client.SetTopic(topic)
	}
	grokClient := grok.NewClient(grokKey)
	if v := strings.TrimSpace(os.Getenv("MAX_GROK_CALLS")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_GROK_CALLS %q: %w", v, err)
		}
		grokClient.SetMaxCalls(n)
		logf("max grok calls=%d", n)
	}

	var excludeURLs []string
	result := &Result{}
	runTS := time.Now().Format(timeLayout)

	var xc *xclient.Client
	if !opts.DryRun {
		xc = xclient.NewClient(
			os.Getenv("X_API_KEY"),
			os.Getenv("X_API_SECRET"),
			os.Getenv("X_ACCESS_TOKEN"),
			os.Getenv("X_ACCESS_SECRET"),
		)
		if !xc.Configured() {
			return nil, fmt.Errorf("X API credentials are not fully set")
		}
	}

	for i := 0; i < opts.Count; i++ {
		articleNum := i + 1
		logf("--- article %d/%d ---", articleNum, opts.Count)

		item, next, err := pickFreshItem(client, store, topicStore, excludeURLs, logf)
		excludeURLs = next
		if err != nil {
			return result, fmt.Errorf("article %d/%d: %w", articleNum, opts.Count, err)
		}
		excludeURLs = append(excludeURLs, item.URL)

		logf("picked: %s (%s)", item.Title, item.URL)

		out, err := copywriter.Generate(grokClient, model, item)
		if err != nil {
			return result, fmt.Errorf("article %d/%d copywriter: %w", articleNum, opts.Count, err)
		}
		logf("article (%d chars), tweet (%d chars)", len([]rune(out.Article)), len([]rune(out.Tweet)))
		logf("tweet: %s", out.Tweet)

		articlePath := articleOutputPath(outputDir, runTS, articleNum, opts.Count)
		doc := copywriter.FormatDocument(item, out)
		if err := os.WriteFile(articlePath, []byte(doc), 0644); err != nil {
			return result, err
		}
		logf("saved %s", articlePath)

		id := strings.TrimSuffix(filepath.Base(articlePath), ".txt")
		result.Articles = append(result.Articles, Article{ID: id, Title: item.Title, Path: articlePath})

		if err := store.Add(item.URL, item.Title, item.Topic); err != nil {
			return result, err
		}
		if err := topicStore.Append(item.Title, item.Topic, item.URL); err != nil {
			return result, err
		}
		logf("recorded history + topics_history: %s", topics.FormatLine(item.Title, item.Topic, item.URL))

		if opts.DryRun {
			continue
		}

		tweetID, err := xc.PostTweet(out.Tweet)
		if err != nil {
			return result, fmt.Errorf("article %d/%d X post: %w (saved: %s)", articleNum, opts.Count, err, articlePath)
		}
		logf("posted tweet id=%s", tweetID)
	}

	if opts.DryRun {
		logf("dry-run complete (%d articles, no X post). Review output/ then run with -dry-run=false", len(result.Articles))
	} else {
		logf("complete (%d articles posted)", len(result.Articles))
	}
	st := grokClient.GetStats()
	logf("grok_calls=%d by_tag=%v", st.CallsTotal, st.CallsByTag)
	return result, nil
}

func pickFreshItem(
	client *newsclient.Client,
	store *history.Store,
	topicStore *topics.Store,
	excludeURLs []string,
	logf logFunc,
) (*newsclient.NewsItem, []string, error) {
	for attempt := 1; attempt <= maxNewsAttempts; attempt++ {
		news, err := client.Fetch(excludeURLs)
		if err != nil {
			return nil, excludeURLs, err
		}
		logf("attempt %d: discovered topic=%q items=%d", attempt, news.Topic, len(news.News))

		if len(news.News) == 0 {
			return nil, excludeURLs, fmt.Errorf("no news items returned (attempt %d)", attempt)
		}
		candidate := &news.News[0]
		if store.IsSimilar(candidate.URL, candidate.Title, candidate.Topic) {
			logf("similar to history, refetching (title=%q url=%s)", candidate.Title, candidate.URL)
			excludeURLs = append(excludeURLs, candidate.URL)
			continue
		}
		if topicStore.IsDuplicate(candidate.Title, candidate.Topic) {
			logf("topic in topics_history, refetching (topic=%q url=%s)", candidate.Topic, candidate.URL)
			excludeURLs = append(excludeURLs, candidate.URL)
			continue
		}
		picked, err := picker.Pick(news.News, store)
		if err == nil {
			return picked, excludeURLs, nil
		}
		logf("pick failed: %v (title=%q)", err, candidate.Title)
		excludeURLs = append(excludeURLs, candidate.URL)
	}
	return nil, excludeURLs, fmt.Errorf("could not find fresh topic after %d attempts", maxNewsAttempts)
}

func articleOutputPath(outputDir, runTS string, index, total int) string {
	name := runTS + ".txt"
	if total > 1 {
		name = fmt.Sprintf("%s_%02d.txt", runTS, index)
	}
	return filepath.Join(outputDir, name)
}

func appendLog(path, line string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format(time.RFC3339), line)
}
