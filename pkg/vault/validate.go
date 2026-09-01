package vault

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// Issue is one validation failure. Code is stable and machine-greppable; Msg is the
// actionable half. Fixable marks the ones `forge validate --fix` can repair mechanically
// — never a judgement call, only date backfill, key order, case, and alias rewriting.
type Issue struct {
	Rel     string
	Key     string
	Code    string
	Msg     string
	Fixable bool
}

func (i Issue) String() string {
	loc := i.Rel
	if i.Key != "" {
		loc += ": " + i.Key
	}
	fix := ""
	if i.Fixable {
		fix = "  [--fix]"
	}
	return fmt.Sprintf("%s: %s: %s%s", loc, i.Code, i.Msg, fix)
}

// Validate checks one note against the schema and returns every problem found. It never
// stops at the first failure: a caller wants the whole list, not a bisect.
func Validate(n *Note, s *Schema) []Issue {
	if n.FMErr != nil {
		return []Issue{frontmatterIssue(n)}
	}
	var out []Issue
	out = append(out, checkPresence(n, s)...)
	for _, key := range s.KeyOrder {
		if n.FM.Has(key) {
			out = append(out, checkField(n, s, key)...)
		}
	}
	out = append(out, checkOrder(n, s)...)
	out = append(out, checkDates(n)...)
	out = append(out, checkSourcesArity(n, s)...)
	out = append(out, checkEngineTrail(n)...)
	return out
}

func frontmatterIssue(n *Note) Issue {
	if n.FMErr == ErrNoFrontmatter {
		return Issue{n.Rel, "", "no-frontmatter",
			"file has no YAML frontmatter block", true}
	}
	return Issue{n.Rel, "", "bad-yaml", n.FMErr.Error(), false}
}

// checkPresence reports required keys that are absent and keys the schema does not know.
func checkPresence(n *Note, s *Schema) []Issue {
	var out []Issue
	for _, key := range s.KeyOrder {
		if f := s.Fields[key]; f.Required && !n.FM.Has(key) {
			out = append(out, Issue{n.Rel, key, "missing",
				"required key is absent", backfillable(n, s, key)})
		}
	}
	for _, key := range n.FM.Keys {
		if _, known := s.Fields[key]; !known {
			out = append(out, Issue{n.Rel, key, "unknown-key",
				"not in the note contract; see references/schema.yaml", false})
		}
	}
	return out
}

// backfillable is isBackfillable plus the one constraint that is not a property of the
// key alone: freshness_days has no single default, only a per-type one, so on a note
// whose type is missing or unrecognised there is nothing mechanical to write. Marking it
// [--fix] there would promise a repair that `forge validate --fix` cannot deliver.
func backfillable(n *Note, s *Schema, key string) bool {
	if key == "freshness_days" {
		return s.FreshnessDefault(n.FM.Str("type")) > 0
	}
	return isBackfillable(key)
}

// isBackfillable marks the keys --fix can invent a correct value for without guessing
// at meaning. Dates come from file mtime; the rest are schema constants.
func isBackfillable(key string) bool {
	switch key {
	case "created", "updated", "verified", "freshness_days", "forge_version",
		"related", "supersedes", "sources", "confidence", "depth", "slug", "title":
		return true
	}
	return false
}

// checkOrder reports frontmatter written in an order other than schema key_order.
// Order is never a hard failure on its own, but it is always fixable.
func checkOrder(n *Note, s *Schema) []Issue {
	rank := make(map[string]int, len(s.KeyOrder))
	for i, k := range s.KeyOrder {
		rank[k] = i
	}
	prev, prevKey := -1, ""
	for _, k := range n.FM.Keys {
		r, known := rank[k]
		if !known {
			continue
		}
		if r < prev {
			return []Issue{{n.Rel, k, "key-order",
				fmt.Sprintf("%q must come before %q", k, prevKey), true}}
		}
		prev, prevKey = r, k
	}
	return nil
}

