package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"knowledge-forge/pkg/recall"
)

// cmdIntent is Phase 5's UserPromptSubmit hook: a cheap, model-free check for whether
// the vault already answers the prompt. It must add under 50ms — reusing loadDocs'
// warm SQLite cache, exactly as forge recall does, is what makes that budget plausible
// — and, like session-context, must never break the session: any failure here just
// means no hint is offered, never a blocked prompt or a nonzero exit.
func cmdIntent(args []string) int {
	fs := flag.NewFlagSet("forge intent", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return 0
	}
	prompt, err := readPrompt(os.Stdin)
	if err != nil || prompt == "" {
		return 0
	}
	if root, err := resolveVault(*vaultDir); err == nil {
		printIntent(root, prompt)
	}
	return 0
}

// readPrompt decodes the UserPromptSubmit stdin payload; forge intent only needs the
// one field it names `user_prompt`.
func readPrompt(r io.Reader) (string, error) {
	var p struct {
		UserPrompt string `json:"user_prompt"`
	}
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return "", err
	}
	return p.UserPrompt, nil
}

// intentGate is the score at or above which the top vault hit is worth interrupting a
// prompt for. It is not the neighbour floor and must not be derived like one: a wrong
// neighbour is one link in a note already under review, while a wrong intent hit injects
// context into a live session on a hook contracted never to disturb it. The floor is
// derived for recall; this is derived for precision.
//
// Measured, testdata/intent-gate-labels.txt: 25 prompts labelled FIRE (the vault answers
// it) or QUIET (it does not — ten of them adjacent-topic hard negatives). The old 0.7
// admitted 3 of 10 FIRE prompts, dropping one at 0.652 that matches a note title almost
// verbatim. That is the decay this constant already suffered once, silently, when B-008
// moved the scale under it.
//
// The labels rule 0.7 out and cannot choose its replacement: the classes separate at
// 0.402/0.407, so every value in [0.405, 0.7] has identical — zero — false positives.
// Within that range 0.50 is the lowest value still a clear step above the QUIET ceiling
// (0.402, ~24% headroom) that admits every FIRE prompt whose phrasing tracks a note
// title. It recovers three the old gate dropped: 0.546, 0.533 and 0.517, all near-verbatim
// titles, which is precisely the case this hook exists for. 8 of 10, and
// TestIntentGateSeparation fails if that number falls.
//
// Deliberately NOT tied to DefaultThresholds.Update, which was the tempting tidy answer:
// printIntent computes no verdict, and emitIntentHit says the vault "may" already answer
// this, so there is no claim about CREATE for Update to keep honest. Binding to it would
// have cost 3 of 10 FIRE prompts for an alignment nothing in this function asserts, and
// coupled hook behaviour to a config key this path never reads.
//
// It is a constant rather than a config key for a reason that does survive: printIntent
// runs on UserPromptSubmit under a 50ms budget and loads no config today, and wiring the
// four-layer chain in here buys a knob nobody turns at the cost of the one budget in the
// tree that is actually tight.
//
// Note the asymmetry with BACKLOG B-033's neighbour floor, derived in the same session:
// same root cause, opposite trade. A wrong neighbour is one link in a note already under
// review, so the floor is derived for recall; a wrong intent hit interrupts a live session
// on a hook contracted never to disturb it, so this is derived for precision first and
// recall only inside what precision leaves free.
const intentGate = 0.50

// printIntent surfaces the top vault hit only at or above intentGate.
func printIntent(root, prompt string) {
	docs, err := loadDocs(root)
	if err != nil || len(docs) == 0 {
		return
	}
	q := recall.Query{Question: prompt}
	cands := recall.Rank(q, docs, time.Now())
	if len(cands) == 0 || cands[0].Score < intentGate {
		return
	}
	emitIntentHit(cands[0])
}

// emitIntentHit prints the JSON output schema UserPromptSubmit hooks use to inject
// context: additionalContext is what Claude sees, continue:true never blocks the prompt.
func emitIntentHit(c recall.Candidate) {
	msg := fmt.Sprintf("The vault may already answer this — %q (%s, score %.2f).",
		c.Title, c.Path, c.Score)
	out := struct {
		AdditionalContext string `json:"additionalContext"`
		Continue          bool   `json:"continue"`
	}{AdditionalContext: msg, Continue: true}
	// Fail-silent like the rest of this hook: a stdout encode failure here has no
	// recovery that doesn't risk corrupting the JSON UserPromptSubmit expects, so it's
	// swallowed rather than surfaced (mirrors logSessionContext's own rationale).
	_ = json.NewEncoder(os.Stdout).Encode(out)
}
