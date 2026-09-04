package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mimir45/Knowledge-Forge/pkg/engine"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

func cmdEngineRecord(args []string) int {
	fs := flag.NewFlagSet("forge engine record", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	stage := fs.String("stage", "", "pipeline stage the tier ran")
	tier := fs.String("tier", "", "engine tier that ran it: host|api|advisor")
	rel := fs.String("rel", "", "note path, relative to the vault root")
	fs.Usage = func() { fmt.Fprint(os.Stderr, engineRecordUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *stage == "" || *tier == "" || *rel == "" {
		fmt.Fprint(os.Stderr, "forge engine record: --stage, --tier and --rel are required\n\n"+engineRecordUsage)
		return 2
	}
	root, code := vaultOrExit("engine record", *vaultDir)
	if code != 0 {
		return code
	}
	return runEngineRecord(root, *rel, *stage, *tier)
}

const engineRecordUsage = `usage: forge engine record --stage NAME --tier T --rel PATH [--vault DIR]

Stamps engine_trail on one note after a host-tier step completes. The skill is
the caller (SKILL.md's quality-gate step) — a binary that never called the model
cannot otherwise know a host step happened.

`

func runEngineRecord(vaultDir, rel, stage, tier string) int {
	if tier != "none" && isLockedStage(stage) {
		fmt.Fprintf(os.Stderr, "forge engine record: stage %q is locked to \"none\" (T0 static core), "+
			"refusing to stamp tier %q\n", stage, tier)
		return 2
	}
	entry, ok := engine.TrailEntry(stage, tier)
	if !ok {
		fmt.Fprintf(os.Stderr, "forge engine record: stage %q is not stamped in engine_trail\n", stage)
		return 2
	}
	root, err := filepath.Abs(vaultDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge engine record: %v\n", err)
		return 2
	}
	n, s, err := loadNoteAndSchema(root, rel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge engine record: %v\n", err)
		return 1
	}
	if n.FM == nil {
		fmt.Fprintf(os.Stderr, "forge engine record: %v\n", vault.ErrNoFM)
		return 1
	}
	items := mergeTrail(n.FM.List("engine_trail"), stage, entry)
	if err := vault.SetList(n, s, "engine_trail", items); err != nil {
		fmt.Fprintf(os.Stderr, "forge engine record: %v\n", err)
		return 1
	}
	fmt.Printf("%s: engine_trail += %s\n", rel, entry)
	return 0
}

// isLockedStage is a third defense-in-depth check.
func isLockedStage(stage string) bool {
	for _, s := range engine.LockedStages {
		if s == stage {
			return true
		}
	}
	return false
}

func loadNoteAndSchema(root, rel string) (*vault.Note, *vault.Schema, error) {
	n, err := vault.Load(filepath.Join(root, rel), rel)
	if err != nil {
		return nil, nil, err
	}
	s, err := vault.LoadSchema()
	if err != nil {
		return nil, nil, err
	}
	return n, s, nil
}

// mergeTrail replaces any existing entry for stage rather than duplicating it — a note
// re-verified after an edit gets one verify= line, its latest, not a growing history.
func mergeTrail(existing []string, stage, entry string) []string {
	out := []string{entry}
	prefix := stage + "="
	for _, e := range existing {
		if !strings.HasPrefix(e, prefix) {
			out = append(out, e)
		}
	}
	sort.Strings(out)
	return out
}