func checkField(n *Note, s *Schema, key string) []Issue {
	f := s.Fields[key]
	if strings.HasPrefix(f.Type, "list<") {
		return checkList(n, s, key, f)
	}
	return checkScalar(n, s, key, f, n.FM.Str(key))
}

func checkScalar(n *Note, s *Schema, key string, f *Field, v string) []Issue {
	var out []Issue
	if v == "" {
		return []Issue{{n.Rel, key, "empty", "key present but has no value", backfillable(n, s, key)}}
	}
	if len(f.allowed) > 0 && !f.allowed[v] {
		out = append(out, Issue{n.Rel, key, "not-in-enum",
			fmt.Sprintf("%q is not one of %s", v, strings.Join(f.Values, "|")), false})
	}
	out = append(out, checkPattern(n, s, key, f, v)...)
	out = append(out, checkNumeric(n, key, f, v)...)
	out = append(out, checkLength(n, key, f, v)...)
	return out
}

func checkPattern(n *Note, s *Schema, key string, f *Field, v string) []Issue {
	if f.Type == "date" && !isISODate(v) {
		return []Issue{{n.Rel, key, "bad-date",
			fmt.Sprintf("%q is not YYYY-MM-DD", v), backfillable(n, s, key)}}
	}
	if f.rePattern != nil && !f.rePattern.MatchString(v) {
		return []Issue{{n.Rel, key, "bad-format",
			fmt.Sprintf("%q does not match %s", v, f.Pattern), key == "slug"}}
	}
	return nil
}

func checkNumeric(n *Note, key string, f *Field, v string) []Issue {
	if f.Type != "int" {
		return nil
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return []Issue{{n.Rel, key, "not-an-int", fmt.Sprintf("%q is not an integer", v), false}}
	}
	if f.Min != nil && i < *f.Min || f.Max != nil && i > *f.Max {
		return []Issue{{n.Rel, key, "out-of-range",
			fmt.Sprintf("%d outside [%d,%d]", i, deref(f.Min), deref(f.Max)), false}}
	}
	return nil
}

func checkLength(n *Note, key string, f *Field, v string) []Issue {
	if f.MinLength != nil && len(v) < *f.MinLength {
		return []Issue{{n.Rel, key, "too-short",
			fmt.Sprintf("%d chars, minimum %d", len(v), *f.MinLength), false}}
	}
	if f.MaxLength != nil && len(v) > *f.MaxLength {
		return []Issue{{n.Rel, key, "too-long",
			fmt.Sprintf("%d chars, maximum %d", len(v), *f.MaxLength), false}}
	}
	return nil
}

