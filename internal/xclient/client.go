package xclient

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const tweetURL = "https://api.twitter.com/2/tweets"

type Client struct {
	apiKey       string
	apiSecret    string
	accessToken  string
	accessSecret string
	httpClient   *http.Client
}

func NewClient(apiKey, apiSecret, accessToken, accessSecret string) *Client {
	return &Client{
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		accessToken:  accessToken,
		accessSecret: accessSecret,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Configured() bool {
	return c.apiKey != "" && c.apiSecret != "" && c.accessToken != "" && c.accessSecret != ""
}

type tweetRequest struct {
	Text string `json:"text"`
}

type tweetResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
	Detail string `json:"detail"`
	Title  string `json:"title"`
}

func (c *Client) PostTweet(text string) (string, error) {
	body := tweetRequest{Text: text}
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, tweetURL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.sign(req, nil); err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tweet failed %d: %s", resp.StatusCode, raw)
	}

	var tr tweetResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return "", fmt.Errorf("decode tweet response: %w", err)
	}
	if tr.Data.ID == "" {
		return "", fmt.Errorf("tweet response missing id: %s", raw)
	}
	return tr.Data.ID, nil
}

func (c *Client) sign(req *http.Request, extra url.Values) error {
	oauth := url.Values{
		"oauth_consumer_key":     {c.apiKey},
		"oauth_nonce":            {fmt.Sprintf("%d", time.Now().UnixNano())},
		"oauth_signature_method": {"HMAC-SHA1"},
		"oauth_timestamp":        {strconv.FormatInt(time.Now().Unix(), 10)},
		"oauth_token":            {c.accessToken},
		"oauth_version":          {"1.0"},
	}

	sigParams := url.Values{}
	for k, v := range oauth {
		sigParams[k] = v
	}
	if extra != nil {
		for k, v := range extra {
			sigParams[k] = v
		}
	}
	if req.URL.RawQuery != "" {
		q, _ := url.ParseQuery(req.URL.RawQuery)
		for k, v := range q {
			sigParams[k] = v
		}
	}

	base := strings.ToUpper(req.Method) + "&" +
		percentEncode(req.URL.Scheme+"://"+req.URL.Host+req.URL.Path) + "&" +
		percentEncode(sigParams.Encode())

	key := percentEncode(c.apiSecret) + "&" + percentEncode(c.accessSecret)
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(base))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	oauth.Set("oauth_signature", sig)
	req.Header.Set("Authorization", "OAuth "+encodeHeader(oauth))
	return nil
}

func encodeHeader(v url.Values) string {
	var parts []string
	for _, k := range sortedKeys(v) {
		parts = append(parts, percentEncode(k)+"=\""+percentEncode(v.Get(k))+"\"")
	}
	return strings.Join(parts, ", ")
}

func sortedKeys(v url.Values) []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func percentEncode(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' {
			buf.WriteByte(c)
		} else {
			buf.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return buf.String()
}
