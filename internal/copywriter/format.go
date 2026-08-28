package copywriter

import (
	"strings"

	"catchup-x-post/internal/newsclient"
)

// FormatDocument は output/*.txt 用の記事フォーマットを組み立てる
func FormatDocument(item *newsclient.NewsItem, out *Output) string {
	topicLine := strings.TrimSpace(item.Topic)
	if topicLine == "" {
		topicLine = strings.TrimSpace(item.Title)
	}

	body := stripSummarySection(strings.TrimSpace(out.Article))
	if body == "" {
		body = buildArticleFromItem(item)
	}
	body = ensureSections(body, item)
	body = stripMeritsPlaceholder(body)

	var b strings.Builder
	b.WriteString("トピック: ")
	b.WriteString(topicLine)
	b.WriteString("\n\n")
	b.WriteString(body)
	b.WriteString("\n\n参考: ")
	b.WriteString(strings.TrimSpace(item.URL))
	b.WriteString("\n\n【X投稿文案】（内容のみ・280字以内）\n")
	b.WriteString(strings.TrimSpace(out.Tweet))
	return b.String()
}

func stripSummarySection(article string) string {
	const marker = "\n\n【まとめ】"
	if idx := strings.Index(article, marker); idx >= 0 {
		return strings.TrimSpace(article[:idx])
	}
	if strings.HasPrefix(article, "【まとめ】") {
		return ""
	}
	return article
}

func ensureSections(body string, item *newsclient.NewsItem) string {
	if strings.Contains(body, "【概要】") {
		return body
	}
	return buildArticleFromItem(item)
}

func stripMeritsPlaceholder(body string) string {
	if !strings.Contains(body, meritsPlaceholder) {
		return body
	}
	// プレースホルダーだけのメリット節は除去（Generate 側で再生成済みのはず）
	if i := strings.Index(body, "【メリット・デメリット】"); i >= 0 {
		end := nextSectionIndex(body, i+len("【メリット・デメリット】"))
		section := body[i:]
		if end > i {
			section = body[i:end]
		}
		if strings.Contains(section, meritsPlaceholder) && !hasValidMeritsDemerits(body) {
			if end < 0 {
				return strings.TrimSpace(body[:i])
			}
			return strings.TrimSpace(body[:i]) + "\n\n" + strings.TrimSpace(body[end:])
		}
	}
	return body
}