func deref(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func checkList(n *Note, s *Schema, key string, f *Field) []Issue {
	if f.Type == "list<object>" {
		return checkObjectList(n, s, key, f)
	}
	items := n.FM.List(key)
	out := checkArity(n, key, f, len(items))
	for _, it := range items {
		out = append(out, checkItem(n, s, key, f, it)...)
	}
	return out
}

func checkArity(n *Note, key string, f *Field, count int) []Issue {
	if f.MinItems != nil && count < *f.MinItems {
		return []Issue{{n.Rel, key, "too-few",
			fmt.Sprintf("%d items, minimum %d", count, *f.MinItems), isBackfillable(key)}}
	}
	if f.MaxItems != nil && count > *f.MaxItems {
		return []Issue{{n.Rel, key, "too-many",
			fmt.Sprintf("%d items, maximum %d", count, *f.MaxItems), false}}
	}
	return nil
}

func checkItem(n *Note, s *Schema, key string, f *Field, it string) []Issue {
	if canon, isAlias := s.Canonical(key, it); isAlias {
		return []Issue{{n.Rel, key, "alias",
			fmt.Sprintf("%q is an alias for %q", it, canon), true}}
	}
	if it != strings.ToLower(it) {
		return []Issue{{n.Rel, key, "not-lowercase", fmt.Sprintf("%q must be lowercase", it), true}}
	}
	if f.Vocabulary == "closed" && !f.allowed[it] {
		return []Issue{{n.Rel, key, "not-in-vocabulary",
			fmt.Sprintf("%q is not in the closed %s vocabulary", it, key), false}}
	}
	if f.reItem != nil && !f.reItem.MatchString(it) {
		return []Issue{{n.Rel, key, "bad-item",
			fmt.Sprintf("%q does not match %s", it, f.ItemPattern), false}}
	}
	return nil
}

func checkObjectList(n *Note, s *Schema, key string, f *Field) []Issue {
	node, ok := n.FM.Vals[key]
	if !ok || node.Kind != yaml.SequenceNode {
		return checkArity(n, key, f, 0)
	}
	out := checkArity(n, key, f, len(node.Content))
	for i, item := range node.Content {
		out = append(out, checkObjectItem(n, s, key, f, i, item)...)
	}
	return out
}

func checkObjectItem(n *Note, s *Schema, key string, f *Field, i int, item *yaml.Node) []Issue {
	if item.Kind != yaml.MappingNode {
		return []Issue{{n.Rel, key, "bad-item",
			fmt.Sprintf("[%d] is not a mapping", i), false}}
	}
	sub := fromMapping(item)
	var out []Issue
	for name, sf := range f.ItemFields {
		if !sub.Has(name) {
			if sf.Required {
				out = append(out, Issue{n.Rel, fmt.Sprintf("%s[%d].%s", key, i, name),
					"missing", "required key is absent", false})
			}
			continue
		}
		out = append(out, retarget(checkScalar(n, s, name, sf, sub.Str(name)), key, i)...)
	}
	return out
}

func retarget(in []Issue, key string, i int) []Issue {
	for j := range in {
		in[j].Key = fmt.Sprintf("%s[%d].%s", key, i, in[j].Key)
		in[j].Fixable = false
	}
	return in
}

// checkSourcesArity applies the per-type citation floor. A decision or incident is a
// first-party record and may cite nothing; everything else needs a source.
func checkSourcesArity(n *Note, s *Schema) []Issue {
	f, ok := s.Fields["sources"]
	if !ok || f.MinItemsByType == nil {
		return nil
	}
	min, ok := f.MinItemsByType[n.FM.Str("type")]
	if !ok || len(n.FM.List("sources")) >= min {
		return nil
	}
	return []Issue{{n.Rel, "sources", "uncited",
		fmt.Sprintf("type %q requires at least %d source", n.FM.Str("type"), min), false}}
}

// checkEngineTrail enforces the T0 invariant at the data layer: recall, write and index
// are static-core stages and may never record a model-backed engine. This is a schema
// error, not a warning — see CLAUDE.md, "Invariants".
func checkEngineTrail(n *Note) []Issue {
	var out []Issue
	for _, e := range n.FM.List("engine_trail") {
		stage, engine, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		if isStaticStage(stage) && engine != "none" {
			out = append(out, Issue{n.Rel, "engine_trail", "engine-invariant",
				fmt.Sprintf("stage %q may only use engine none, found %q", stage, engine), false})
		}
	}
	return out
}

func isStaticStage(s string) bool {
	return slices.Contains(config.LockedStages, s)
}

// checkDates enforces the ordering the schema states as `constraint`.
func checkDates(n *Note) []Issue {
	created, ok := parseDate(n.FM.Str("created"))
	if !ok {
		return nil
	}
	var out []Issue
	for _, key := range []string{"updated", "verified"} {
		d, ok := parseDate(n.FM.Str(key))
		if ok && d.Before(created) {
			out = append(out, Issue{n.Rel, key, "date-order",
				fmt.Sprintf("%s precedes created", key), true})
		}
	}
	return out
}

func parseDate(s string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", strings.Trim(s, `"' `))
	return t, err == nil
}

func isISODate(s string) bool { _, ok := parseDate(s); return ok }
