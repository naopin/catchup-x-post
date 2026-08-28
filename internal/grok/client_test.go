package grok

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientCountsCallsAndTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","output":[{"type":"message","content":[{"type":"output_text","text":"ok"}]}]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClientWithBaseURL("test", srv.URL)
	if _, err := c.ChatTagged("a", "m", "p1"); err != nil {
		t.Fatalf("ChatTagged a: %v", err)
	}
	if _, err := c.ChatTagged("b", "m", "p2"); err != nil {
		t.Fatalf("ChatTagged b: %v", err)
	}

	st := c.GetStats()
	if st.CallsTotal != 2 {
		t.Fatalf("CallsTotal=%d", st.CallsTotal)
	}
	if st.CallsByTag["a"] != 1 || st.CallsByTag["b"] != 1 {
		t.Fatalf("CallsByTag=%v", st.CallsByTag)
	}
}

func TestClientMaxCalls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","output":[{"type":"message","content":[{"type":"output_text","text":"ok"}]}]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClientWithBaseURL("test", srv.URL)
	c.SetMaxCalls(1)

	if _, err := c.ChatTagged("a", "m", "p1"); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if _, err := c.ChatTagged("a", "m", "p2"); err == nil {
		t.Fatalf("expected max calls error")
	}
}

