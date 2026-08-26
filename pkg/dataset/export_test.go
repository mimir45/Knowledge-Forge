package dataset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// seeded is the fixture AUDIT §8.4 D-6 is really about: one record carrying every shape
// the scrubber is supposed to catch, plus the structural fields anonymize.go rewrites.
var seeded = D5Pair{
	Kind:  D5Kind,
	Topic: "payment-outbox",
	Rel:   "notes/concept/payment-outbox.md",
	Note: "Mail dev@acme-corp.example.com about /Users/someone/work/svc/Outbox.java.\n" +
		"Key sk-abcdefghij1234567890, digest a3f5b1c9d7e2048516273849506172839a0b1c2d.\n" +
		"Runbook: https://wiki.acme.internal/outbox and http://localhost:8080/health.\n",
	Sources:    []string{"https://build.acme.corp/job/42", "https://kafka.apache.org/docs"},
	Profile:    map[string]string{"primary_language": "java", "infra": "/Users/someone/k8s"},
	CapturedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
}

func seedVault(t *testing.T, tier Tier, recs ...any) string {
	t.Helper()
	root := t.TempDir()
	for _, r := range recs {
		if err := tier.Append(root, r); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func exportTo(t *testing.T, root string, o ExportOptions) (ExportReport, string) {
	t.Helper()
	rep, err := Export(root, o)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(o.Out, rep.OutFile))
	if err != nil {
		t.Fatalf("reading export: %v", err)
	}
	return rep, string(b)
}

// TestAnonymizeRemovesEverySeededSecret is the D-6 regression guard. The round-trip check
// in export_records.go proves shape survived redaction; only this proves content did not.
func TestAnonymizeRemovesEverySeededSecret(t *testing.T) {
	root := seedVault(t, D5, seeded)
	out := filepath.Join(t.TempDir(), "export")
	rep, body := exportTo(t, root, ExportOptions{Set: D5Tag, Format: FormatSFT,
		Anonymize: true, Out: out})

	for _, secret := range []string{"dev@acme-corp.example.com", "/Users/someone",
		"sk-abcdefghij1234567890", "a3f5b1c9d7e2048516273849506172839a0b1c2d",
		"wiki.acme.internal", "localhost:8080", "build.acme.corp"} {
		if strings.Contains(body, secret) {
			t.Errorf("export still contains %q:\n%s", secret, body)
		}
	}
	if strings.Contains(body, "payment-outbox.md") {
		t.Errorf("the note path was not hashed:\n%s", body)
	}
	if !strings.Contains(body, "notes/concept/") {
		t.Errorf("the note type was dropped along with the slug:\n%s", body)
	}
	if rep.Redactions == 0 {
		t.Error("report claims zero redactions on a fixture full of them")
	}
	// The public source must survive: over-redaction is a correctness bug too.
	if !strings.Contains(body, "kafka.apache.org") {
		t.Errorf("a public source URL was redacted:\n%s", body)
	}
}

func TestExportWithoutAnonymizeKeepsRawText(t *testing.T) {
	root := seedVault(t, D5, seeded)
	out := filepath.Join(t.TempDir(), "export")
	rep, body := exportTo(t, root, ExportOptions{Set: D5Tag, Format: FormatSFT, Out: out})
	if !strings.Contains(body, "dev@acme-corp.example.com") {
		t.Error("--no-anonymize path redacted anyway; the two paths must stay distinct")
	}
	if rep.Anonymized {
		t.Error("report claims anonymized on a raw export")
	}
	sheet := readFile(t, filepath.Join(out, rep.Datasheet))
	if !strings.Contains(sheet, "**no** — this export contains raw captured text") {
		t.Errorf("datasheet does not flag the raw export:\n%s", sheet)
	}
}

// TestExportFailsClosedOnAnUnparseableLine is scrub_test.go's TestScrubFailsClosed for the
// export path. A line nobody can parse is a line nobody could redact, so it must abort the
// run — with --out never created, not created-and-empty.
func TestExportFailsClosedOnAnUnparseableLine(t *testing.T) {
	root := seedVault(t, D5, seeded, seeded)
	appendRaw(t, filepath.Join(root, D5Path), "{\"kind\":\"d5-style\",\"note\":\n")
	out := filepath.Join(t.TempDir(), "export")

	_, err := Export(root, ExportOptions{Set: D5Tag, Format: FormatSFT, Anonymize: true, Out: out})
	if err == nil {
		t.Fatal("Export succeeded over a torn line, want failure")
	}
	if !strings.Contains(err.Error(), "d5.jsonl:3") {
		t.Errorf("error must name the file and line so it is fixable, got: %v", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("--out exists after a failed export: %v", err)
	}
}

func TestExportRefusesAnUndefinedFormatCombination(t *testing.T) {
	root := seedVault(t, D5, seeded)
	out := filepath.Join(t.TempDir(), "export")
	for _, c := range []struct{ set, format string }{
		{D5Tag, string(FormatDPO)}, {D2Tag, string(FormatDPO)},
		{D1Tag, string(FormatDPO)}, {D3Tag, string(FormatCSV)},
	} {
		_, err := Export(root, ExportOptions{Set: c.set, Format: Format(c.format), Out: out})
		if err == nil {
			t.Errorf("--set %s --format %s was accepted, want refusal", c.set, c.format)
		}
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Error("a refused combination created --out")
	}
}

func TestExportRefusesAnUnknownSet(t *testing.T) {
	_, err := Export(t.TempDir(), ExportOptions{Set: "d9", Format: FormatSFT, Out: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "unknown --set") {
		t.Errorf("got %v, want an unknown-set error", err)
	}
}

func TestExportSinceFiltersByRecordTime(t *testing.T) {
	old := seeded
	old.CapturedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	root := seedVault(t, D5, old, seeded)
	out := filepath.Join(t.TempDir(), "export")
	rep, body := exportTo(t, root, ExportOptions{Set: D5Tag, Format: FormatSFT,
		Since: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), Out: out})
	if rep.Records != 1 || rep.Available != 2 {
		t.Errorf("got %d of %d records, want 1 of 2", rep.Records, rep.Available)
	}
	if strings.Count(body, "\n") != 1 {
		t.Errorf("expected one line in the export:\n%s", body)
	}
	if rep.DroppedBy == "" {
		t.Error("report does not say what --since dropped")
	}
}

// TestExportIsLoggedBeforeItIsWritten pins ADDENDUM §D.4's "every export is logged".
func TestExportIsLoggedBeforeItIsWritten(t *testing.T) {
	root := seedVault(t, D5, seeded)
	out := filepath.Join(t.TempDir(), "export")
	exportTo(t, root, ExportOptions{Set: D5Tag, Format: FormatSFT, Anonymize: true, Out: out})
	log := readFile(t, filepath.Join(root, ExportLogPath))
	if !strings.Contains(log, `"set":"d5"`) || !strings.Contains(log, `"anonymized":true`) {
		t.Errorf("export log missing the run:\n%s", log)
	}
	if strings.Contains(log, "payment-outbox") {
		t.Errorf("export log leaked record content:\n%s", log)
	}
}

func TestExportOfAnAbsentTierIsEmptyNotAnError(t *testing.T) {
	out := filepath.Join(t.TempDir(), "export")
	rep, body := exportTo(t, t.TempDir(), ExportOptions{Set: D2Tag, Format: FormatSFT, Out: out})
	if rep.Records != 0 || body != "" {
		t.Errorf("got %d records and %q, want an empty export", rep.Records, body)
	}
}

// TestEveryTierRendersInEveryDefinedFormat is the exhaustiveness check the type switches
// cannot give us at compile time: a tier added to Tiers() with no case in sftOf/loadTier
// fails here rather than in someone's export.
func TestEveryTierRendersInEveryDefinedFormat(t *testing.T) {
	for _, tier := range Tiers() {
		// D6 has no capture record to seed — seedVault's tier.Append would try to open a
		// directory as a file, since a derived tier's Path is empty. Its own exhaustiveness
		// coverage is TestD6RendersFromCodeIndexAndCitations below, over its real fixture
		// shape (a code index cache plus a citing note) instead of a JSONL sample. A
		// *second* derived tier would also skip here silently — it owes its own render
		// test the same way D6 does, this loop cannot discover a missing one for it.
		if tier.Derived {
			continue
		}
		root := seedVault(t, tier, sampleFor(tier))
		for _, f := range formatsFor[tier.Tag] {
			out := filepath.Join(t.TempDir(), "export")
			rep, body := exportTo(t, root, ExportOptions{Set: tier.Tag, Format: f,
				Anonymize: true, Out: out})
			if rep.Records != 1 || body == "" {
				t.Errorf("%s/%s produced %d records, body %q", tier.Tag, f, rep.Records, body)
			}
			if sheet := readFile(t, filepath.Join(out, rep.Datasheet)); !strings.Contains(
				sheet, "## Limitations") {
				t.Errorf("%s/%s datasheet has no Limitations section", tier.Tag, f)
			}
		}
	}
}

func sampleFor(t Tier) any {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	switch t.Tag {
	case D1Tag:
		return D1Pair{Kind: D1Kind, QHash: "abc123", Topic: "kafka-rebalancing",
			Decision: "ANSWER_FROM_VAULT", Stack: []string{"java"}, RecallTopScore: 0.91,
			Candidates: 4, CapturedAt: now}
	case D2Tag:
		return D2Pair{Kind: D2Kind, Stage: "verify", Draft: "d", Critique: "c", CapturedAt: now}
	case D3Tag:
		return Pair{Kind: D3Kind, Note: "notes/concept/x.md", Topic: "x", Generated: "g",
			Preferred: "p", Stack: []string{"java"}, EngineTrail: []string{"api"},
			GenCommit: "aaa", EditCommit: "bbb", EditedAt: now}
	case D4Tag:
		return D4Pair{Kind: D4Kind, Stage: "gate", FailingDraft: "bad", GateError: "schema",
			FixedDraft: "good", CapturedAt: now}
	}
	return seeded
}

// TestDPOChoosesThePreferredSide guards the one thing a preference corpus cannot get
// wrong: swapping chosen and rejected trains the model backwards and nothing downstream
// would notice.
func TestDPOChoosesThePreferredSide(t *testing.T) {
	root := seedVault(t, D3, sampleFor(D3))
	out := filepath.Join(t.TempDir(), "export")
	_, body := exportTo(t, root, ExportOptions{Set: D3Tag, Format: FormatDPO, Out: out})
	if !strings.Contains(body, `"chosen":"p"`) || !strings.Contains(body, `"rejected":"g"`) {
		t.Errorf("D3 DPO put the generated note on the chosen side:\n%s", body)
	}
}

// TestAnonymizeDoesNotMutateItsInput pins the slice- and map-copying in anonymize.go. A
// type switch copies the struct but not the arrays behind its slices, so an in-place
// redaction would reach back into the record read off disk. Nothing depends on that
// today — anonymizeAll discards the originals — which is exactly why it needs a test:
// the first caller to compare a corpus before and after redaction would otherwise get
// two identical tables and no clue why.
func TestAnonymizeDoesNotMutateItsInput(t *testing.T) {
	in := seeded
	in.Stack = []string{"/Users/someone/sdk"}
	in.Profile = map[string]string{"infra": "/Users/someone/k8s"}
	if _, n := anonymizeRecord(in); n == 0 {
		t.Fatal("fixture redacted nothing; the assertions below would be vacuous")
	}
	if in.Stack[0] != "/Users/someone/sdk" {
		t.Errorf("the input's stack slice was rewritten in place: %q", in.Stack[0])
	}
	if in.Profile["infra"] != "/Users/someone/k8s" {
		t.Errorf("the input's profile map was rewritten in place: %q", in.Profile["infra"])
	}
}

func TestAnonymizeBlanksCommitSHAs(t *testing.T) {
	p, _ := anonymizeRecord(sampleFor(D3))
	got := p.(Pair)
	if got.GenCommit != "" || got.EditCommit != "" {
		t.Errorf("commit SHAs survived: %q %q", got.GenCommit, got.EditCommit)
	}
	if got.Note == "notes/concept/x.md" {
		t.Error("the note path was not hashed")
	}
}

func appendRaw(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		t.Fatal(err)
	}
}

// TestD1ExportJoinsOutcomeByRunID is BACKLOG B-035's export-side regression guard: a pair
// whose run_id matches an outcome record renders that outcome; one that doesn't stays
// unjoined, and the strict reader must not choke on either shape.
func TestD1ExportJoinsOutcomeByRunID(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	joined := D1Pair{Kind: D1Kind, RunID: "run-joined", QHash: "aaa", Topic: "kafka",
		Decision: "CREATE", RecallTopScore: 0.3, Candidates: 2, CapturedAt: now}
	unjoined := D1Pair{Kind: D1Kind, RunID: "run-orphan", QHash: "bbb", Topic: "docker",
		Decision: "CREATE", RecallTopScore: 0.2, Candidates: 1, CapturedAt: now}

	root := seedVault(t, D1, joined, unjoined)
	if err := AppendD1Outcome(root, D1Outcome{Kind: D1OutcomeKind, RunID: "run-joined",
		Published: true, CapturedAt: now}); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "export")
	rep, body := exportTo(t, root, ExportOptions{Set: D1Tag, Format: FormatSFT, Out: out})
	if rep.D1Joined != 1 {
		t.Errorf("D1Joined = %d, want 1", rep.D1Joined)
	}
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2:\n%s", len(lines), body)
	}
	if !strings.Contains(lines[0], `"outcome":true`) {
		t.Errorf("joined record missing outcome:true:\n%s", lines[0])
	}
	if strings.Contains(lines[1], `"outcome"`) {
		t.Errorf("unjoined record rendered an outcome field:\n%s", lines[1])
	}

	sheet := readFile(t, filepath.Join(out, rep.Datasheet))
	if !strings.Contains(sheet, "50% (1 of 2 records)") {
		t.Errorf("datasheet does not state the join rate:\n%s", sheet)
	}
}

