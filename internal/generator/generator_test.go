package generator

import "testing"

func TestArticleOutputPath(t *testing.T) {
	tests := []struct {
		name  string
		index int
		total int
		want  string
	}{
		{name: "single article has no suffix", index: 1, total: 1, want: "out/202608280900.txt"},
		{name: "multiple articles get zero-padded suffix", index: 2, total: 3, want: "out/202608280900_02.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := articleOutputPath("out", "202608280900", tt.index, tt.total)
			if got != tt.want {
				t.Fatalf("articleOutputPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
