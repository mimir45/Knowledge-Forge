package scrub

import (
	"regexp"

	"gopkg.in/yaml.v3"
	"knowledge-forge/pkg/vault"
)

// Patterns are applied in this order deliberately: email before the generic long-token
// heuristic (so a local-part never gets a second, uglier redaction), key-prefixed
// tokens before the generic heuristic (so "sk-..." reads as [REDACTED-KEY] once, not
// twice), and absolute home paths on their own pass since they use a different alphabet
// (slashes) than the token heuristics.
//
// reLongToken deliberately excludes '-' and '_' from its character class, even though
// real secrets sometimes contain them (base64url, some API keys). In this vault those
// two characters are word separators, not token characters: slugs
// (`^[a-z0-9]+(-[a-z0-9]+)*$`) and dated filenames are long, hyphenated, and completely
// unremarkable — e.g. "2026-04-13-local-ai-continue-rag-spring" is 40 characters and
// would false-positive under a class that allows '-'. A real vault dry run caught this:
// it was corrupting `sources: url:` and wikilink citations throughout the vault. Secrets
// shaped like sk-/ghp_/AKIA-prefixed keys are already caught by reKeyPrefixed, which
// keeps '-'/'_' in its class deliberately since those prefixes require it.
var (
	reEmail       = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)
	reHomePath    = regexp.MustCompile(`(?:/Users|/home)/[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.\-]+)*`)
	reKeyPrefixed = regexp.MustCompile(`\b(?:sk-|ghp_|AKIA)[A-Za-z0-9_-]{10,}\b`)
	reLongToken   = regexp.MustCompile(`\b[A-Za-z0-9]{32,}\b`)
)

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

// redactLongTokens filters reLongToken's matches to ones containing at least one
// digit. RE2 has no lookahead, so the filter lives here rather than in the pattern.
// A pure-letter 32+ char run is overwhelmingly a camelCase identifier (howto notes'
// code samples are full of them, e.g. getPaymentOutboxMessageBySagaIdAndSagaStatus)
// — found scrubbing the real vault after the slug/filename fix above. A random draw
// of 32+ characters from [A-Za-z0-9] is a digit with near-certainty, so hex hashes,
// base64 blobs and JWT segments still get caught. Known residual false positive,
// left as-is rather than chased further: an identifier that happens to embed a digit
// (e.g. TestE2ESessionContextRespectsTheBudget, "E2E") still trips this — real vault
// scrub found exactly one such case in 122 notes, an acceptable trade-off for the
// alternative (real digit-bearing secrets going unredacted).
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

// freeTextKeys are the only frontmatter fields scrubFields treats as free text. Every
// other schema field is pattern-, enum-, or date-constrained (slug, type, stack, tags,
// created/updated/verified, forge_version, origin, ...) — a value there was already
// validated against a fixed shape, so redacting it blind risks turning a legitimately
// long, hyphenated value (a slug in particular: `^[a-z0-9]+(-[a-z0-9]+)*$`, no length
// cap under 80 chars) into a schema failure instead of catching an actual secret.
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
