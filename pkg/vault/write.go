package vault

import (
	"errors"
	"os"
	"sort"
)

// ErrNoFM is returned when a caller tries to stamp a note that has no frontmatter block
// to stamp. Drift meets these in the wild — the fixture has three — and must skip them
// rather than invent a header.
var ErrNoFM = errors.New("note has no frontmatter")

// SetScalars rewrites one note's frontmatter with the given scalar values set, leaving
// the body byte-identical. Drift is the caller: it stamps drift_checked_at and, on a
// demotion or a restore, moves confidence — and nothing else about the note may move
// with it. Writing through render() keeps schema key order, so a stamped note still
// passes `forge validate`.
func SetScalars(n *Note, s *Schema, kv map[string]string) error {
	if n.FM == nil {
		return ErrNoFM
	}
	for _, k := range sortedKeys(kv) {
		setScalar(n.FM, k, kv[k])
	}
	out, err := render(n.FM, s, n.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(n.Path, out, 0o644)
}

func sortedKeys(kv map[string]string) []string {
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
