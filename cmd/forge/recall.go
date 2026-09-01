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
// from a var in pkg/recall. A zero means the key was
// absent from every layer, which cannot happen with the packaged layer present but would
// silently resolve everything to ANSWER_FROM_VAULT if it did — so each falls back to the
// compiled-in default rather than to zero.
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

// logAsk records the telemetry ask event when telemetry is enabled. Sources and
// DurationMS stay zero: forge recall has no research-time or citation-count signal to
// report — a known limitation, not an omission, until a caller upstream supplies one.
//
// It takes the whole Query rather than the question string so Stack gets filled. That
// field has existed on Event since early on and production never wrote it, which
// quietly starved two real readers: pkg/report/index.go:67 and coverage.go:43 both fan
// out over e.Stack.
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

// captureD1 records the routing pair (see pkg/dataset/d1.go). Its gate is deliberately
// not logAsk's: telemetry.enabled consents to a local ask log, dataset.capture consents to
// building a corpus meant to be exported, and one is not the other. Everything written
// here is already in the telemetry event — hash and slug, never the question — plus the
// candidate count, which is the one routing feature the log has no field for.
//
// A capture error only reaches stderr. Recall has already scored the vault correctly at
// this point and the caller is waiting on that answer; a side-channel write must not cost
// it. Same posture as captureD2 (engine_run.go) and captureRepairIfRetry (gate.go).
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

// recallEnvelope adds run_id to recall.Result's JSON shape without teaching pkg/recall —
// a zero-model-call, deterministic scoring package — anything about dataset capture or
// telemetry. Embedding promotes Result's own fields, so this is purely additive to the
// documented envelope (recall-spec.md §4); RunID is the only field this package adds.
type recallEnvelope struct {
	recall.Result
	RunID string `json:"run_id"`
}

// emit writes the output contract from recall-spec.md §4: one object carrying the
// verdict, the candidates, and a run_id a caller can thread back through `forge gate
// --run-id` to join this call to the note write it led to. A caller
// that ignores the field gets exactly today's behaviour; the join is optional on both
// ends. `candidates` and `neighbours` are always arrays, never `null` — a vault that
// matches nothing prints `[]` for both, so no consumer has to special-case the empty case
// it will hit on every genuinely new topic.
func emit(res recall.Result, runID string) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(recallEnvelope{Result: res, RunID: runID}); err != nil {
		fmt.Fprintf(os.Stderr, "forge recall: %v\n", err)
		return 1
	}
	return 0
}
