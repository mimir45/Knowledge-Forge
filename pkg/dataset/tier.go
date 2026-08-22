package dataset

import (
	"encoding/json"
	"os"
	"path/filepath"

	"knowledge-forge/pkg/config"
)

// Tier is one of ADDENDUM §D.1's five datasets, described in one place so a sixth cannot
// be added by copying a file. It replaces the pair of hand-written gate functions this
// package shipped through Phase 6 — Enabled() hardcoded D2Tag despite its general name,
// and D4Enabled() existed only because of that (d4.go's own comment said so). Anyone
// adding a tier by reaching for the general-sounding one would silently have taken D2's
// gate; BACKLOG B-030 is the same defect seen from the config side.
type Tier struct {
	Tag  string // the cfg.Dataset.Capture entry that gates this tier
	Kind string // the "kind" field stamped on every record
	Path string // vault-relative JSONL file
}

// The registry. Built from each tier's own constants rather than restating them, so the
// per-tier files stay the place a reader looks for what a tier means.
var (
	D1 = Tier{Tag: D1Tag, Kind: D1Kind, Path: D1Path}
	D2 = Tier{Tag: D2Tag, Kind: D2Kind, Path: D2Path}
	D3 = Tier{Tag: D3Tag, Kind: D3Kind, Path: D3Path}
	D4 = Tier{Tag: D4Tag, Kind: D4Kind, Path: D4Path}
	D5 = Tier{Tag: D5Tag, Kind: D5Kind, Path: D5Path}
)

// Tiers returns the five in d1…d5 order. Export and dataset-stats iterate this, so a new
// tier appears in both without either learning its name.
func Tiers() []Tier { return []Tier{D1, D2, D3, D4, D5} }

// Enabled reports whether this tier may capture. Both gates are checked here on purpose:
// dataset.enabled is the master switch and dataset.capture is the per-tier list, and
// taking the config struct rather than the bare list makes the master switch impossible
// for a call site to forget. Neither of the two Phase 6 call sites checked it.
func (t Tier) Enabled(d config.Dataset) bool {
	if !d.Enabled {
		return false
	}
	for _, c := range d.Capture {
		if c == t.Tag {
			return true
		}
	}
	return false
}

// Append writes one record as a JSONL line under vaultRoot, creating .forge/datasets/ on
// first use. Distinct from the package-level Append(path, []Pair), which is D3's only and
// dedupes on Key() because a post-commit hook can legitimately re-fire; the four
// CLI-triggered tiers have no re-fire to guard against and append unconditionally.
func (t Tier) Append(vaultRoot string, rec any) error {
	return appendJSONL(filepath.Join(vaultRoot, t.Path), rec)
}

func appendJSONL(path string, rec any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return f.Sync()
}
