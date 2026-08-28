// Package webarticle は output/*.txt（internal/copywriter.FormatDocument 準拠）を
// 読み込み、HTTP API 向けの一覧・詳細データに変換する。
package webarticle

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const dateLayout = "200601021504"

// Summary は記事一覧のレスポンス項目。
type Summary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Filename string `json:"filename"`
}

// Detail は記事詳細のレスポンス。
type Detail struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Tweet     string `json:"tweet"`
	SourceURL string `json:"source_url"`
	Date      string `json:"date"`
}

// List は dir 内の *.txt をファイル名降順（新しいものが先）で一覧化する。
// dir が存在しない場合は空の一覧を返す。
func List(dir string) ([]Summary, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []Summary{}, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))

	summaries := make([]Summary, 0, len(names))
	for _, name := range names {
		id := strings.TrimSuffix(name, ".txt")

		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		title, _, _, _, err := parse(string(raw))
		if err != nil {
			title = id
		}

		summaries = append(summaries, Summary{
			ID:       id,
			Title:    title,
			Date:     formatDate(id),
			Filename: name,
		})
	}
	return summaries, nil
}

// ErrNotFound は指定した id の記事が存在しない場合に返る。
var ErrNotFound = fmt.Errorf("article not found")

// Get は id（拡張子抜きのファイル名）に対応する記事詳細を返す。
func Get(dir, id string) (*Detail, error) {
	if id == "" || strings.ContainsAny(id, "/\\") || strings.Contains(id, "..") {
		return nil, ErrNotFound
	}

	path := filepath.Join(dir, id+".txt")
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	title, body, sourceURL, tweet, err := parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &Detail{
		ID:        id,
		Title:     title,
		Body:      body,
		Tweet:     tweet,
		SourceURL: sourceURL,
		Date:      formatDate(id),
	}, nil
}

// parse は copywriter.FormatDocument が出力する txt を title/body/sourceURL/tweet に分解する。
func parse(content string) (title, body, sourceURL, tweet string, err error) {
	const topicPrefix = "トピック: "
	if !strings.HasPrefix(content, topicPrefix) {
		return "", "", "", "", fmt.Errorf("missing topic line")
	}
	rest := content[len(topicPrefix):]

	nlIdx := strings.Index(rest, "\n")
	if nlIdx < 0 {
		return "", "", "", "", fmt.Errorf("missing body")
	}
	title = strings.TrimSpace(rest[:nlIdx])
	rest = strings.TrimPrefix(rest[nlIdx:], "\n\n")

	const refMarker = "\n\n参考: "
	refIdx := strings.Index(rest, refMarker)
	if refIdx < 0 {
		return "", "", "", "", fmt.Errorf("missing 参考 section")
	}
	body = strings.TrimSpace(rest[:refIdx])
	rest = rest[refIdx+len(refMarker):]

	const tweetMarker = "\n\n【X投稿文案】（内容のみ・280字以内）\n"
	tweetIdx := strings.Index(rest, tweetMarker)
	if tweetIdx < 0 {
		return "", "", "", "", fmt.Errorf("missing 【X投稿文案】 section")
	}
	sourceURL = strings.TrimSpace(rest[:tweetIdx])
	tweet = strings.TrimSpace(rest[tweetIdx+len(tweetMarker):])
	return title, body, sourceURL, tweet, nil
}

// formatDate は id 先頭の YYYYMMDDHHmm 部分を RFC3339 に変換する。
// パースできない場合は空文字を返す。
func formatDate(id string) string {
	datePart := id
	if i := strings.Index(id, "_"); i >= 0 {
		datePart = id[:i]
	}
	t, err := time.ParseInLocation(dateLayout, datePart, time.Local)
	if err != nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
