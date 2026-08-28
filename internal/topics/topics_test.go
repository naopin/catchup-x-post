package topics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLineAndDuplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, historyFile)
	content := "選んだトピック: Grok V9-Medium (リリース) https://x.com/a/status/1\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !store.IsDuplicate("Grok V9-Mediumトレーニング", "Grok V9-Medium") {
		t.Fatal("expected duplicate topic")
	}
	if store.IsDuplicate("Claude Code workflows", "Claude") {
		t.Fatal("expected different topic")
	}
}

func TestFormatLine(t *testing.T) {
	got := FormatLine("リリース発表", "Grok V9-Medium", "https://x.com/foo/status/1")
	want := "選んだトピック: Grok V9-Medium (リリース発表) https://x.com/foo/status/1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
