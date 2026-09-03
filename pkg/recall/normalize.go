// Package recall is the retrieval-before-research engine: it decides whether the vault
// already answers a question before any research runs.
package recall

import (
	"sort"
	"strings"
	"unicode"
)

// stopwords are the scaffolding of the phrasings that trigger the skill.
var stopwords = map[string]bool{}

func init() {
	const list = `a an and are as at be best between but by can do does doing done
	explain for from get gets getting handle handled handles has have how in into is it
	its me my of on or practices should so than that the their them then there these
	they this to use used uses using vs was way we what when where which while who why
	will with work working works would you your`
	for _, w := range strings.Fields(list) {
		stopwords[w] = true
	}
}

// Tokens splits text into lowercase alphanumeric tokens, dropping single characters.
func Tokens(text string) []string {
	split := func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }
	var out []string
	for _, f := range strings.FieldsFunc(strings.ToLower(text), split) {
		if len(f) > 1 {
			out = append(out, f)
		}
	}
	return out
}

// Terms is Tokens minus stopwords, deduplicated and sorted so the set is deterministic.
// Sorted output is what makes --explain and matched_on byte-stable across runs.
func Terms(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range Tokens(text) {
		if !stopwords[t] && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	sort.Strings(out)
	return out
}

// setOf builds a lookup from an arbitrary token slice, folding each value through Tokens
// so that a tag like "spring-boot" is matched by the query terms "spring" and "boot".
func setOf(values []string) map[string]bool {
	s := map[string]bool{}
	for _, v := range values {
		for _, t := range Tokens(v) {
			s[t] = true
		}
	}
	return s
}

// termSetOf is setOf with stopwords removed — the note-side counterpart of Terms, for
// the channels that compare against prose rather than against a controlled vocabulary.
func termSetOf(values ...string) map[string]bool {
	s := setOf(values)
	for w := range s {
		if stopwords[w] {
			delete(s, w)
		}
	}
	return s
}

// intersect returns the members of terms present in set, order preserved.
func intersect(terms []string, set map[string]bool) []string {
	var out []string
	for _, t := range terms {
		if set[t] {
			out = append(out, t)
		}
	}
	return out
}
