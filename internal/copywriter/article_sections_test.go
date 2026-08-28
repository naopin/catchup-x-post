package copywriter

import (
	"strings"
	"testing"
)

func TestHasValidMeritsDemerits(t *testing.T) {
	valid := `【概要】
x

【メリット・デメリット】
メリット:
・高速
・無料

デメリット:
・移行コスト
・検証工数

【解説】
y`
	if !hasValidMeritsDemerits(valid) {
		t.Fatal("expected valid")
	}

	placeholder := `【メリット・デメリット】
メリット:
・（文案生成時に追記）

デメリット:
・（文案生成時に追記）`
	if hasValidMeritsDemerits(placeholder) {
		t.Fatal("expected invalid placeholder")
	}
}

func TestInsertMeritsBlock_replacesPlaceholderSection(t *testing.T) {
	article := `【概要】
概要文

【メリット・デメリット】
メリット:
・（文案生成時に追記）

デメリット:
・（文案生成時に追記）

【解説】
解説文`

	block := `【メリット・デメリット】
メリット:
・A
・B

デメリット:
・C
・D`

	got := insertMeritsBlock(article, block)
	if !hasValidMeritsDemerits(got) {
		t.Fatalf("expected valid merits after insert:\n%s", got)
	}
	if strings.Contains(got, meritsPlaceholder) {
		t.Fatal("placeholder should be removed")
	}
}
