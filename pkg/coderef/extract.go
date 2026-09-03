package coderef

import (
	"regexp"
	"strconv"
	"strings"
)

// sourceExt is the set of extensions a citation may name.
var sourceExt = map[string]bool{
	".java": true, ".kt": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
}

var (
	// Inline code spans only. Fenced blocks are excluded on purpose — they are worked
	// examples, and a `class Foo` inside one is illustration, not citation.
	inlineSpan = regexp.MustCompile("`([^`\n]{1,200})`")
	// A trailing :NN or #LNN is the line number; both spellings appear in the vault.
	lineSuffix = regexp.MustCompile(`(?::|#L)(\d{1,6})$`)
	// A type or member name: CamelCase, optionally .member, optionally ().
	symbolPat = regexp.MustCompile(`^([A-Z][A-Za-z0-9_]{2,})(?:\.([A-Za-z_][A-Za-z0-9_]*))?(?:\(\))?$`)
)

// FromBody extracts every citation in a note body. It never guesses a repository —
// prose has no room to name one — so Repo stays empty and Resolve searches all of them.
func FromBody(noteRel string, body []byte) []Ref {
	var out []Ref
	seen := map[string]bool{}
	for _, m := range inlineSpan.FindAllStringSubmatch(string(body), -1) {
		r, ok := parseSpan(strings.TrimSpace(m[1]))
		if !ok || seen[r.Raw] {
			continue
		}
		seen[r.Raw] = true
		r.Note = noteRel
		out = append(out, r)
	}
	return out
}

// parseSpan classifies one inline code span. Most spans are neither — `mvn test`,
// `spring.datasource.url`, a shell flag — and returning false for them is the job.
func parseSpan(s string) (Ref, bool) {
	if s == "" || strings.ContainsAny(s, " \t\"'|$<>*?") {
		return Ref{}, false
	}
	raw := s
	line := 0
	if m := lineSuffix.FindStringSubmatch(s); m != nil {
		line, _ = strconv.Atoi(m[1])
		s = s[:len(s)-len(m[0])]
	}
	if ext := extOf(s); sourceExt[ext] {
		return Ref{Raw: raw, Kind: KindPath, Path: filepathSlash(s), Line: line}, true
	}
	return parseSymbolSpan(raw, s)
}

// parseSymbolSpan handles the no-extension case: a bare type or member name. Anything
// carrying a slash without a source extension is a path to something that is not code.
func parseSymbolSpan(raw, s string) (Ref, bool) {
	if strings.Contains(s, "/") {
		return Ref{}, false
	}
	m := symbolPat.FindStringSubmatch(s)
	if m == nil {
		return Ref{}, false
	}
	sym := m[1]
	if m[2] != "" {
		sym += "." + m[2]
	}
	return Ref{Raw: raw, Kind: KindSymbol, Symbol: sym}, true
}

func extOf(s string) string {
	if i := strings.LastIndex(s, "."); i >= 0 {
		return strings.ToLower(s[i:])
	}
	return ""
}

// FromFrontmatter reads the canonical `code_refs:` block.
func FromFrontmatter(noteRel string, entries []string) []Ref {
	var out []Ref
	for _, e := range entries {
		if r, ok := parseCanonical(strings.TrimSpace(e)); ok {
			r.Note = noteRel
			out = append(out, r)
		}
	}
	return out
}

func parseCanonical(e string) (Ref, bool) {
	r := Ref{Raw: e, Kind: KindPath}
	if i := strings.Index(e, "#"); i >= 0 {
		r.Symbol, e = e[i+1:], e[:i]
	}
	repo, rest, ok := strings.Cut(e, ":")
	if !ok || repo == "" || rest == "" {
		return Ref{}, false
	}
	r.Repo = repo
	if m := lineSuffix.FindStringSubmatch(rest); m != nil {
		r.Line, _ = strconv.Atoi(m[1])
		rest = rest[:len(rest)-len(m[0])]
	}
	r.Path = filepathSlash(rest)
	return r, true
}
