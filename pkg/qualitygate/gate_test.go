package qualitygate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mutate returns goodNote with one line's exact text replaced — a small helper so each
// gate test changes exactly one fact instead of hand-maintaining a second full fixture.
func mutate(t *testing.T, old, new string) string {
	t.Helper()
	if !strings.Contains(goodNote, old) {
		t.Fatalf("mutate: %q not found in goodNote", old)
	}
	return strings.Replace(goodNote, old, new, 1)
}

func TestRunCleanDraftPassesUnquarantined(t *testing.T) {
	root := emptyVault(t)
	draft := noteFrom(t, goodNote, "notes/concept/kafka-consumer-group-rebalancing.md")
	rep, err := Run(testConfig(t), root, draft, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Quarantine {
		t.Fatalf("expected Quarantine=false, got true: %+v", rep.Outcomes)
	}
	for _, o := range rep.Outcomes {
		if o.Verdict == Fail {
			t.Errorf("gate %s unexpectedly failed: %s", o.Gate, o.Detail)
		}
	}
	if len(rep.Outcomes) != 7 {
		t.Fatalf("expected 7 gate outcomes, got %d", len(rep.Outcomes))
	}
}

func TestRunSchemaFailureBlocksWrite(t *testing.T) {
	root := emptyVault(t)
	src := mutate(t, "type: concept", "type: bogus-type")
	draft := noteFrom(t, src, "notes/concept/kafka-consumer-group-rebalancing.md")
	rep, err := Run(testConfig(t), root, draft, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Quarantine {
		t.Fatalf("expected Quarantine=true on invalid type: %+v", rep.Outcomes)
	}
	if o := findOutcome(rep, "schema"); o == nil || o.Verdict != Fail || o.Remedy != RetryOnce {
		t.Fatalf("schema outcome = %+v, want Fail/RetryOnce", o)
	}
}

func TestRunCitationFailureBlocksWrite(t *testing.T) {
	root := emptyVault(t)
	src := mutate(t, "sources:\n  - url: https://kafka.apache.org/documentation/\n    accessed: 2026-08-07\n    kind: official", "sources: []")
	draft := noteFrom(t, src, "notes/concept/kafka-consumer-group-rebalancing.md")
	rep, err := Run(testConfig(t), root, draft, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Quarantine {
		t.Fatalf("expected Quarantine=true on empty sources: %+v", rep.Outcomes)
	}
	if o := findOutcome(rep, "citation"); o == nil || o.Verdict != Fail || o.Remedy != MarkUnverified {
		t.Fatalf("citation outcome = %+v, want Fail/MarkUnverified", o)
	}
}

func TestRunFreshnessStaleBlocksWrite(t *testing.T) {
	root := emptyVault(t)
	src := mutate(t, "verified: 2026-08-07", "verified: 2020-01-01")
	src = strings.Replace(src, "freshness_days: 365", "freshness_days: 30", 1)
	draft := noteFrom(t, src, "notes/concept/kafka-consumer-group-rebalancing.md")
	rep, err := Run(testConfig(t), root, draft, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Quarantine {
		t.Fatalf("expected Quarantine=true on a stale verified date: %+v", rep.Outcomes)
	}
	if o := findOutcome(rep, "freshness"); o == nil || o.Verdict != Fail || o.Remedy != DropConfidence {
		t.Fatalf("freshness outcome = %+v, want Fail/DropConfidence", o)
	}
}

func TestRunAntislopBannedPhraseBlocksWrite(t *testing.T) {
	root := emptyVault(t)
	src := mutate(t, "# Kafka consumer group rebalancing\n", "# Kafka consumer group rebalancing\n\nLet's leverage the rebalance protocol.\n")
	draft := noteFrom(t, src, "notes/concept/kafka-consumer-group-rebalancing.md")
	rep, err := Run(testConfig(t), root, draft, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Quarantine {
		t.Fatalf("expected Quarantine=true on a banned phrase: %+v", rep.Outcomes)
	}
	if o := findOutcome(rep, "antislop"); o == nil || o.Verdict != Fail || o.Remedy != RewritePass {
		t.Fatalf("antislop outcome = %+v, want Fail/RewritePass", o)
	}
}

func TestRunDanglingLinkFailsButDoesNotBlockWrite(t *testing.T) {
	root := emptyVault(t)
	src := mutate(t, "# Kafka consumer group rebalancing\n", "# Kafka consumer group rebalancing\n\nSee [[nonexistent-note]].\n")
	draft := noteFrom(t, src, "notes/concept/kafka-consumer-group-rebalancing.md")
	rep, err := Run(testConfig(t), root, draft, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	o := findOutcome(rep, "link")
	if o == nil || o.Verdict != Fail || o.Remedy != DelegateToLibrarian {
		t.Fatalf("link outcome = %+v, want Fail/DelegateToLibrarian", o)
	}
	if rep.Quarantine {
		t.Fatalf("a dangling link must not block the write (librarian fixes it post-write): %+v", rep.Outcomes)
	}
}

func TestRunDuplicateNearMatchFailsButDoesNotBlockWrite(t *testing.T) {
	root := emptyVault(t)
	existing := mutate(t, "slug: kafka-consumer-group-rebalancing", "slug: kafka-consumer-group-rebalancing-2")
	existing = strings.Replace(existing, `title: "Kafka consumer group rebalancing"`, `title: "Kafka consumer group rebalancing, take two"`, 1)
	writeVaultNote(t, root, "notes/concept/existing.md", existing)

	draft := noteFrom(t, goodNote, "notes/concept/kafka-consumer-group-rebalancing.md")
	cfg := testConfig(t)
	rep, err := Run(cfg, root, draft, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	o := findOutcome(rep, "duplicate")
	if o == nil || o.Verdict != Fail || o.Remedy != SwitchToUpdate {
		t.Fatalf("duplicate outcome = %+v, want Fail/SwitchToUpdate", o)
	}
	if rep.Quarantine {
		t.Fatalf("a near-duplicate must not block the write (routing recommendation only): %+v", rep.Outcomes)
	}
}

// TestRunDeterminism pins the B-020 convention: two runs on byte-identical inputs must
// produce byte-identical Outcomes, in the same order, so a retry's open-questions bullets
// never drift for reasons unrelated to the draft itself.
func TestRunDeterminism(t *testing.T) {
	root := emptyVault(t)
	draft1 := noteFrom(t, goodNote, "notes/concept/kafka-consumer-group-rebalancing.md")
	draft2 := noteFrom(t, goodNote, "notes/concept/kafka-consumer-group-rebalancing.md")
	cfg := testConfig(t)

	rep1, err := Run(cfg, root, draft1, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	rep2, err := Run(cfg, root, draft2, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	j1, _ := json.Marshal(rep1)
	j2, _ := json.Marshal(rep2)
	if string(j1) != string(j2) {
		t.Fatalf("Run is not deterministic on unchanged state:\n%s\n!=\n%s", j1, j2)
	}
}

func findOutcome(rep Report, gate string) *Outcome {
	for i := range rep.Outcomes {
		if rep.Outcomes[i].Gate == gate {
			return &rep.Outcomes[i]
		}
	}
	return nil
}

func writeVaultNote(t *testing.T, root, rel, src string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}
