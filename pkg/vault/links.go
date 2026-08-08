package vault

import (
	"path"
	"regexp"
	"strings"
)

var (
	wikilinkRe = regexp.MustCompile(`\[\[([^\]\[|#]+)(?:#[^\]\[|]*)?(?:\|[^\]\[]*)?\]\]`)
	fenceRe    = regexp.MustCompile("(?s)```.*?```|`[^`\n]*`")
)

// Wikilinks returns every [[target]] in a body, in order, with any #heading and
// |alias stripped. Code fences and inline code are removed first: a `[[x]]` inside a
// fenced block is documentation of the syntax, not a link, and counting it corrupts
// every orphan and dangling-link metric downstream.
func Wikilinks(body []byte) []string {
	stripped := fenceRe.ReplaceAll(body, []byte(" "))
	ms := wikilinkRe.FindAllSubmatch(stripped, -1)
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		if t := strings.TrimSpace(string(m[1])); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// LinkKey reduces a wikilink target to the form the resolver matches on.
//
// Resolution is deliberately extension-agnostic: [[log]] and [[log.md]] name the same
// note. The fixture vault contains exactly one link of the second form — index.md's
// [[log.md]] — and a resolver that blindly appends ".md" reads it as dangling, dropping
// log.md's inbound count to zero. See docs/AUDIT.md section 11.
func LinkKey(target string) string {
	t := strings.TrimSpace(strings.TrimPrefix(target, "./"))
	t = strings.TrimSuffix(t, ".md")
	return strings.ToLower(strings.Trim(t, "/"))
}

// Index resolves wikilinks against a set of note paths. It matches a path-qualified
// link (`[[issues/foo]]`) on the full relative path and a bare link (`[[foo]]`) on the
// basename, which is how Obsidian itself behaves.
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
// Two notes named soft-delete.md and soft-deletion.md do not collide; two notes both
// named testcontainers-docker-socket.md in different directories do.
func (ix *Index) Ambiguous(target string) bool {
	return ix.ambig[LinkKey(path.Base(LinkKey(target)))]
}
