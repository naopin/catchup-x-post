package copywriter

import (
	"fmt"
	"strings"

	"catchup-x-post/internal/grok"
	"catchup-x-post/internal/newsclient"
)

const meritsPlaceholder = "（文案生成時に追記）"

// ensureMeritsDemerits は記事に有効なメリット・デメリットがなければ Grok で追記する
func ensureMeritsDemerits(client *grok.Client, model string, item *newsclient.NewsItem, article string) (string, error) {
	if hasValidMeritsDemerits(article) {
		return article, nil
	}
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		block, err := generateMeritsDemeritsBlock(client, model, item, article)
		if err != nil {
			lastErr = err
			continue
		}
		merged := insertMeritsBlock(article, block)
		if hasValidMeritsDemerits(merged) {
			return merged, nil
		}
		lastErr = fmt.Errorf("invalid merits/demerits block")
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("could not generate merits/demerits")
	}
	return article, lastErr
}

func hasValidMeritsDemerits(article string) bool {
	article = strings.TrimSpace(article)
	if !strings.Contains(article, "メリット:") || !strings.Contains(article, "デメリット:") {
		return false
	}
	if strings.Contains(article, meritsPlaceholder) {
		return false
	}
	return countMeritBullets(article) >= 2 && countDemeritBullets(article) >= 2
}

func countMeritBullets(article string) int {
	return countBulletsInSection(article, "メリット:", "デメリット:")
}

func countDemeritBullets(article string) int {
	start := strings.Index(article, "デメリット:")
	if start < 0 {
		return 0
	}
	rest := article[start+len("デメリット:"):]
	end := strings.Index(rest, "\n\n【")
	if end >= 0 {
		rest = rest[:end]
	}
	return countBulletLines(rest)
}

func countBulletsInSection(article, startMarker, endMarker string) int {
	start := strings.Index(article, startMarker)
	if start < 0 {
		return 0
	}
	rest := article[start+len(startMarker):]
	if end := strings.Index(rest, endMarker); end >= 0 {
		rest = rest[:end]
	}
	return countBulletLines(rest)
}

func countBulletLines(s string) int {
	n := 0
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "・") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "•") {
			if !strings.Contains(line, meritsPlaceholder) {
				n++
			}
		}
	}
	return n
}

func generateMeritsDemeritsBlock(client *grok.Client, model string, item *newsclient.NewsItem, article string) (string, error) {
	snippet := article
	if len([]rune(snippet)) > 1200 {
		snippet = string([]rune(snippet)[:1200])
	}
	prompt := fmt.Sprintf(`あなたはエンジニア向け技術メディアの編集者です。以下のネタについて【メリット・デメリット】ブロックだけを日本語で書いてください。

【ネタ】
タイトル: %s
概要: %s
解説: %s
トピック: %s

【記事抜粋（参考）】
%s

【出力形式（厳守）】
- 先頭行は必ず「【メリット・デメリット】」
- 次行「メリット:」のあと、箇条書き2〜4項目（各行先頭は「・」）
- 空行1つ
- 「デメリット:」のあと、箇条書き2〜4項目（各行先頭は「・」）
- エンジニアの採用・導入判断に役立つ具体（コスト・互換性・運用・リスク）。宣伝のみ禁止
- それ以外の見出し・説明・JSON・コードブロックは書かない`,
		item.Title, item.Summary, item.Explanation, item.Topic, snippet,
	)

	resp, err := client.ChatTagged("merits", model, prompt)
	if err != nil {
		return "", err
	}
	return normalizeMeritsBlock(stripCodeFences(strings.TrimSpace(resp.Text()))), nil
}

func normalizeMeritsBlock(block string) string {
	block = strings.TrimSpace(block)
	if !strings.HasPrefix(block, "【メリット・デメリット】") {
		if strings.Contains(block, "メリット:") {
			block = "【メリット・デメリット】\n" + block
		}
	}
	return block
}

func insertMeritsBlock(article, block string) string {
	block = strings.TrimSpace(block)
	if block == "" {
		return article
	}

	if i := strings.Index(article, "【メリット・デメリット】"); i >= 0 {
		end := nextSectionIndex(article, i+len("【メリット・デメリット】"))
		if end < 0 {
			return strings.TrimSpace(article[:i]) + "\n\n" + block
		}
		return strings.TrimSpace(article[:i]) + "\n\n" + block + "\n\n" + strings.TrimSpace(article[end:])
	}

	// 【概要】の直後に挿入
	const summaryEnd = "【概要】"
	if i := strings.Index(article, summaryEnd); i >= 0 {
		bodyStart := i + len(summaryEnd)
		end := nextSectionIndex(article, bodyStart)
		if end < 0 {
			return strings.TrimSpace(article) + "\n\n" + block
		}
		return strings.TrimSpace(article[:end]) + "\n\n" + block + "\n\n" + strings.TrimSpace(article[end:])
	}

	return block + "\n\n" + strings.TrimSpace(article)
}

func nextSectionIndex(article string, from int) int {
	if from < 0 || from >= len(article) {
		return -1
	}
	rest := article[from:]
	idx := strings.Index(rest, "\n\n【")
	if idx < 0 {
		return -1
	}
	return from + idx + 2 // skip leading \n\n
}
