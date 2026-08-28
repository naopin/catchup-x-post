package webarticle

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleDoc = `トピック: Bunの新機能

【概要】
Bunに新しいAPIが追加された。

【解説】
詳細な解説文。

参考: https://x.com/example/status/1

【X投稿文案】（内容のみ・280字以内）
Bunの新機能について紹介します。 https://x.com/example/status/1`

func writeFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestListOrdersDescendingAndSkipsNonTxt(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "202607031817_01.txt", sampleDoc)
	writeFixture(t, dir, "202607031900_01.txt", sampleDoc)
	writeFixture(t, dir, ".gitkeep", "")

	got, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 articles, got %d", len(got))
	}
	if got[0].ID != "202607031900_01" {
		t.Fatalf("want newest first, got %q", got[0].ID)
	}
	if got[0].Title != "Bunの新機能" {
		t.Fatalf("unexpected title %q", got[0].Title)
	}
	if got[0].Date == "" {
		t.Fatal("want non-empty date")
	}
}

func TestListMissingDirReturnsEmpty(t *testing.T) {
	got, err := List(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want empty, got %d", len(got))
	}
}

func TestGet(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "202607031817_01.txt", sampleDoc)

	got, err := Get(dir, "202607031817_01")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Bunの新機能" {
		t.Fatalf("unexpected title %q", got.Title)
	}
	if got.SourceURL != "https://x.com/example/status/1" {
		t.Fatalf("unexpected source_url %q", got.SourceURL)
	}
	if got.Tweet == "" {
		t.Fatal("want non-empty tweet")
	}
	if got.Body == "" || got.Body == got.Tweet {
		t.Fatalf("unexpected body %q", got.Body)
	}
}

func TestGetNotFound(t *testing.T) {
	dir := t.TempDir()
	if _, err := Get(dir, "does-not-exist"); err != ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestGetRejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	if _, err := Get(dir, "../secret"); err != ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
