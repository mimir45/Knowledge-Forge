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

// printIntent surfaces the top vault hit only above 0.7. BACKLOG B-008: this threshold
// can false-positive near 0.740 because the tags/stack channel weighting the fix
// prescribed didn't fix either calibration case — a known, un-fixed risk left for the
// §3.1 recalibration, not this hook.
func printIntent(root, prompt string) {
	docs, err := loadDocs(root)
	if err != nil || len(docs) == 0 {
		return
	}
	q := recall.Query{Question: prompt}
	cands := recall.Rank(q, docs, time.Now())
	if len(cands) == 0 || cands[0].Score <= 0.7 {
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
