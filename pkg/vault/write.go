package vault

import (
	"errors"
	"maps"
	"os"
	"slices"
)

// ErrNoFM is returned when a caller tries to stamp a note that has no frontmatter block
// to stamp.
var ErrNoFM = errors.New("note has no frontmatter")

// SetScalars rewrites one note's frontmatter with the given scalar values set, leaving
// the body byte-identical.
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

// SetList rewrites one frontmatter list key, leaving every other key — including
// anything SetScalars would touch — untouched.
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
