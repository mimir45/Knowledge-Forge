package main

import (
	"fmt"

	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/engine"
	"github.com/mimir45/Knowledge-Forge/pkg/report"
)

// aiPass runs check.ai_pass's three sub-tasks when enabled.
func aiPass(d *checkData) {
	if d.cfg.config == nil || !d.cfg.config.Check.AIPass {
		return
	}
	fmt.Println("\nAI pass (check.ai_pass) — print-only, no auto-apply:")
	aiPassDraftRefresh(d)
	aiPassDuplicateMerge(d)
	aiPassADRStub(d)
}

func aiPassDraftRefresh(d *checkData) {
	f, ok := topBroken(d.findings)
	if !ok {
		fmt.Println("  draft refresh: no BROKEN drift findings")
		return
	}
	req := engine.Request{
		Stage:  "synthesize",
		Prompt: "Draft a refreshed note body reconciling this BROKEN citation.",
		Context: map[string]string{
			"note": f.Note, "ref": f.Ref, "reason": f.Reason,
		},
	}
	printInstruction("draft refresh", req)
}

func aiPassDuplicateMerge(d *checkData) {
	p, ok := report.TopDuplicatePair(d.pairs)
	if !ok {
		fmt.Println("  duplicate-merge: no pair clears the 0.85 similarity threshold")
		return
	}
	req := engine.Request{
		Stage:  "synthesize",
		Prompt: "Propose a merge of these two near-duplicate notes.",
		Context: map[string]string{
			"a": p.A, "b": p.B, "score": fmt.Sprintf("%.2f", p.Score),
		},
	}
	printInstruction("duplicate-merge proposal", req)
}

func aiPassADRStub(d *checkData) {
	u, ok := report.TopUncovered(d.allUncovered())
	if !ok {
		fmt.Println("  ADR stub: no undocumented churny module found")
		return
	}
	req := engine.Request{
		Stage:  "plan",
		Prompt: "Draft an ADR stub for this undocumented, actively-changing module.",
		Context: map[string]string{
			"symbol": u.Symbol, "path": u.Path,
			"loc": fmt.Sprintf("%d", u.LOC), "commits": fmt.Sprintf("%d", u.Commits),
		},
	}
	printInstruction("ADR stub", req)
}

func printInstruction(label string, req engine.Request) {
	res, err := (engine.Host{}).Run(req)
	if err != nil {
		fmt.Printf("  %s: %v\n", label, err)
		return
	}
	fmt.Printf("  %s (%s):\n    %s\n", label, req.Stage, res.Instruction)
}

// topBroken picks the top BROKEN finding by (Note, Ref) — the same lexicographic tiebreak
// pattern pkg/report's groupByNote callers use, kept local since only this file needs it.
func topBroken(fs []drift.Finding) (drift.Finding, bool) {
	var top drift.Finding
	found := false
	for _, f := range fs {
		if f.Verdict != drift.Broken {
			continue
		}
		if !found || f.Note < top.Note || (f.Note == top.Note && f.Ref < top.Ref) {
			top, found = f, true
		}
	}
	return top, found
}
