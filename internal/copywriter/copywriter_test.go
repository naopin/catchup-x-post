package copywriter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"catchup-x-post/internal/grok"
	"catchup-x-post/internal/newsclient"
)

func TestGenerateSingleCallWhenMeritsValid(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		// 1回目の生成でメリデメが揃っていれば追加呼び出し不要
		_, _ = w.Write([]byte(`{"id":"1","output":[{"type":"message","content":[{"type":"output_text","text":"{\"tweet\":\"テスト\\nhttps://example.com\",\"article\":\"【概要】\\nA\\n\\n【メリット・デメリット】\\nメリット:\\n・m1\\n・m2\\n\\nデメリット:\\n・d1\\n・d2\\n\\n【解説】\\nB\\n\\n【活用事例】\\nX\\nY\\nZ\\nW\"}"}]}]}`))
	}))
	t.Cleanup(srv.Close)

	c := grok.NewClientWithBaseURL("test", srv.URL)
	item := &newsclient.NewsItem{
		Title:          "t",
		Summary:        "s",
		Explanation:    "e",
		UseCases:       "u",
		Topic:          "tp",
		TrendingReason: "tr",
		URL:            "https://example.com",
	}

	out, err := Generate(c, "m", item)
	if err != nil {
		// 追加情報（デバッグ用）
		article := buildArticleFromItem(item)
		block, berr := generateMeritsDemeritsBlock(c, "m", item, article)
		merged := insertMeritsBlock(article, block)
		t.Fatalf("Generate: %v\nblock_err=%v\nblock=\n%s\n\nmerged_valid=%v\nmerged=\n%s", err, berr, block, hasValidMeritsDemerits(merged), merged)
	}
	if out.Tweet == "" || out.Article == "" {
		t.Fatalf("empty output: %+v", out)
	}
	if calls != 1 {
		t.Fatalf("calls=%d want=1", calls)
	}
	st := c.GetStats()
	if st.CallsTotal != 1 || st.CallsByTag["generate"] != 1 {
		t.Fatalf("stats=%+v", st)
	}
}

func TestGenerateTriggersMeritsRepair(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			// 初回: メリデメ無し
			_, _ = w.Write([]byte(`{"id":"1","output":[{"type":"message","content":[{"type":"output_text","text":"{\"tweet\":\"テスト\\nhttps://example.com\",\"article\":\"【概要】\\nA\\n\\n【解説】\\nB\\n\\n【活用事例】\\nX\\nY\\nZ\\nW\"}"}]}]}`))
			return
		}
		// 2回目: merits タグで補完
		_, _ = w.Write([]byte("{\"id\":\"2\",\"output\":[{\"type\":\"message\",\"content\":[{\"type\":\"output_text\",\"text\":\"【メリット・デメリット】\\nメリット:\\n- m1\\n- m2\\n\\nデメリット:\\n- d1\\n- d2\"}]}]}"))
	}))
	t.Cleanup(srv.Close)

	c := grok.NewClientWithBaseURL("test", srv.URL)
	item := &newsclient.NewsItem{
		Title:          "t",
		Summary:        "s",
		Explanation:    "e",
		UseCases:       "u",
		Topic:          "tp",
		TrendingReason: "tr",
		URL:            "https://example.com",
	}

	out, err := Generate(c, "m", item)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !hasValidMeritsDemerits(out.Article) {
		t.Fatalf("merits not repaired:\n%s", out.Article)
	}
	if calls != 2 {
		t.Fatalf("calls=%d want=2", calls)
	}
	st := c.GetStats()
	if st.CallsTotal != 2 || st.CallsByTag["generate"] != 1 || st.CallsByTag["merits"] != 1 {
		t.Fatalf("stats=%+v", st)
	}
}

