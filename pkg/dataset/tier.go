package dataset

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// Tier is one of the six datasets, described in one place so a seventh
// cannot be added by copying a file. It replaces the pair of hand-written gate functions
// this package shipped through Phase 6 — Enabled() hardcoded D2Tag despite its general
// name, and D4Enabled() existed only because of that (d4.go's own comment said so).
// Anyone adding a tier by reaching for the general-sounding one would silently have
// taken D2's gate — the same defect this package also closed from the config side:
// `dataset.capture` accepting five tiers while only some of them gated anything.
//
// Derived marks D6: it has no Path, no capture gate worth checking (D3
// Enabled call sites, above, are the only real callers of Enabled and none names D6),
// and no per-record timestamp. loadTier's D6 case ignores Path entirely; the field stays
// on the struct, empty, rather than splitting into a second type, because Tiers() must
// keep returning one slice export and dataset-stats can iterate without learning a
// second shape.
type Tier struct {
	// Tag is always load-bearing (--set matching, ExportReport.Set, the datasheet
	// switch); only its *gating* role — matching a cfg.Dataset.Capture entry — goes
	// unused when Derived, since a derived tier is never captured and so is never gated.
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

// Tiers returns all six in d1…d6 order. Export and dataset-stats iterate this, so a new
// tier appears in both without either learning its name — D6 included, since both
// callers' per-tier handling already branches on Tag or on Err, not on an assumption
// that every tier accumulates over time (see loadTier's D6 case and dataset-stats'
// header wording).
func Tiers() []Tier { return []Tier{D1, D2, D3, D4, D5, D6} }

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
