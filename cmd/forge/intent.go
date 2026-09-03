package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/recall"
)

const intentUsage = `usage: forge intent [--vault DIR]

UserPromptSubmit hook. Recall-scores the prompt against the vault with zero model calls
and, only when the top hit scores at or above 0.50, prints it as additionalContext.

stdin: a UserPromptSubmit JSON payload. This command reads exactly one field:

    user_prompt   string   the prompt text to score

stdout, only above the gate: {"additionalContext": "...", "continue": true}
Silence is the normal case: below the gate, on an empty prompt, on malformed stdin, or
on an unresolvable vault, this command prints nothing.
Fail-silent: the exit code is always 0.
`

// cmdIntent is Phase 5's UserPromptSubmit hook: a cheap.
func cmdIntent(args []string) int {
	fs := flag.NewFlagSet("forge intent", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	fs.SetOutput(io.Discard)
	// flag's own error path must stay silent (fail-silent contract), so Usage is stubbed
	// and an explicit -h/--help is handled below instead — in any flag position.
	fs.Usage = func() {}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			// stderr, not stdout: stdout is this hook's JSON output channel.
			fmt.Fprint(os.Stderr, intentUsage)
			fs.SetOutput(os.Stderr)
			fs.PrintDefaults()
		}
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
// prompt for. Derived in references/recall-spec.md, guarded by intent-gate.golden.
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
	_ = json.NewEncoder(os.Stdout).Encode(out)
}
