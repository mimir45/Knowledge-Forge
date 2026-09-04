package dataset

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// Tier is one of the six datasets, described in one place so a seventh cannot be added
// by copying a file.
type Tier struct {
	// Tag is always load-bearing (--set matching, ExportReport.Set, the datasheet switch);
	// only its *gating* role — matching a cfg.Dataset.Capture entry.
	Tag     string
	Kind    string // the "kind" field stamped on every record
	Path    string // vault-relative JSONL file; empty when Derived
	Derived bool   // true for D6: recomputed on every read, never captured
}

// The registry. Built from each tier's own constants rather than restating them, so the
// per-tier files stay the place a reader looks for what a tier means.
var (
	D1 = Tier{Tag: D1Tag, Kind: D1Kind, Path: D1Path}
	D2 = Tier{Tag: D2Tag, Kind: D2Kind, Path: D2Path}
	D3 = Tier{Tag: D3Tag, Kind: D3Kind, Path: D3Path}
	D4 = Tier{Tag: D4Tag, Kind: D4Kind, Path: D4Path}
	D5 = Tier{Tag: D5Tag, Kind: D5Kind, Path: D5Path}
	D6 = Tier{Tag: D6Tag, Kind: D6Kind, Derived: true}
)

// Tiers returns all six in d1…d6 order.
func Tiers() []Tier { return []Tier{D1, D2, D3, D4, D5, D6} }

// Enabled reports whether this tier may capture.
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

// Append writes one record as a JSONL line under vaultRoot, creating .forge/datasets/
// on first use.
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