// TestD1ExportLastOutcomeWinsOnRepair pins joinD1Outcomes's documented last-wins rule:
// a quarantine-then-repair sequence (forge gate --previous-draft, re-passing --run-id per
// SKILL.md's Stage 4) writes two D1Outcome records sharing one RunID, and export must
// report the later one — the retry's actual disposition, not the original quarantine.
func TestD1ExportLastOutcomeWinsOnRepair(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	pair := D1Pair{Kind: D1Kind, RunID: "run-repaired", QHash: "ddd", Topic: "kafka",
		Decision: "CREATE", CapturedAt: now}
	root := seedVault(t, D1, pair)
	// Quarantine first, repair second — the order forge gate would actually append them.
	if err := AppendD1Outcome(root, D1Outcome{Kind: D1OutcomeKind, RunID: "run-repaired",
		Published: false, CapturedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := AppendD1Outcome(root, D1Outcome{Kind: D1OutcomeKind, RunID: "run-repaired",
		Published: true, CapturedAt: now.Add(time.Minute)}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "export")
	_, body := exportTo(t, root, ExportOptions{Set: D1Tag, Format: FormatSFT, Out: out})
	if !strings.Contains(body, `"outcome":true`) {
		t.Errorf("expected the repair's outcome (true) to win over the original "+
			"quarantine (false):\n%s", body)
	}
}

// TestD1ExportOutcomeInCSV pins the CSV lane's outcome column alongside the SFT lane's.
func TestD1ExportOutcomeInCSV(t *testing.T) {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	quarantined := D1Pair{Kind: D1Kind, RunID: "run-q", QHash: "ccc", Topic: "docker",
		Decision: "CREATE", CapturedAt: now}
	root := seedVault(t, D1, quarantined)
	if err := AppendD1Outcome(root, D1Outcome{Kind: D1OutcomeKind, RunID: "run-q",
		Published: false, CapturedAt: now}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "export")
	_, body := exportTo(t, root, ExportOptions{Set: D1Tag, Format: FormatCSV, Out: out})
	if !strings.Contains(body, "outcome") || !strings.Contains(body, "quarantined") {
		t.Errorf("CSV export missing the outcome column or value:\n%s", body)
	}
}

// TestD1ExportSurvivesAnUnreadableOutcomeFile checks the strict reader is wired the same
// way for d1-outcomes.jsonl as it is for every capture file: a torn line aborts the
// export rather than silently dropping the join.
func TestD1ExportSurvivesAnUnreadableOutcomeFile(t *testing.T) {
	root := seedVault(t, D1, sampleFor(D1))
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	if err := AppendD1Outcome(root, D1Outcome{Kind: D1OutcomeKind, RunID: "x",
		Published: true, CapturedAt: now}); err != nil {
		t.Fatal(err)
	}
	appendRaw(t, filepath.Join(root, D1OutcomePath), "{\"kind\":\n")
	out := filepath.Join(t.TempDir(), "export")
	_, err := Export(root, ExportOptions{Set: D1Tag, Format: FormatSFT, Out: out})
	if err == nil {
		t.Fatal("Export succeeded over a torn outcome line, want failure")
	}
	if !strings.Contains(err.Error(), "d1-outcomes.jsonl:2") {
		t.Errorf("error must name the file and line, got: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}
