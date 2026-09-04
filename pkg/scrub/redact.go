package scrub

import (
	"regexp"

	"github.com/mimir45/Knowledge-Forge/pkg/vault"
	"gopkg.in/yaml.v3"
)

// Patterns are applied in this order deliberately, and reLongToken excludes '-' and '_'
// on purpose: allowing them redacted slugs and dated filenames, corrupting a real vault.
var (
	reEmail       = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	reHomePath    = regexp.MustCompile(`(?:/Users|/home)/[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.\-]+)*`)
	reKeyPrefixed = regexp.MustCompile(`\b(?:sk-|ghp_|AKIA)[A-Za-z0-9_-]{10,}\b`)
	reLongToken   = regexp.MustCompile(`\b[A-Za-z0-9]{32,}\b`)
)

// Redact applies the vault's redaction patterns to one string and reports how many
// replacements it made.
func Redact(s string) (string, int) { return redact(s) }

// redact applies every pattern in turn and reports how many replacements were made.
func redact(s string) (string, int) {
	count := 0
	s = replaceCount(reEmail, s, "[REDACTED-EMAIL]", &count)
	s = replaceCount(reHomePath, s, "[REDACTED-PATH]", &count)
	s = replaceCount(reKeyPrefixed, s, "[REDACTED-KEY]", &count)
	s = redactLongTokens(s, &count)
	return s, count
}

func replaceCount(re *regexp.Regexp, s, repl string, count *int) string {
	return re.ReplaceAllStringFunc(s, func(string) string {
		*count++
		return repl
	})
}

// redactLongTokens filters reLongToken's matches to ones containing at least one digit.
func redactLongTokens(s string, count *int) string {
	return reLongToken.ReplaceAllStringFunc(s, func(m string) string {
		if !hasDigit(m) {
			return m
		}
		*count++
		return "[REDACTED-KEY]"
	})
}

func hasDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func redactBytes(b []byte) ([]byte, int) {
	s, n := redact(string(b))
	return []byte(s), n
}

// freeTextKeys are the only frontmatter fields scrubFields treats as free text.
var freeTextKeys = map[string]bool{"title": true, "sources": true}

// scrubFields redacts title and every sources[].url — the two free-text fields the
// schema imposes no fixed shape on — and leaves every other key untouched.
func scrubFields(fm *vault.Frontmatter) int {
	count := 0
	for _, k := range fm.Keys {
		if freeTextKeys[k] {
			count += scrubNode(fm.Vals[k])
		}
	}
	return count
}

// scrubNode redacts a scalar node's value in place, or recurses into a sequence/mapping.
func scrubNode(n *yaml.Node) int {
	if n == nil {
		return 0
	}
	if n.Kind == yaml.ScalarNode {
		redacted, count := redact(n.Value)
		n.Value = redacted
		return count
	}
	count := 0
	for _, c := range n.Content {
		count += scrubNode(c)
	}
	return count
}
