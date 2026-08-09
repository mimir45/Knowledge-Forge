package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"knowledge-forge/pkg/recall"
)

func cmdRecall(args []string) int {
	fs := flag.NewFlagSet("forge recall", flag.ContinueOnError)
	vaultDir := fs.String("vault", ".", "vault root")
	question := fs.String("question", "", "the question to match against the vault")
	stack := fs.String("stack", "", "comma-separated stack hints, e.g. java,spring-boot")
	explain := fs.Bool("explain", false, "print the per-candidate score breakdown to stderr")
	fs.Usage = func() { fmt.Fprint(os.Stderr, recallUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*question) == "" {
		fmt.Fprint(os.Stderr, "forge recall: --question is required\n\n"+recallUsage)
		return 2
	}
	return runRecall(*vaultDir, *question, *stack, *explain)
}

const recallUsage = `usage: forge recall --question "..." [--stack a,b] [--vault DIR] [--explain]

Scores the vault against a question and prints the top 10 candidates as JSON on
stdout, highest first. Deterministic and model-free; see references/recall-spec.md
for the scoring blend and DESIGN 5.3's decision tree.

`

func runRecall(vaultDir, question, stack string, explain bool) int {
	root, err := filepath.Abs(vaultDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge recall: %v\n", err)
		return 2
	}
	docs, err := loadDocs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge recall: %v\n", err)
		return 1
	}
	q := recall.Query{Question: question, Stack: splitStack(stack)}
	res := recall.DefaultThresholds.Result(q, recall.Rank(q, docs, time.Now()))
	if explain {
		printExplain(os.Stderr, q, res)
	}
	return emit(res)
}

func splitStack(s string) []string {
	var out []string
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// emit writes the output contract from recall-spec.md §4: one object carrying the
// verdict and the candidates. `candidates` and `neighbours` are always arrays, never
// `null` — a vault that matches nothing prints `[]` for both, so no consumer has to
// special-case the empty case it will hit on every genuinely new topic.
func emit(res recall.Result) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		fmt.Fprintf(os.Stderr, "forge recall: %v\n", err)
		return 1
	}
	return 0
}
