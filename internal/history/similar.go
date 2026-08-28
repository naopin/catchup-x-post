package history

import (
	"strings"
	"unicode"
)

func similarEntry(a, b Entry) bool {
	if a.URL != "" && a.URL == b.URL {
		return true
	}

	at := normalizeKey(a.Title)
	bt := normalizeKey(b.Title)
	if at != "" && bt != "" {
		if at == bt {
			return true
		}
		if strings.Contains(at, bt) || strings.Contains(bt, at) {
			return true
		}
		if tokenOverlapRatio(at, bt) >= 0.55 {
			return true
		}
	}

	ap := normalizeKey(a.Topic)
	bp := normalizeKey(b.Topic)
	if ap != "" && bp != "" && ap == bp {
		return true
	}

	return false
}

func normalizeKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func tokenOverlapRatio(a, b string) float64 {
	tokensA := tokenSet(a)
	tokensB := tokenSet(b)
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0
	}
	inter := 0
	for t := range tokensA {
		if tokensB[t] {
			inter++
		}
	}
	denom := len(tokensA)
	if len(tokensB) < denom {
		denom = len(tokensB)
	}
	return float64(inter) / float64(denom)
}

func tokenSet(s string) map[string]bool {
	set := make(map[string]bool)
	for _, part := range strings.Fields(s) {
		if len([]rune(part)) >= 2 {
			set[part] = true
		}
	}
	if len(set) == 0 && len(s) >= 4 {
		// CJK などスペースなしタイトル用に4文字窓
		runes := []rune(s)
		for i := 0; i+4 <= len(runes); i++ {
			set[string(runes[i:i+4])] = true
		}
	}
	return set
}
