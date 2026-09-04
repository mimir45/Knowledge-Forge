package vault

import (
	"strings"
	"unicode"
)

// Slug turns a title into a canonical kebab-case identifier.
func Slug(title string) string {
	folded := foldRunes(title)
	parts := strings.FieldsFunc(folded, func(r rune) bool { return r == '-' })
	s := strings.Join(parts, "-")
	if len(s) > maxSlugLen {
		s = truncateAtBoundary(s, maxSlugLen)
	}
	return s
}

const maxSlugLen = 80

// foldRunes lowercases, transliterates the Latin-1/Turkish letters that actually occur
// in this vault, and reduces everything else to a single separator rune.
func foldRunes(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case translit[r] != "":
			b.WriteString(translit[r])
		default:
			b.WriteByte('-')
		}
	}
	return b.String()
}

// translit is intentionally explicit rather than a Unicode normalization pass.
var translit = map[rune]string{
	'ç': "c", 'ğ': "g", 'ı': "i", 'ö': "o", 'ş': "s", 'ü': "u",
	'á': "a", 'à': "a", 'â': "a", 'ä': "a", 'å': "a", 'ã': "a",
	'é': "e", 'è': "e", 'ê': "e", 'ë': "e",
	'í': "i", 'ì': "i", 'î': "i", 'ï': "i",
	'ó': "o", 'ò': "o", 'ô': "o", 'õ': "o",
	'ú': "u", 'ù': "u", 'û': "u",
	'ñ': "n", 'ß': "ss", 'æ': "ae", 'ø': "o",
	'+': "plus", '&': "and", '#': "sharp",
}

// truncateAtBoundary cuts to at most n bytes without leaving a trailing partial word.
func truncateAtBoundary(s string, n int) string {
	s = s[:n]
	if i := strings.LastIndexByte(s, '-'); i > 0 {
		s = s[:i]
	}
	return strings.Trim(s, "-")
}

// SlugUnique returns Slug(title), suffixed with -2, -3 … until it does not collide with
// taken. The suffix search is ordered, so a given (title, taken) pair is reproducible.
func SlugUnique(title string, taken map[string]bool) string {
	base := Slug(title)
	if base == "" {
		base = "untitled"
	}
	if !taken[base] {
		return base
	}
	for n := 2; ; n++ {
		cand := base + "-" + itoa(n)
		if !taken[cand] {
			return cand
		}
	}
}

func itoa(n int) string {
	var d []rune
	for ; n > 0; n /= 10 {
		d = append([]rune{rune('0' + n%10)}, d...)
	}
	return string(d)
}

// IsSlug reports whether s already satisfies the schema's slug pattern.
func IsSlug(s string) bool {
	if s == "" || s[0] == '-' || s[len(s)-1] == '-' || strings.Contains(s, "--") {
		return false
	}
	return !strings.ContainsFunc(s, func(r rune) bool {
		return !(unicode.IsLower(r) && r < 128 || unicode.IsDigit(r) && r < 128 || r == '-')
	})
}
