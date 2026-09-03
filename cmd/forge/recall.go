package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/dataset"
	"github.com/mimir45/Knowledge-Forge/pkg/recall"
	"github.com/mimir45/Knowledge-Forge/pkg/telemetry"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

func cmdRecall(args []string) int {
	fs := flag.NewFlagSet("forge recall", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
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
	root, code := vaultOrExit("recall", *vaultDir)
	if code != 0 {
		return code
	}
	cfg, code := configOrExit("recall")
	if code != 0 {
		return code
	}
	return runRecall(root, *question, *stack, *explain, thresholdsFrom(cfg), cfg)
}

// thresholdsFrom pulls the decision tree's numbers from the config chain rather than
// from a var in pkg/recall.
func thresholdsFrom(cfg *config.Config) recall.Thresholds {
	t := recall.DefaultThresholds
	if v := cfg.Recall.AnswerThreshold; v > 0 {
		t.Answer = v
	}
	if v := cfg.Recall.UpdateThreshold; v > 0 {
		t.Update = v
	}
	if v := cfg.Recall.NeighbourMinScore; v > 0 {
		t.Neighbour = v
	}
	return t
}

const recallUsage = `usage: forge recall --question "..." [--stack a,b] [--vault DIR] [--explain]

Scores the vault against a question and prints the top 10 candidates as JSON on
stdout, highest first, plus a run_id correlating this call to a later
"forge gate --run-id" write. Deterministic and model-free; see
references/recall-spec.md for the scoring blend and threshold decision tree.

`

func runRecall(vaultDir, question, stack string, explain bool, th recall.Thresholds,
	cfg *config.Config) int {
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
	res := th.ResultFrom(q, recall.RankPool(q, docs, time.Now()))
	if explain {
		printExplain(os.Stderr, q, res)
	}
	runID := telemetry.NewRunID()
	logAsk(root, cfg, q, res, runID)
	captureD1(root, cfg, q, res, runID)
	return emit(res, runID)
}

// logAsk records the telemetry ask event when telemetry is enabled.
func logAsk(root string, cfg *config.Config, q recall.Query, res recall.Result, runID string) {
	if cfg == nil || !cfg.Telemetry.Enabled {
		return
	}
	ev := telemetry.Event{TS: time.Now().UTC(), Event: "ask", QHash: telemetry.QHash(q.Question),
		Topic: vault.Slug(q.Question), Stack: q.Stack, Decision: string(res.Verdict),
		RecallTopScore: res.TopScore, RunID: runID}
	if err := telemetry.Append(root, ev); err != nil {
		fmt.Fprintf(os.Stderr, "forge recall: telemetry: %v\n", err)
	}
}

// captureD1 records the routing pair (see pkg/dataset/d1.go).
func captureD1(root string, cfg *config.Config, q recall.Query, res recall.Result, runID string) {
	if cfg == nil || !dataset.D1.Enabled(cfg.Dataset) {
		return
	}
	p := dataset.D1Pair{Kind: dataset.D1Kind, RunID: runID, QHash: telemetry.QHash(q.Question),
		Topic: vault.Slug(q.Question), Decision: string(res.Verdict), Stack: q.Stack,
		RecallTopScore: res.TopScore, Candidates: len(res.Candidates), CapturedAt: time.Now()}
	if err := dataset.AppendD1(root, p); err != nil {
		fmt.Fprintf(os.Stderr, "forge recall: d1 capture: %v\n", err)
	}
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

// recallEnvelope adds run_id to recall.Result's JSON shape without teaching pkg/recall.
type recallEnvelope struct {
	recall.Result
	RunID string `json:"run_id"`
}

// emit writes the output contract from recall-spec.md §4.
func emit(res recall.Result, runID string) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(recallEnvelope{Result: res, RunID: runID}); err != nil {
		fmt.Fprintf(os.Stderr, "forge recall: %v\n", err)
		return 1
	}
	return 0
}
