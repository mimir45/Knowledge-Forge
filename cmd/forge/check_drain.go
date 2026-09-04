package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/engine"
	"github.com/mimir45/Knowledge-Forge/pkg/store"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// drainAdvisorQueue is the budget queue drain.
func drainAdvisorQueue(d *checkData) {
	cfg := d.cfg.config
	if cfg == nil || !cfg.Check.DrainAdvisorQueue || cfg.Engines.Budget.OnExhausted != "queue" {
		return
	}
	if d.cfg.offline {
		fmt.Println("\nBudget queue drain: skipped (--offline)")
		return
	}
	st, err := store.Open(d.root)
	if err != nil {
		fmt.Printf("\nBudget queue drain: %v\n", err)
		return
	}
	defer st.Close()
	drainQueuedNotes(d, cfg, st)
}

// drainQueuedNotes walks d.notes rather than re-scanning the vault.
func drainQueuedNotes(d *checkData, cfg *config.Config, st *store.Store) {
	drained, failed := 0, 0
	for _, n := range d.notes {
		if !isQueued(n) {
			continue
		}
		if advisorExhausted(cfg, st) {
			fmt.Println("  advisor budget exhausted mid-drain; remaining notes stay queued")
			break
		}
		if err := drainOne(d.root, n.Rel, cfg, st); err != nil {
			fmt.Printf("  %s: %v\n", n.Rel, err)
			failed++
			continue
		}
		drained++
	}
	fmt.Printf("\nBudget queue drain: %d dispatched, %d failed\n", drained, failed)
}

// advisorExhausted checks the advisor tier's own ledger directly rather than routing
// through engine.Exhausted's stage-chain walk.
func advisorExhausted(cfg *config.Config, st *store.Store) bool {
	remaining, err := st.Remaining("advisor", cfg.Engines.Budget.AdvisorUSDPerDay, time.Now)
	if err != nil {
		return true // lookup failed; treat as exhausted rather than risk overspend
	}
	return remaining <= 0
}

// drainOne re-loads the note fresh.
func drainOne(root, rel string, cfg *config.Config, st *store.Store) error {
	n, s, err := loadNoteAndSchema(root, rel)
	if err != nil {
		return err
	}
	req := engine.Request{Stage: "synthesize", Prompt: string(n.Body)}
	res, err := buildEngine(cfg, "advisor").Run(req)
	if err != nil {
		return err
	}
	if err := st.Spend("advisor", res.CostUSD, time.Now); err != nil {
		return err
	}
	captureD2(cfg, root, "synthesize", string(n.Body), res.Output)
	printCritique(rel, res.Output)
	return vault.SetScalars(n, s, map[string]string{"pending_advisor": "false"})
}

// printCritique surfaces the advisor's verdict for human approval.
func printCritique(rel, output string) {
	var c engine.Critique
	if err := json.Unmarshal([]byte(output), &c); err != nil {
		fmt.Printf("  %s: advisor critique (unparsed): %s\n", rel, output)
		return
	}
	fmt.Printf("  %s: confidence=%s, patch=%v\n", rel, c.Confidence, c.Patch != "")
}
