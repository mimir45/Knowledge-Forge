package qualitygate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestOpenQuestionsOneBulletPerFailInGateOrder pins the determinism guarantee at
// its source: OpenQuestions must walk rep.Outcomes in Run's fixed gate order, not re-sort
// or filter differently, so two runs on unchanged state produce byte-identical bullets.
func TestOpenQuestionsOneBulletPerFailInGateOrder(t *testing.T) {
	rep := Report{Outcomes: []Outcome{
		{Gate: "schema", Verdict: Pass},
		{Gate: "citation", Verdict: Fail, Detail: "0 source(s), type requires at least 1"},
		{Gate: "code", Verdict: Skipped},
		{Gate: "freshness", Verdict: Fail, Detail: "stale"},
		{Gate: "antislop", Verdict: Pass},
		{Gate: "link", Verdict: Fail, Detail: "1 dangling link(s)"},
		{Gate: "duplicate", Verdict: Pass},
	}}
	got := OpenQuestions(rep)
	want := []string{
		"citation: 0 source(s), type requires at least 1",
		"freshness: stale",
		"link: 1 dangling link(s)",
	}
	if len(got) != len(want) {
		t.Fatalf("OpenQuestions = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("OpenQuestions[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestOpenQuestionsNoFailuresIsEmpty(t *testing.T) {
	rep := Report{Outcomes: []Outcome{{Gate: "schema", Verdict: Pass}, {Gate: "code", Verdict: Skipped}}}
	if got := OpenQuestions(rep); len(got) != 0 {
		t.Fatalf("OpenQuestions = %v, want empty", got)
	}
}

func TestQuarantineCreateWritesToInboxLowConfidence(t *testing.T) {
	root := emptyVault(t)
	draft := noteFrom(t, goodNote, "notes/concept/kafka-consumer-group-rebalancing.md")
	rep := Report{Outcomes: []Outcome{{Gate: "schema", Verdict: Fail, Remedy: RetryOnce, Detail: "bad type"}}}

	if err := Quarantine(root, draft, testSchema(t), rep, ModeCreate, ""); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(root, "_inbox", "kafka-consumer-group-rebalancing.md"))
	if err != nil {
		t.Fatalf("expected a file under _inbox/: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "confidence: low") {
		t.Errorf("quarantined draft missing confidence: low:\n%s", s)
	}
	if !strings.Contains(s, "## Open questions") || !strings.Contains(s, "schema: bad type") {
		t.Errorf("quarantined draft missing open questions bullet:\n%s", s)
	}
	if !strings.Contains(s, "supersedes: []") {
		t.Errorf("CREATE mode must not overwrite the draft's own supersedes field:\n%s", s)
	}
}

// TestQuarantineUpdateSetsSupersedesBackPointer pins the plan's CREATE/UPDATE split: an
// UPDATE draft that fails gate must not touch the published note it was proposing to
// change, but must not be silently dropped either — it lands in _inbox/ with a
// supersedes-style back-pointer to the note a human can later apply it to.
func TestQuarantineUpdateSetsSupersedesBackPointer(t *testing.T) {
	root := emptyVault(t)
	draft := noteFrom(t, goodNote, "notes/concept/kafka-consumer-group-rebalancing.md")
	rep := Report{Outcomes: []Outcome{{Gate: "citation", Verdict: Fail, Remedy: MarkUnverified, Detail: "0 sources"}}}

	if err := Quarantine(root, draft, testSchema(t), rep, ModeUpdate, "kafka-partitions"); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(root, "_inbox", "kafka-consumer-group-rebalancing.md"))
	if err != nil {
		t.Fatalf("expected a file under _inbox/: %v", err)
	}
	if !strings.Contains(string(out), "supersedes: kafka-partitions") {
		t.Errorf("UPDATE quarantine must carry a supersedes back-pointer to the target slug:\n%s", out)
	}
}
