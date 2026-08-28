package copywriter

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"catchup-x-post/internal/grok"
	"catchup-x-post/internal/newsclient"
)

const maxTweetRunes = 280

type Output struct {
	Tweet   string // X 投稿用（280字以内）
	Article string // プレビュー用の解説記事
}

type modelResult struct {
	Tweet   string `json:"tweet"`
	Article string `json:"article"`
}

func Generate(client *grok.Client, model string, item *newsclient.NewsItem) (*Output, error) {
	prompt := fmt.Sprintf(`あなたはエンジニア向け技術メディアの編集者です。以下のネタをもとに、(1) X用の短い投稿文案と (2) 解説・活用事例を含む読み物を日本語で書いてください。

【ネタ】
タイトル: %s
概要: %s
解説: %s
活用事例: %s
トピック: %s
注目理由: %s
参考URL: %s

【article の要件】
- 見出しは付けず、次の5ブロックを空行で区切る
  1) 「【概要】」で始める段落（3〜5文。何が起きたか）
  2) 「【メリット・デメリット】」で始めるブロック（必須・省略禁止）。1行目に「メリット:」、その下に箇条書き2〜4項目（先頭「・」）。空行のあと「デメリット:」、同様に2〜4項目。エンジニアが採用・導入・学習するときの判断材料（コスト・互換性・運用負荷・リスクも含む）。宣伝だけでなく実務のトレードオフを正直に書く。「（文案生成時に追記）」などのプレースホルダーは禁止
  3) 「【解説】」で始める段落（背景・仕組み・他技術との違い。8文程度）
  4) 「【活用事例】」で始める段落（エンジニアが明日から試せる具体例を4つ以上。1行1事例、先頭に「・」は付けない）
- 「【まとめ】」は書かない（参考URLも article 内には含めない）
- 検索結果の転載ではなく、読者に価値が伝わるよう自分の言葉で書く
- 「Xで話題」「Twitter」など情報源が X だとわかる表現は避ける
- 【メリット・デメリット】の箇条書きは、各行の先頭を必ず「・」にする（"-" や "*" にはしない）

【tweet の要件】
- 280文字以内（厳守）
- 1ポスト完結。スレッド前提にしない
- 概要＋解説の要点を1〜2文で凝縮し、最後に参考URLを1行で含める
- 絵文字は0〜2個、ハッシュタグは #エンジニア #技術 から最大2個

【出力の要件】
- 出力は有効なJSONのみ（前後に説明文を付けない）
- コードブロック（バッククォート3つ）で囲まない

有効なJSONのみ出力:
{"tweet": "...", "article": "..."}`,
		item.Title, item.Summary, item.Explanation, item.UseCases,
		item.Topic, item.TrendingReason, item.URL,
	)

	resp, err := client.ChatTagged("generate", model, prompt)
	if err != nil {
		return nil, err
	}

	content := stripCodeFences(strings.TrimSpace(resp.Text()))

	result, err := parseModelResult(content)
	if err != nil {
		return fallbackOutput(client, model, item, content)
	}

	tweet := normalizeTweet(strings.TrimSpace(result.Tweet), content, item.URL)
	article := strings.TrimSpace(result.Article)
	if tweet == "" {
		return fallbackOutput(client, model, item, content)
	}
	if article == "" {
		article = buildArticleFromItem(item)
	}
	article, err = ensureMeritsDemerits(client, model, item, article)
	if err != nil {
		return nil, fmt.Errorf("merits/demerits: %w", err)
	}
	return &Output{Tweet: tweet, Article: article}, nil
}

func parseModelResult(content string) (modelResult, error) {
	var result modelResult
	if err := parseJSON(content, &result); err != nil {
		return result, err
	}
	if strings.HasPrefix(strings.TrimSpace(result.Tweet), "{") {
		var nested modelResult
		if err := parseJSON(result.Tweet, &nested); err == nil {
			if strings.TrimSpace(nested.Tweet) != "" {
				result.Tweet = nested.Tweet
			}
			if strings.TrimSpace(nested.Article) != "" {
				result.Article = nested.Article
			}
		}
	}
	if strings.TrimSpace(result.Tweet) == "" && strings.TrimSpace(result.Article) == "" {
		return result, fmt.Errorf("empty fields")
	}
	return result, nil
}

