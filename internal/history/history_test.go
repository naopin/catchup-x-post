package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreAddAndSimilarURL(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Add("https://x.com/foo/status/1", "Bun高速化", "Bun"); err != nil {
		t.Fatal(err)
	}
	if !store.IsSimilar("https://x.com/foo/status/1", "別タイトル", "Other") {
		t.Fatal("expected same URL to match")
	}

	store2, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !store2.IsSimilar("https://x.com/foo/status/1", "", "") {
		t.Fatal("expected persistence")
	}

	path := filepath.Join(dir, entriesFile)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("entries file missing: %v", err)
	}
}

func TestSimilarTitleOverlap(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	_ = store.Add("https://x.com/a/1", "Bun次バージョンでのCJK文字幅計算の大幅高速化", "Bun")

	if !store.IsSimilar("https://x.com/b/2", "Bunの次バージョンでCJK文字幅計算が大幅高速化", "Bun") {
		t.Fatal("expected similar titles to match")
	}
	if store.IsSimilar("https://x.com/c/3", "Claude Code workflows 発表", "Claude") {
		t.Fatal("expected unrelated title to not match")
	}
}
