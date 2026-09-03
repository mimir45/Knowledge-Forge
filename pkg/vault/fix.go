package vault

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ForgeVersion is stamped into notes that lack the key.
const ForgeVersion = "2.0.0"

// Fix rewrites a note's frontmatter into contract shape and returns the full file
// content. It only ever touches frontmatter: Body is copied through byte-for-byte, which
// is what makes --fix and the migration safe to run on notes nobody has reviewed.
//
// Fix does not invent meaning. It will not guess `type`, `stack`, `tags` or `title`
// content — those stay missing and stay reported by Validate.
func Fix(n *Note, s *Schema) ([]byte, []string, error) {
	if n.FMErr != nil && n.FMErr != ErrNoFrontmatter {
		return nil, nil, n.FMErr
	}
	fm := n.FM
	if fm == nil {
		fm = &Frontmatter{Vals: map[string]*yaml.Node{}}
	}
	changes := applyAll(n, s, fm)
	out, err := render(fm, s, n.Body)
	return out, changes, err
}

// RenderNote emits frontmatter in schema key order followed by body, byte-for-byte the
// same as Fix's own render step. Exported for pkg/vault/quarantine.go, which builds a
// fresh Frontmatter for an _inbox/ draft rather than fixing an existing note.
func RenderNote(fm *Frontmatter, s *Schema, body []byte) ([]byte, error) {
	return render(fm, s, body)
}

func applyAll(n *Note, s *Schema, fm *Frontmatter) []string {
	var changes []string
	changes = append(changes, normalizeLists(s, fm)...)
	changes = append(changes, backfillDates(n, fm)...)
	changes = append(changes, carryLegacySource(fm)...)
	changes = append(changes, backfillDefaults(n, s, fm)...)
	if wasOutOfOrder(fm, s) {
		changes = append(changes, "key order normalized")
	}
	return changes
}

// normalizeLists lowercases and de-aliases every item in stack and tags.
func normalizeLists(s *Schema, fm *Frontmatter) []string {
	var changes []string
	for _, key := range []string{"stack", "tags"} {
		if !fm.Has(key) {
			continue
		}
		fixed, note := canonicalizeItems(s, key, fm.List(key))
		fm.Vals[key] = seqNode(fixed)
		changes = append(changes, note...)
	}
	return changes
}

func canonicalizeItems(s *Schema, key string, items []string) ([]string, []string) {
	out := make([]string, 0, len(items))
	var changes []string
	seen := map[string]bool{}
	for _, it := range items {
		c, _ := s.Canonical(key, strings.ToLower(strings.TrimSpace(it)))
		if c != it {
			changes = append(changes, fmt.Sprintf("%s: %q -> %q", key, it, c))
		}
		if !seen[c] && c != "" {
			seen[c] = true
			out = append(out, c)
		}
	}
	return out, changes
}

// backfillDates fills absent or malformed dates from the file's mtime. mtime is the only
// honest source available: the vault's git history is a single import commit.
func backfillDates(n *Note, fm *Frontmatter) []string {
	stamp := time.Unix(n.ModTime, 0).UTC().Format("2006-01-02")
	if legacy := strings.Trim(fm.Str("date"), `"' `); isISODate(legacy) {
		stamp = legacy
	}
	var changes []string
	for _, key := range []string{"created", "updated", "verified"} {
		if isISODate(fm.Str(key)) {
			continue
		}
		setScalar(fm, key, stamp)
		changes = append(changes, fmt.Sprintf("%s: <- %s", key, stamp))
	}
	return changes
}

// carryLegacySource converts the v1 `source:` key into one `sources` entry before render
// drops it. 63 of the real vault's 93 notes carry it, and it is the only provenance those
// notes have: dropping it would silently destroy their citation. The value is copied verbatim
// — schema.yaml's `url` accepts both an http(s) URL and a vault-relative path precisely
// because this key holds both.
func carryLegacySource(fm *Frontmatter) []string {
	src := strings.Trim(fm.Str("source"), `"' `)
	if src == "" || len(fm.List("sources")) > 0 {
		return nil
	}
	accessed := fm.Str("created")
	if !isISODate(accessed) {
		return nil // backfillDates always sets created; a non-date here means bad input
	}
	if !fm.Has("sources") {
		fm.Keys = append(fm.Keys, "sources")
	}
	fm.Vals["sources"] = sourceSeq(src, accessed)
	return []string{fmt.Sprintf("sources: <- source: %q (kind: session)", src)}
}

