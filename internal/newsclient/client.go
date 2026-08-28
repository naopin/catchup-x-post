package newsclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type NewsItem struct {
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	Explanation    string `json:"explanation"`
	UseCases       string `json:"use_cases"`
	Topic          string `json:"topic"`
	TrendingReason string `json:"trending_reason"`
	URL            string `json:"url,omitempty"`
}

type NewsResponse struct {
	News  []NewsItem `json:"news"`
	Topic string     `json:"topic"`
	Total int        `json:"total"`
}

type Client struct {
	baseURL    string
	topic      string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// SetTopic は catchup-news の検索トピックを指定する（空なら自動選定）。
func (c *Client) SetTopic(topic string) {
	c.topic = strings.TrimSpace(topic)
}

// Fetch は catchup-news に X 検索のトピック選定を依頼する。
// topic が設定されていればその分野に絞る。未設定なら自動選定。
// excludeURLs は過去に生成した X 投稿 URL（再選定の除外用）。
func (c *Client) Fetch(excludeURLs []string) (*NewsResponse, error) {
	u, err := url.Parse(c.baseURL + "/api/news")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if c.topic != "" {
		q.Set("topic", c.topic)
	}
	if len(excludeURLs) > 0 {
		q.Set("exclude_url", strings.Join(excludeURLs, ","))
	}
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Error != "" {
			// catchup-news が下流（例: Grok 403）を 500 の error 文字列に包むケースがあるため、
			// そのまま表示できるよう整形する。
			return nil, fmt.Errorf("news API %d: %s", resp.StatusCode, strings.TrimSpace(errBody.Error))
		}
		return nil, fmt.Errorf("news API status %d", resp.StatusCode)
	}

	var result NewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode news: %w", err)
	}
	return &result, nil
}
