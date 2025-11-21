package util

import (
	"strings"
	"unicode"
)

// NormalizeName lowers, trims, removes obvious punctuation, and collapses spaces.
func NormalizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	// Replace common separators with space.
	replacer := strings.NewReplacer("_", " ", "-", " ", ".", " ", ",", " ", "/", " ", "\\", " ")
	s = replacer.Replace(s)
	// Remove remaining punctuation, keep letters/numbers/spaces.
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	parts := strings.Fields(b.String())
	return strings.Join(parts, " ")
}
