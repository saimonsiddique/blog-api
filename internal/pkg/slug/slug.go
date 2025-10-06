package slug

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
	multiDashRegex       = regexp.MustCompile(`-+`)
)

// Generate creates a URL-friendly slug from a string
func Generate(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove accents and normalize unicode
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMark), norm.NFC)
	s, _, _ = transform.String(t, s)

	// Replace non-alphanumeric characters with dashes
	s = nonAlphanumericRegex.ReplaceAllString(s, "-")

	// Replace multiple dashes with single dash
	s = multiDashRegex.ReplaceAllString(s, "-")

	// Trim dashes from start and end
	s = strings.Trim(s, "-")

	return s
}

func isMark(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
