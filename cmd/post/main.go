package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"catchup-x-post/internal/generator"
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

	topic := strings.TrimSpace(os.Getenv("TOPIC"))
	if _, err := generator.Run(generator.Options{DryRun: *dryRun, Count: count, Topic: topic}); err != nil {
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
	if n > generator.MaxArticleCount {
		return 0, fmt.Errorf("count %d exceeds maximum %d", n, generator.MaxArticleCount)
	}
	return n, nil
}
