package vault

import (
	"errors"
	"maps"
	"os"
	"slices"
)

// ErrNoFM is returned when a caller tries to stamp a note that has no frontmatter block
// to stamp. Drift meets these in the wild — the fixture has three — and must skip them
// rather than invent a header.
var ErrNoFM = errors.New("note has no frontmatter")

// SetScalars rewrites one note's frontmatter with the given scalar values set, leaving
// the body byte-identical. Drift is one caller: it stamps drift_checked_at and, on a
// demotion or a restore, moves confidence. `forge engine run`'s on_exhausted: queue path
// is the other: it stamps pending_advisor: true. Neither caller may move anything else
// about the note. Writing through render() keeps schema key order, so a stamped note
// still passes `forge validate`.
func SetScalars(n *Note, s *Schema, kv map[string]string) error {
	if n.FM == nil {
		return ErrNoFM
	}
	for _, k := range slices.Sorted(maps.Keys(kv)) {
		setScalar(n.FM, k, kv[k])
	}
	out, err := render(n.FM, s, n.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(n.Path, out, 0o644)
}

// SetList rewrites one frontmatter list key, leaving every other key — including anything
// SetScalars would touch — untouched. `forge engine record` is the caller, replacing
// engine_trail without reimplementing render()'s key-order and body-preservation
// guarantees.
func SetList(n *Note, s *Schema, key string, items []string) error {
	if n.FM == nil {
		return ErrNoFM
	}
	if !n.FM.Has(key) {
		n.FM.Keys = append(n.FM.Keys, key)
	}
	n.FM.Vals[key] = seqNode(items)
	out, err := render(n.FM, s, n.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(n.Path, out, 0o644)
}
