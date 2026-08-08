package vault

import (
	"bytes"
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrNoFrontmatter means the file has no leading `---` block at all. Fifteen notes in
// the real vault are in this state, so it is an expected condition, not a parse failure.
var ErrNoFrontmatter = errors.New("no YAML frontmatter")

var fence = []byte("---")

// SplitFrontmatter separates a note into its raw YAML block and the body that follows.
// The body is returned byte-identical, including its leading newline behaviour, because
// the migration is required never to reorder or rewrite body content.
func SplitFrontmatter(src []byte) (yamlSrc, body []byte, err error) {
	rest, ok := bytes.CutPrefix(normalizeEOL(src), append(fence, '\n'))
	if !ok {
		return nil, src, ErrNoFrontmatter
	}
	end := bytes.Index(rest, []byte("\n---"))
	if end < 0 {
		return nil, src, ErrNoFrontmatter
	}
	after := rest[end+len("\n---"):]
	if len(after) > 0 && after[0] != '\n' {
		return nil, src, ErrNoFrontmatter
	}
	return rest[:end], bytes.TrimPrefix(after, []byte("\n")), nil
}

func normalizeEOL(b []byte) []byte {
	if !bytes.Contains(b, []byte("\r\n")) {
		return b
	}
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}

// Frontmatter is a parsed YAML mapping that remembers the order its keys were written
// in. Order matters because `forge validate --fix` normalizes it, and knowing the
// original is what lets --fix report the change rather than silently reflowing a file.
type Frontmatter struct {
	Keys []string
	Vals map[string]*yaml.Node
}

// ParseFrontmatter parses a raw YAML block into an ordered mapping.
func ParseFrontmatter(yamlSrc []byte) (*Frontmatter, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(yamlSrc, &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return &Frontmatter{Vals: map[string]*yaml.Node{}}, nil
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, errors.New("frontmatter is not a YAML mapping")
	}
	return fromMapping(root), nil
}

func fromMapping(root *yaml.Node) *Frontmatter {
	fm := &Frontmatter{Vals: make(map[string]*yaml.Node, len(root.Content)/2)}
	for i := 0; i+1 < len(root.Content); i += 2 {
		k := root.Content[i].Value
		if _, dup := fm.Vals[k]; !dup {
			fm.Keys = append(fm.Keys, k)
		}
		fm.Vals[k] = root.Content[i+1]
	}
	return fm
}

// Has reports whether the key was present in the source, even if its value was null.
func (f *Frontmatter) Has(k string) bool { _, ok := f.Vals[k]; return ok }

// Str returns a scalar value as a string, or "" if absent or not a scalar.
func (f *Frontmatter) Str(k string) string {
	n, ok := f.Vals[k]
	if !ok || n.Kind != yaml.ScalarNode {
		return ""
	}
	return n.Value
}

// List returns a sequence value as strings. A scalar is treated as a one-element list,
// which is how several existing vault notes write `tags:`.
func (f *Frontmatter) List(k string) []string {
	n, ok := f.Vals[k]
	if !ok {
		return nil
	}
	if n.Kind == yaml.ScalarNode {
		return scalarAsList(n.Value)
	}
	out := make([]string, 0, len(n.Content))
	for _, c := range n.Content {
		out = append(out, c.Value)
	}
	return out
}

func scalarAsList(v string) []string {
	if v = strings.TrimSpace(v); v == "" || v == "null" {
		return nil
	}
	return []string{v}
}
