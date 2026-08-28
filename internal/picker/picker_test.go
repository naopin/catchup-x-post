package picker

import (
	"testing"

	"catchup-x-post/internal/history"
	"catchup-x-post/internal/newsclient"
)

func TestPickSkipsSimilar(t *testing.T) {
	store, err := history.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_ = store.Add("https://x.com/a/status/1", "Bun高速化", "Bun")

	items := []newsclient.NewsItem{
		{Title: "Bun高速化の続報", Summary: "sa", URL: "https://x.com/a/status/1", Topic: "Bun"},
		{Title: "Claude新機能", Summary: "sb", URL: "https://x.com/b/status/2", Topic: "Claude"},
	}

	got, err := Pick(items, store)
	if err != nil {
		t.Fatal(err)
	}
	if got.URL != "https://x.com/b/status/2" {
		t.Fatalf("got %q", got.URL)
	}
}
