package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"catchup-x-post/internal/copywriter"
	"catchup-x-post/internal/grok"
	"catchup-x-post/internal/history"
	"catchup-x-post/internal/newsclient"
	"catchup-x-post/internal/picker"
	"catchup-x-post/internal/topics"
	"catchup-x-post/internal/xclient"
)

const (
	maxNewsAttempts  = 5
	maxArticleCount  = 20
	defaultLogLayout = "200601021504"
)

func main() {
	dryRun := flag.Bool("dry-run", true, "save artifacts only, do not post to X")
	countFlag := flag.Int("count", 0, "number of articles to generate (also ARTICLE_COUNT env)")
	flag.Parse()

	_ = godotenv.Load()

	count, err := resolveArticleCount(*countFlag)
	if err != nil {
		log.Printf("post: %v", err)
		os.Exit(1)
	}

	if err := run(*dryRun, count); err != nil {
		log.Printf("post: %v", err)
		if looksLikeGrokCreditsExceeded(err) {
			log.Print("post: Grok API のクレジット枯渇/上限到達が原因の可能性があります。")
			log.Print("post: 対処: (1) X.AI 側でクレジット購入 or 月次上限を引き上げ (2) `GROK_API_KEY` を別チーム/別キーに切替")
			log.Print("post: 参考: `.env` の `GROK_API_KEY` / `GROK_MODEL` と、catchup-news のログも確認してください。")
		}
		os.Exit(1)
	}
}

func looksLikeGrokCreditsExceeded(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	// x.ai の典型的な文言（現状の catchup-news が 500 の中に文字列として包むケースもある）
	return strings.Contains(s, "used all available credits") ||
		strings.Contains(s, "monthly spending limit") ||
		strings.Contains(s, "does not have permission to execute the specified operation") ||
		strings.Contains(s, `"code":"The caller does not have permission to execute the specified operation"`)
}

func resolveArticleCount(flagCount int) (int, error) {
	n := flagCount
	if n <= 0 {
		if v := strings.TrimSpace(os.Getenv("ARTICLE_COUNT")); v != "" {
			parsed, err := strconv.Atoi(v)
			if err != nil {
				return 0, fmt.Errorf("invalid ARTICLE_COUNT %q: %w", v, err)
			}
			n = parsed
		}
	}
	if n <= 0 {
		n = 1
	}
	if n > maxArticleCount {
		return 0, fmt.Errorf("count %d exceeds maximum %d", n, maxArticleCount)
	}
	return n, nil
}

func run(dryRun bool, count int) error {
	grokKey := os.Getenv("GROK_API_KEY")
	if grokKey == "" {
		return fmt.Errorf("GROK_API_KEY is not set")
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
		return err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	ts := time.Now().Format("200601021504")
	logPath := filepath.Join(logsDir, ts+".log")

	store, err := history.NewStore(historyDir)
	if err != nil {
		return err
	}
	topicStore, err := topics.NewStore(logsDir)
	if err != nil {
		return err
	}

	logf := func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		log.Print(msg)
		appendLog(logPath, msg)
	}

	topic := strings.TrimSpace(os.Getenv("TOPIC"))
	logf("start dry_run=%v count=%d topic=%q history=%d topics_history=%d", dryRun, count, topic, store.Count(), topicStore.Count())

	client := newsclient.NewClient(baseURL)
	if topic != "" {
		client.SetTopic(topic)
	}
	grokClient := grok.NewClient(grokKey)
	if v := strings.TrimSpace(os.Getenv("MAX_GROK_CALLS")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid MAX_GROK_CALLS %q: %w", v, err)
		}
		grokClient.SetMaxCalls(n)
		logf("max grok calls=%d", n)
	}
	var excludeURLs []string
	var savedPaths []string
	runTS := time.Now().Format(defaultLogLayout)

	var xc *xclient.Client
	if !dryRun {
		xc = xclient.NewClient(
			os.Getenv("X_API_KEY"),
			os.Getenv("X_API_SECRET"),
			os.Getenv("X_ACCESS_TOKEN"),
			os.Getenv("X_ACCESS_SECRET"),
		)
		if !xc.Configured() {
			return fmt.Errorf("X API credentials are not fully set")
		}
	}

	for i := 0; i < count; i++ {
		articleNum := i + 1
		logf("--- article %d/%d ---", articleNum, count)

		item, excludeURLs, err := pickFreshItem(client, store, topicStore, excludeURLs, logf)
		if err != nil {
			if len(savedPaths) > 0 {
				return fmt.Errorf("article %d/%d: %w (saved: %s)", articleNum, count, err, strings.Join(savedPaths, ", "))
			}
			return fmt.Errorf("article %d/%d: %w", articleNum, count, err)
		}
		excludeURLs = append(excludeURLs, item.URL)

		logf("picked: %s (%s)", item.Title, item.URL)

		out, err := copywriter.Generate(grokClient, model, item)
		if err != nil {
			return fmt.Errorf("article %d/%d copywriter: %w", articleNum, count, err)
		}
		logf("article (%d chars), tweet (%d chars)", len([]rune(out.Article)), len([]rune(out.Tweet)))
		logf("tweet: %s", out.Tweet)

		articlePath := articleOutputPath(outputDir, runTS, articleNum, count)
		doc := copywriter.FormatDocument(item, out)
		if err := os.WriteFile(articlePath, []byte(doc), 0644); err != nil {
			return err
		}
		logf("saved %s", articlePath)
		savedPaths = append(savedPaths, articlePath)

		if err := store.Add(item.URL, item.Title, item.Topic); err != nil {
			return err
		}
		if err := topicStore.Append(item.Title, item.Topic, item.URL); err != nil {
			return err
		}
		logf("recorded history + topics_history: %s", topics.FormatLine(item.Title, item.Topic, item.URL))

		if dryRun {
			continue
		}

		tweetID, err := xc.PostTweet(out.Tweet)
		if err != nil {
			return fmt.Errorf("article %d/%d X post: %w (saved: %s)", articleNum, count, err, articlePath)
		}
		logf("posted tweet id=%s", tweetID)
	}

	if dryRun {
		logf("dry-run complete (%d articles, no X post). Review output/ then run with -dry-run=false", len(savedPaths))
	} else {
		logf("complete (%d articles posted)", len(savedPaths))
	}
	st := grokClient.GetStats()
	logf("grok_calls=%d by_tag=%v", st.CallsTotal, st.CallsByTag)
	return nil
}

type logFunc func(format string, args ...any)

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
