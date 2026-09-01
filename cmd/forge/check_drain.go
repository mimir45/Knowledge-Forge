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

// drainAdvisorQueue is the budget queue drain: pending_advisor:true notes
// get their deferred T3 pass batched into this one scheduled run instead of blocking an
// interactive session. Gated on both the config flag and on_exhausted:queue —
// draining notes queued under any other policy would spend budget the config
// never asked this run to spend. --offline skips it the same way deadlinks.md skips its
// own network probes: this is a real HTTP call to the advisor tier, not a local check.
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

// drainQueuedNotes walks d.notes rather than re-scanning the vault — the same collected
// pass every other check.go job reuses. It stops, rather than erroring, the moment
// today's advisor budget is exhausted again mid-drain: the remaining notes stay queued
// for next week exactly as if this run had never touched them.
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
// through engine.Exhausted's stage-chain walk: drainOne always dispatches to the advisor
// tier regardless of what pipeline.synthesize names (the packaged default sets that
// stage's engine to "host"), so gating on "synthesize"'s chain would have skipped every
// candidate and never fired — this checks the tier that actually runs.
func advisorExhausted(cfg *config.Config, st *store.Store) bool {
	remaining, err := st.Remaining("advisor", cfg.Engines.Budget.AdvisorUSDPerDay, time.Now)
	if err != nil {
		return true // lookup failed; treat as exhausted rather than risk overspend
	}
	return remaining <= 0
}

// drainOne re-loads the note fresh (queueNote's own pattern in engine_run.go) rather
// than reusing the in-memory copy, sends its body to the real advisor tier, books the
// spend, and clears the flag only once both the call and the spend succeed — a note
// must never come off the queue with nothing recorded for it. The critique is captured
// via D2 and printed for approval, matching aiPass's own no-auto-apply posture —
// a T3 pass whose output is silently dropped is not a queue drain worth having.
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

// printCritique surfaces the advisor's verdict for human approval — the critique itself
// is never auto-applied. A critique that fails to parse still prints, since silence on a
// spent dollar is worse than raw JSON.
func printCritique(rel, output string) {
	var c engine.Critique
	if err := json.Unmarshal([]byte(output), &c); err != nil {
		fmt.Printf("  %s: advisor critique (unparsed): %s\n", rel, output)
		return
	}
	fmt.Printf("  %s: confidence=%s, patch=%v\n", rel, c.Confidence, c.Patch != "")
}
