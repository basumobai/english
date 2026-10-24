package vocabulary

import (
	"regexp"
	"strings"
	"unicode"
)

var expansion = regexp.MustCompile(` +\([^)]*\)$`)

// Kind treats multiword expressions and separated phrase patterns as phrases.
// An abbreviation's expansion, such as AI (artificial intelligence), is not a phrase.
func Kind(word string) string {
	head := expansion.ReplaceAllString(word, "")
	if strings.ContainsFunc(head, unicode.IsSpace) || strings.Contains(head, "...") || strings.Contains(head, "…") {
		return "phrase"
	}
	return "word"
}
