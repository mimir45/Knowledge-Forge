package qualitygate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGateAdversarialDraftIsQuarantinedWithBothDefectsNamed is the plan's mandatory
// acceptance test: a draft with a genuine Java syntax error (missing brace) co-occurring
// with an unresolved import — the exact ordering case compile_test.go pins at the
// CompileCheck layer — plus an unsourced claim, run through the full Run + Quarantine
// pipeline exactly as cmd/forge/gate.go's runGate does. If the gates were decorative,
// this draft would publish clean; it must not.
func TestGateAdversarialDraftIsQuarantinedWithBothDefectsNamed(t *testing.T) {
	if _, err := exec.LookPath("javac"); err != nil {
		t.Skip("javac not installed")
	}
	root := emptyVault(t)

	brokenJava := "import org.springframework.stereotype.Service;\n\n" +
		"public class Snippet {\n    void broken() {\n        if (true) {\n"
	src := mutate(t, "sources:\n  - url: https://kafka.apache.org/documentation/\n    accessed: 2026-08-07\n    kind: official", "sources: []")
	src = strings.Replace(src, "# Kafka consumer group rebalancing\n",
		"# Kafka consumer group rebalancing\n\n"+
			"Kafka 4.1 changed the default rebalance protocol.\n\n"+
			"```java\n"+brokenJava+"```\n", 1)
	draft := noteFrom(t, src, "notes/concept/kafka-consumer-group-rebalancing.md")
	cfg := testConfig(t)

	rep, err := Run(cfg, root, draft, ModeCreate)
	if err != nil {
		t.Fatal(err)
	}
	if o := findOutcome(rep, "code"); o == nil || o.Verdict != Fail {
		t.Fatalf("code gate = %+v, want Fail (syntax error must dominate the co-occurring unresolved import)", o)
	}
	if o := findOutcome(rep, "citation"); o == nil || o.Verdict != Fail || o.Remedy != MarkUnverified {
		t.Fatalf("citation gate = %+v, want Fail/MarkUnverified", o)
	}
	if !rep.Quarantine {
		t.Fatalf("expected Report.Quarantine=true: %+v", rep.Outcomes)
	}

	if err := Quarantine(root, draft, testSchema(t), rep, ModeCreate, ""); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(root, "_inbox", "kafka-consumer-group-rebalancing.md"))
	if err != nil {
		t.Fatalf("expected the draft under _inbox/: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "confidence: low") {
		t.Errorf("quarantined draft missing confidence: low:\n%s", s)
	}
	if !strings.Contains(s, "## Open questions") {
		t.Fatalf("quarantined draft missing ## Open questions:\n%s", s)
	}
	if !strings.Contains(s, "code:") {
		t.Errorf("open questions must name the code defect:\n%s", s)
	}
	if !strings.Contains(s, "citation:") {
		t.Errorf("open questions must name the citation defect:\n%s", s)
	}
}
