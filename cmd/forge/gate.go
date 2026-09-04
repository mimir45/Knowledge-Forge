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
	"github.com/mimir45/Knowledge-Forge/pkg/qualitygate"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

const gateUsage = `usage: forge gate --file PATH --rel VAULT/RELATIVE/PATH.md [--vault DIR]
                   [--mode create|update] [--target-slug SLUG] [--previous-draft FILE]
                   [--run-id ID]

Runs the seven quality gates (pkg/qualitygate.Run) against one candidate
note and prints the JSON Report to stdout. --file is the rendered draft to check — it
need not exist under --vault yet, since CREATE drafts never do. --rel is the note's
intended vault-relative path; the link and duplicate gates need it to know which
directory-group to score the draft against, so it is required even for CREATE.

On a Fail whose remedy blocks the write (Report.Quarantine=true), the draft is written
to _inbox/ with confidence: low and an ## Open questions section naming every failed
gate, then forge index is re-run so the SQLite cache does not go stale. A copy of the
failing draft is also saved under .forge/drafts/ and its path printed to stderr:
pass that path back as --previous-draft on the fix-and-retry run. If that retry then
passes, the (failing draft, gate error, fixed draft) triple is captured as dataset D4
when dataset.capture includes "d4" — this is the only join D4 makes; there is no
slug-based auto-pairing, so a stale or unrelated draft can never pair silently.
--mode update also stamps a supersedes back-pointer to --target-slug, per the
CREATE/UPDATE split: an UPDATE's target note is never touched, but the proposed edit is
not silently dropped either — it lands in _inbox/ for a human to find and apply.

--run-id is optional and pairs with the run_id a preceding forge recall call emitted:
when set and dataset capture includes "d1", this write's outcome
(published or quarantined) is recorded keyed by that id, so export can join a routing
decision to whether the note it led to was actually published. Omitted --run-id is the
normal case for any write that did not originate from a recall call, and costs nothing —
today's behaviour exactly.

Exit 0 = Quarantine false (published cleanly). Exit 1 = Quarantine true (still not an
error — the note was handled correctly, just not published). Exit 2 = usage error.
Exit 3 = internal error: gate execution or the quarantine write itself failed, so the
draft was NOT handled — not published, not quarantined, left untouched at --file.
`

func cmdGate(args []string) int {
	fs := flag.NewFlagSet("forge gate", flag.ContinueOnError)
	file := fs.String("file", "", "path to the rendered draft note")
	rel := fs.String("rel", "", "the note's intended vault-relative path")
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	mode := fs.String("mode", "create", "create or update")
	targetSlug := fs.String("target-slug", "", "update mode: slug of the note being extended")
	prevDraft := fs.String("previous-draft", "", "path from a prior quarantine, to pair for D4")
	runID := fs.String("run-id", "", "run_id from the preceding forge recall call; optional")
	fs.Usage = func() { fmt.Fprint(os.Stderr, gateUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	return runGate(gateArgs{*file, *rel, *vaultDir, *mode, *targetSlug, *prevDraft, *runID})
}

type gateArgs struct{ file, rel, vaultDir, mode, targetSlug, previousDraft, runID string }

func runGate(a gateArgs) int {
	root, m, code := resolveGateInputs(a)
	if code != 0 {
		return code
	}
	cfg, code := configOrExit("gate")
	if code != 0 {
		return code
	}
	draft, s, code := loadDraft(a.file, a.rel)
	if code != 0 {
		return code
	}
	rep, err := qualitygate.Run(cfg, root, draft, m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: %v\n", err)
		return 3
	}
	return reportAndQuarantine(root, cfg, draft, s, rep, m, a)
}

// resolveGateInputs validates --file/--rel are both set and parses --mode.
func resolveGateInputs(a gateArgs) (root string, mode qualitygate.Mode, code int) {
	if a.file == "" || a.rel == "" {
		fmt.Fprintln(os.Stderr, "forge gate: --file and --rel are both required")
		return "", 0, 2
	}
	root, code = vaultOrExit("gate", a.vaultDir)
	if code != 0 {
		return "", 0, code
	}
	switch a.mode {
	case "create":
		return root, qualitygate.ModeCreate, 0
	case "update":
		return root, qualitygate.ModeUpdate, 0
	default:
		fmt.Fprintf(os.Stderr, "forge gate: --mode must be create or update, got %q\n", a.mode)
		return "", 0, 2
	}
}

func loadDraft(file, rel string) (*vault.Note, *vault.Schema, int) {
	abs, err := filepath.Abs(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: %v\n", err)
		return nil, nil, 3
	}
	n, err := vault.Load(abs, rel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: %v\n", err)
		return nil, nil, 3
	}
	s, err := vault.LoadSchema()
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: %v\n", err)
		return nil, nil, 3
	}
	return n, s, 0
}

func reportAndQuarantine(root string, cfg *config.Config, draft *vault.Note, s *vault.Schema,
	rep qualitygate.Report, m qualitygate.Mode, a gateArgs) int {
	b, _ := json.Marshal(rep)
	fmt.Println(string(b))
	if !rep.Quarantine {
		captureRepairIfRetry(cfg, root, draft, a.previousDraft)
		captureAccepted(cfg, root, draft)
		captureD1Outcome(cfg, root, a.runID, true)
		return 0
	}
	if err := qualitygate.Quarantine(root, draft, s, rep, m, a.targetSlug); err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: quarantine write: %v\n", err)
		return 3
	}
	fmt.Fprintf(os.Stderr, "forge gate: quarantined to %s\n", draft.Path)
	captureD1Outcome(cfg, root, a.runID, false)
	saveForRetry(root, draft, qualitygate.OpenQuestions(rep))
	return reindexAfterQuarantine(root, cfg.Paths.Index)
}