// sourceSeq builds the single-entry `sources` list. kind is always `session`: the v1
// til-writer wrote `source:` to point at the Claude Code transcript a note came from.
func sourceSeq(url, accessed string) *yaml.Node {
	item := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, kv := range [][2]string{{"url", url}, {"accessed", accessed}, {"kind", "session"}} {
		item.Content = append(item.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: kv[0]},
			&yaml.Node{Kind: yaml.ScalarNode, Value: kv[1]})
	}
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{item}}
}

// backfillDefaults adds the keys whose correct value is a schema constant.
func backfillDefaults(n *Note, s *Schema, fm *Frontmatter) []string {
	var changes []string
	for _, key := range []string{"related", "supersedes", "sources"} {
		if !fm.Has(key) {
			fm.Vals[key] = seqNode(nil)
			fm.Keys = append(fm.Keys, key)
			changes = append(changes, key+": <- []")
		}
	}
	changes = append(changes, defaultScalar(fm, "forge_version", ForgeVersion)...)
	changes = append(changes, defaultScalar(fm, "confidence", "low")...)
	changes = append(changes, defaultScalar(fm, "depth", "3")...)
	if d := s.FreshnessDefault(fm.Str("type")); d > 0 && !fm.Has("freshness_days") {
		changes = append(changes, defaultScalar(fm, "freshness_days", itoa(d))...)
	}
	// Title falls back to the first H1 and then the filename, so it is derived, never
	// invented. Slug is derived from whatever title survives that chain.
	if t := n.Title(); t != "" {
		changes = append(changes, defaultScalar(fm, "title", t)...)
		changes = append(changes, defaultScalar(fm, "slug", Slug(t))...)
	}
	return changes
}

func defaultScalar(fm *Frontmatter, key, val string) []string {
	if fm.Str(key) != "" {
		return nil
	}
	setScalar(fm, key, val)
	return []string{fmt.Sprintf("%s: <- %s", key, val)}
}

// setScalar writes a value with an empty tag so the encoder infers the YAML type from
// the text. Forcing !!str would emit dates and integers as quoted strings, which is
// valid YAML but reads as a machine artefact in a file a human is meant to edit.
func setScalar(fm *Frontmatter, key, val string) {
	if !fm.Has(key) {
		fm.Keys = append(fm.Keys, key)
	}
	fm.Vals[key] = &yaml.Node{Kind: yaml.ScalarNode, Value: val}
}

func seqNode(items []string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
	for _, it := range items {
		n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: it})
	}
	return n
}

func wasOutOfOrder(fm *Frontmatter, s *Schema) bool {
	rank, prev := map[string]int{}, -1
	for i, k := range s.KeyOrder {
		rank[k] = i
	}
	for _, k := range fm.Keys {
		r, ok := rank[k]
		if !ok {
			return true // an unknown key is about to be dropped from the emitted order
		}
		if r < prev {
			return true
		}
		prev = r
	}
	return false
}

// render emits frontmatter in schema key order, followed by the untouched body.
// Keys the schema does not define are dropped — `status` and `date` are the two the
// migration retires, and taxonomy.md §5 records why.
func render(fm *Frontmatter, s *Schema, body []byte) ([]byte, error) {
	m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, key := range s.KeyOrder {
		v, ok := fm.Vals[key]
		if !ok {
			continue
		}
		m.Content = append(m.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, v)
	}
	return marshalDoc(m, body)
}

func marshalDoc(m *yaml.Node, body []byte) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	out := append([]byte("---\n"), buf.Bytes()...)
	return append(append(out, []byte("---\n")...), body...), nil
}