func fallbackOutput(client *grok.Client, model string, item *newsclient.NewsItem, raw string) (*Output, error) {
	tweet := normalizeTweet("", raw, item.URL)
	article := buildArticleFromItem(item)
	if parsed, err := parseModelResult(raw); err == nil {
		if s := strings.TrimSpace(parsed.Article); s != "" {
			article = s
		}
	}
	article, err := ensureMeritsDemerits(client, model, item, article)
	if err != nil {
		return nil, err
	}
	return &Output{
		Tweet:   tweet,
		Article: article,
	}, nil
}

// normalizeTweet はパース結果や生レスポンスから X 用プレーンテキストを取り出す
func normalizeTweet(parsedTweet, rawContent, refURL string) string {
	candidates := []string{parsedTweet}
	if strings.HasPrefix(strings.TrimSpace(parsedTweet), "{") {
		var nested modelResult
		if err := parseJSON(parsedTweet, &nested); err == nil {
			candidates = append([]string{nested.Tweet}, candidates...)
		}
	}
	var fromRaw modelResult
	if err := parseJSON(rawContent, &fromRaw); err == nil {
		candidates = append(candidates, fromRaw.Tweet)
	}
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c == "" || strings.HasPrefix(c, "{") {
			continue
		}
		return ensureURL(trimToMax(c, maxTweetRunes), refURL)
	}
	if t := extractTweetFieldLoose(rawContent); t != "" {
		return ensureURL(trimToMax(t, maxTweetRunes), refURL)
	}
	return ""
}

// extractTweetFieldLoose は壊れた・切り詰められた JSON から tweet 文字列を抜く
func extractTweetFieldLoose(content string) string {
	const marker = `"tweet"`
	idx := strings.Index(content, marker)
	if idx < 0 {
		return ""
	}
	rest := content[idx+len(marker):]
	colon := strings.Index(rest, ":")
	if colon < 0 {
		return ""
	}
	rest = strings.TrimSpace(rest[colon+1:])
	if len(rest) == 0 {
		return ""
	}
	if rest[0] != '"' {
		return ""
	}
	var b strings.Builder
	for i := 1; i < len(rest); i++ {
		if rest[i] == '\\' && i+1 < len(rest) {
			switch rest[i+1] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case '"', '\\':
				b.WriteByte(rest[i+1])
			default:
				b.WriteByte(rest[i])
				b.WriteByte(rest[i+1])
			}
			i++
			continue
		}
		if rest[i] == '"' {
			return b.String()
		}
		b.WriteByte(rest[i])
	}
	return b.String()
}

func buildArticleFromItem(item *newsclient.NewsItem) string {
	var b strings.Builder
	b.WriteString("【概要】\n")
	b.WriteString(strings.TrimSpace(item.Summary))
	b.WriteString("\n\n【解説】\n")
	if s := strings.TrimSpace(item.Explanation); s != "" {
		b.WriteString(s)
	} else {
		b.WriteString(strings.TrimSpace(item.TrendingReason))
	}
	b.WriteString("\n\n【活用事例】\n")
	if s := strings.TrimSpace(item.UseCases); s != "" {
		b.WriteString(s)
	} else {
		b.WriteString("（活用事例はネタ取得時に生成されます）")
	}
	return b.String()
}

func stripCodeFences(content string) string {
	const fence = "```"
	if strings.HasPrefix(content, fence) {
		lines := strings.Split(content, "\n")
		if len(lines) > 2 {
			return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
		}
	}
	return content
}

func parseJSON(content string, v *modelResult) error {
	obj, err := extractFirstJSONObject(content)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(obj), v)
}

func extractFirstJSONObject(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", fmt.Errorf("no json object")
	}
	// モデルが余計な文を出した場合でも {"tweet":...,"article":...} を優先して拾う
	start := strings.Index(content, `{"tweet"`)
	if start < 0 {
		start = strings.Index(content, "{")
	}
	if start < 0 {
		return "", fmt.Errorf("no json object")
	}
	s := content[start:]

	depth := 0
	inStr := false
	esc := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inStr {
			if esc {
				esc = false
				continue
			}
			if ch == '\\' {
				esc = true
				continue
			}
			if ch == '"' {
				inStr = false
			}
			continue
		}
		switch ch {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[:i+1], nil
			}
		}
	}
	return "", fmt.Errorf("no complete json object")
}

func trimToMax(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max])
}

func ensureURL(text, refURL string) string {
	refURL = strings.TrimSpace(refURL)
	if refURL == "" || strings.Contains(text, refURL) {
		return text
	}
	combined := text + "\n" + refURL
	if utf8.RuneCountInString(combined) <= maxTweetRunes {
		return combined
	}
	return trimToMax(text, maxTweetRunes-len(refURL)-1) + "\n" + refURL
}
