package grok

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const defaultBaseURL = "https://api.x.ai/v1"

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	mu         sync.Mutex
	callsTotal int
	callsByTag map[string]int
	maxCalls   int
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
		callsByTag: map[string]int{},
	}
}

// NewClientWithBaseURL はテスト用に baseURL を差し替える。
func NewClientWithBaseURL(apiKey, baseURL string) *Client {
	c := NewClient(apiKey)
	if baseURL != "" {
		c.baseURL = baseURL
	}
	return c
}

// SetMaxCalls は API の呼び出し上限（回数）を設定する。0以下は無制限。
func (c *Client) SetMaxCalls(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxCalls = n
}

type Stats struct {
	CallsTotal int
	CallsByTag map[string]int
	MaxCalls   int
}

func (c *Client) GetStats() Stats {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make(map[string]int, len(c.callsByTag))
	for k, v := range c.callsByTag {
		cp[k] = v
	}
	return Stats{
		CallsTotal: c.callsTotal,
		CallsByTag: cp,
		MaxCalls:   c.maxCalls,
	}
}

type InputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Tool struct {
	Type string `json:"type"`
}

type responsesRequest struct {
	Model string         `json:"model"`
	Input []InputMessage `json:"input"`
	Tools []Tool         `json:"tools,omitempty"`
}

type outputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type outputItem struct {
	Type    string          `json:"type"`
	Content []outputContent `json:"content"`
}

type Response struct {
	ID     string       `json:"id"`
	Output []outputItem `json:"output"`
}

func (r *Response) Text() string {
	for _, item := range r.Output {
		for _, c := range item.Content {
			if c.Type == "output_text" && c.Text != "" {
				return c.Text
			}
		}
	}
	return ""
}

func (c *Client) post(tag, path string, body responsesRequest) (*Response, error) {
	c.mu.Lock()
	if c.maxCalls > 0 && c.callsTotal >= c.maxCalls {
		c.mu.Unlock()
		return nil, fmt.Errorf("grok: max calls exceeded (max=%d)", c.maxCalls)
	}
	c.callsTotal++
	if tag == "" {
		tag = "unknown"
	}
	c.callsByTag[tag]++
	baseURL := c.baseURL
	c.mu.Unlock()

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, raw)
	}

	var sr Response
	if err := json.Unmarshal(raw, &sr); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &sr, nil
}

func (c *Client) Chat(model, prompt string) (*Response, error) {
	return c.ChatTagged("generate", model, prompt)
}

func (c *Client) ChatTagged(tag, model, prompt string) (*Response, error) {
	return c.post(tag, "/responses", responsesRequest{
		Model: model,
		Input: []InputMessage{{Role: "user", Content: prompt}},
	})
}
