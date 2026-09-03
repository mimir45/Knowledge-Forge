package vault

import (
	"bytes"
	"path"
	"regexp"
	"strings"
)

var (
	wikilinkRe = regexp.MustCompile(`\[\[([^\]\[|#]+)(?:#[^\]\[|]*)?(?:\|[^\]\[]*)?\]\]`)
	fenceRe    = regexp.MustCompile("(?s)```.*?```|`[^`\n]*`")
)

// StripCode blanks every fenced block and inline code span in a body, replacing each
// with a single space.
func StripCode(body []byte) []byte { return fenceRe.ReplaceAll(body, []byte(" ")) }

// Wikilinks returns every [[target]] in a body, in order, with any #heading and |alias
// stripped.
func Wikilinks(body []byte) []string {
	stripped := StripCode(body)
	ms := wikilinkRe.FindAllSubmatch(stripped, -1)
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		if t := strings.TrimSpace(string(m[1])); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// FencedBlocks returns the content of every fenced code block in a body, in order,
// fence markers stripped.
func FencedBlocks(body []byte) [][]byte {
	ms := fenceRe.FindAll(body, -1)
	out := make([][]byte, 0, len(ms))
	for _, m := range ms {
		if !bytes.HasPrefix(m, []byte("```")) {
			continue // inline `code`, not a fenced block
		}
		out = append(out, bytes.Trim(stripFence(m), "\n"))
	}
	return out
}

// stripFence removes the opening ``` line (with its language tag) and the closing ```.
func stripFence(m []byte) []byte {
	s := bytes.TrimPrefix(m, []byte("```"))
	if i := bytes.IndexByte(s, '\n'); i >= 0 {
		s = s[i+1:]
	}
	return bytes.TrimSuffix(s, []byte("```"))
}

// FencedBlockLangs returns the info-string language tag for every block FencedBlocks
// returns, same order, "" when the fence has none.
func FencedBlockLangs(body []byte) []string {
	ms := fenceRe.FindAll(body, -1)
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		if !bytes.HasPrefix(m, []byte("```")) {
			continue
		}
		out = append(out, fenceLang(m))
	}
	return out
}

// fenceLang parses the same opening line stripFence discards, so the two can never
// disagree about where the info-string ends.
func fenceLang(m []byte) string {
	s := bytes.TrimPrefix(m, []byte("```"))
	if i := bytes.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(string(s[:i]))
	}
	return ""
}

// LinkKey reduces a wikilink target to the form the resolver matches on.
func LinkKey(target string) string {
	t := strings.TrimSpace(strings.TrimPrefix(target, "./"))
	t = strings.TrimSuffix(t, ".md")
	return strings.ToLower(strings.Trim(t, "/"))
}

// Index resolves wikilinks against a set of note paths.
type Index struct {
	byPath map[string]string // "issues/foo" -> "issues/foo.md"
	byName map[string]string // "foo"        -> "issues/foo.md"
	ambig  map[string]bool   // basenames claimed by more than one note
}

// NewIndex builds a resolver over vault-relative note paths (e.g. "issues/foo.md").
func NewIndex(relPaths []string) *Index {
	ix := &Index{
		byPath: make(map[string]string, len(relPaths)),
		byName: make(map[string]string, len(relPaths)),
		ambig:  map[string]bool{},
	}
	for _, p := range relPaths {
		ix.byPath[LinkKey(p)] = p
		name := LinkKey(path.Base(p))
		if _, dup := ix.byName[name]; dup {
			ix.ambig[name] = true
		}
		ix.byName[name] = p
	}
	return ix
}

// Resolve returns the note path a wikilink points at, and whether it resolved.
func (ix *Index) Resolve(target string) (string, bool) {
	k := LinkKey(target)
	if p, ok := ix.byPath[k]; ok {
		return p, true
	}
	p, ok := ix.byName[LinkKey(path.Base(k))]
	return p, ok
}

// Ambiguous reports whether a bare link to this basename has more than one candidate.
func (ix *Index) Ambiguous(target string) bool {
	return ix.ambig[LinkKey(path.Base(LinkKey(target)))]
}
