package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/engine"
	"github.com/mimir45/Knowledge-Forge/pkg/store"
)

func cmdEngineSelect(args []string) int {
	fs := flag.NewFlagSet("forge engine select", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	stage := fs.String("stage", "", "pipeline stage to resolve, e.g. research")
	asJSON := fs.Bool("json", false, "print the result as JSON")
	fs.Usage = func() { fmt.Fprint(os.Stderr, engineSelectUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *stage == "" {
		fmt.Fprint(os.Stderr, "forge engine select: --stage is required\n\n"+engineSelectUsage)
		return 2
	}
	root, code := vaultOrExit("engine select", *vaultDir)
	if code != 0 {
		return code
	}
	cfg, code := configOrExit("engine select")
	if code != 0 {
		return code
	}
	return runEngineSelect(root, cfg, *stage, *asJSON)
}

const engineSelectUsage = `usage: forge engine select --stage NAME [--json] [--vault DIR]

Dry-runs the resolution chain for one pipeline stage: no HTTP, no spend. Exit 0
always — "can_generate: false" under an offline config is the correct answer,
not a failure.

`

// selectResult is forge engine select --json's shape. Engine carries the literal winning
// name ("local" included) so a caller can tell that case apart from a plain "api"; Tier is
// the narrowed value forge engine run actually dispatches on.
type selectResult struct {
	Stage       string `json:"stage"`
	Engine      string `json:"engine"`
	Tier        string `json:"tier"`
	CanGenerate bool   `json:"can_generate"`
	Reason      string `json:"reason"`
}

func runEngineSelect(vaultDir string, cfg *config.Config, stage string, asJSON bool) int {
	root, err := filepath.Abs(vaultDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge engine select: %v\n", err)
		return 2
	}
	st, err := store.Open(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge engine select: cache: %v\n", err)
		return 1
	}
	defer st.Close()
	res, err := resolveResult(cfg, st, stage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge engine select: %v\n", err)
		return 2
	}
	return printSelect(res, asJSON)
}

// resolveResult calls Resolve and Select rather than duplicating either's logic — both are
// pure reads against the same (cfg, ledger, clock), so calling twice costs nothing a caller
// would notice and keeps tierOf's mapping (e.g. "local"->api) out of cmd/forge entirely.
func resolveResult(cfg *config.Config, ledger engine.Ledger, stage string) (selectResult, error) {
	name, reason, err := engine.Resolve(cfg, ledger, time.Now, stage)
	if err != nil {
		return selectResult{}, err
	}
	tier, _, err := engine.Select(cfg, ledger, time.Now, stage)
	if err != nil {
		return selectResult{}, err
	}
	return selectResult{
		Stage: stage, Engine: name, Tier: string(tier),
		CanGenerate: name != "none", Reason: reason,
	}, nil
}

func printSelect(res selectResult, asJSON bool) int {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return 0
	}
	fmt.Printf("%s: engine=%s tier=%s can_generate=%v (%s)\n",
		res.Stage, res.Engine, res.Tier, res.CanGenerate, res.Reason)
	return 0
}