// captureD1Outcome joins a --run-id passed back from the recall call that led to this
// write to a D1Outcome record.
func captureD1Outcome(cfg *config.Config, root, runID string, published bool) {
	if runID == "" || cfg == nil || !dataset.D1.Enabled(cfg.Dataset) {
		return
	}
	o := dataset.D1Outcome{Kind: dataset.D1OutcomeKind, RunID: runID, Published: published,
		CapturedAt: time.Now()}
	if err := dataset.AppendD1Outcome(root, o); err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: d1 outcome capture: %v\n", err)
	}
}

// captureRepairIfRetry closes the loop --previous-draft opened: this run just passed.
func captureRepairIfRetry(cfg *config.Config, root string, draft *vault.Note, previousDraft string) {
	if previousDraft == "" {
		return
	}
	failing, gateErr, err := dataset.TakePreviousDraft(previousDraft)
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: d4 capture: %v\n", err)
		return
	}
	if !dataset.D4.Enabled(cfg.Dataset) {
		return
	}
	p := dataset.D4Pair{Kind: dataset.D4Kind, Stage: "gate", FailingDraft: string(failing),
		GateError: string(gateErr), FixedDraft: string(draft.Body), CapturedAt: time.Now()}
	if err := dataset.AppendD4(root, p); err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: d4 capture: %v\n", err)
	}
}

// saveForRetry persists the failing draft under .forge/drafts/ so a fix-and-retry run
// can join back to it via --previous-draft.
func saveForRetry(root string, draft *vault.Note, openQuestions []string) {
	path, err := dataset.SaveFailingDraft(root, draftSlug(draft), draft.Body,
		[]byte(strings.Join(openQuestions, "\n")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "forge gate: save draft for retry: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "forge gate: retry with --previous-draft %s\n", path)
}

// draftSlug prefers the frontmatter slug.
func draftSlug(draft *vault.Note) string {
	if draft.FM != nil {
		if s := draft.FM.Str("slug"); s != "" {
			return s
		}
	}
	return strings.TrimSuffix(filepath.Base(draft.Path), filepath.Ext(draft.Path))
}

func reindexAfterQuarantine(root, indexName string) int {
	if indexName == "" {
		indexName = "_index.md"
	}
	if code := runIndex(root, indexName, 4096, false); code != 0 {
		fmt.Fprintln(os.Stderr, "forge gate: warning: forge index failed after quarantine write")
	}
	return 1
}
