package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/dataset"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// theQuestion is deliberately a phrase no slug or hash could reproduce by accident, so
// "the raw text is absent" is a real assertion rather than a coincidence of tokenisation.
const theQuestion = "why does kafka rebalancing stall on our staging cluster"

func datasetCfg(on bool) *config.Config {
	tags := []string{"d1", "d2", "d3", "d4", "d5"}
	return &config.Config{Dataset: config.Dataset{Enabled: on, Capture: tags}}
}

func TestRecallCapturesD1WhenEnabled(t *testing.T) {
	root := fixtureCopy(t)
	cfg := datasetCfg(true)
	if code := runRecall(root, theQuestion, "java,spring-boot", false, thresholdsFrom(cfg), cfg); code != 0 {
		t.Fatalf("runRecall exit = %d", code)
	}
	line := readOneLine(t, filepath.Join(root, dataset.D1Path))
	for _, want := range []string{`"kind":"d1-routing"`, `"decision":`, `"candidates":`,
		`"stack":["java","spring-boot"]`, `"run_id":"`} {
		if !strings.Contains(line, want) {
			t.Errorf("d1.jsonl missing %s:\n%s", want, line)
		}
	}
}

// TestRecallD1NeverStoresTheQuestion is ADDENDUM §D's invariant stated as a test: "never
// store raw question text — hash + extracted topic only". The slug is derived from the
// question and is allowed; the sentence itself is not.
func TestRecallD1NeverStoresTheQuestion(t *testing.T) {
	root := fixtureCopy(t)
	cfg := datasetCfg(true)
	runRecall(root, theQuestion, "", false, thresholdsFrom(cfg), cfg)
	line := readOneLine(t, filepath.Join(root, dataset.D1Path))
	if strings.Contains(line, theQuestion) {
		t.Errorf("d1.jsonl contains the raw question:\n%s", line)
	}
	if strings.Contains(line, "stall on our staging") {
		t.Errorf("d1.jsonl contains a raw question fragment:\n%s", line)
	}
}

func TestRecallSkipsD1WhenDisabled(t *testing.T) {
	root := fixtureCopy(t)
	cfg := datasetCfg(false) // dataset.enabled off, every tag still listed
	if code := runRecall(root, theQuestion, "", false, thresholdsFrom(cfg), cfg); code != 0 {
		t.Fatalf("runRecall exit = %d", code)
	}
	if _, err := os.Stat(filepath.Join(root, dataset.D1Path)); !os.IsNotExist(err) {
		t.Fatalf("expected no d1.jsonl, got err = %v", err)
	}
}

func TestGateCapturesD5WithProfile(t *testing.T) {
	root := t.TempDir()
	writeProfile(t, root, d5Profile)
	captureAccepted(datasetCfg(true), root, acceptedDraft(t))
	line := readOneLine(t, filepath.Join(root, dataset.D5Path))
	for _, want := range []string{`"kind":"d5-style"`, `"topic":"kafka-rebalancing"`,
		`"note":"the accepted body"`, `"primary_language":"java"`,
		`"frameworks":"spring-boot,quarkus"`, `https://example.test/a`} {
		if !strings.Contains(line, want) {
			t.Errorf("d5.jsonl missing %s:\n%s", want, line)
		}
	}
	// The free-text profile fields are excluded on purpose; see d5ProfileKeys.
	if strings.Contains(line, "acme-internal") {
		t.Errorf("d5.jsonl carried a free-text profile field:\n%s", line)
	}
}

func TestGateSkipsD5WhenTagAbsent(t *testing.T) {
	root := t.TempDir()
	cfg := &config.Config{Dataset: config.Dataset{Enabled: true, Capture: []string{"d4"}}}
	captureAccepted(cfg, root, acceptedDraft(t))
	if _, err := os.Stat(filepath.Join(root, dataset.D5Path)); !os.IsNotExist(err) {
		t.Fatalf("expected no d5.jsonl, got err = %v", err)
	}
}

// TestGateD5SurvivesAMissingProfile pins the normal case on a vault where forge init has
// not run: no profiles/me.md, so the pair is captured with no Profile rather than skipped.
func TestGateD5SurvivesAMissingProfile(t *testing.T) {
	root := t.TempDir()
	captureAccepted(datasetCfg(true), root, acceptedDraft(t))
	line := readOneLine(t, filepath.Join(root, dataset.D5Path))
	if strings.Contains(line, `"profile"`) {
		t.Errorf("expected profile omitted with no me.md:\n%s", line)
	}
}

// TestCaptureD1OutcomeWhenRunIDGiven pins BACKLOG B-035's join half: a --run-id passed
// back from a recall call records the write's actual outcome, keyed by that id.
func TestCaptureD1OutcomeWhenRunIDGiven(t *testing.T) {
	root := t.TempDir()
	captureD1Outcome(datasetCfg(true), root, "abc123", true)
	line := readOneLine(t, filepath.Join(root, dataset.D1OutcomePath))
	for _, want := range []string{`"kind":"d1-outcome"`, `"run_id":"abc123"`, `"published":true`} {
		if !strings.Contains(line, want) {
			t.Errorf("d1-outcomes.jsonl missing %s:\n%s", want, line)
		}
	}
}

func TestCaptureD1OutcomeQuarantined(t *testing.T) {
	root := t.TempDir()
	captureD1Outcome(datasetCfg(true), root, "abc123", false)
	line := readOneLine(t, filepath.Join(root, dataset.D1OutcomePath))
	if !strings.Contains(line, `"published":false`) {
		t.Errorf("d1-outcomes.jsonl missing published:false:\n%s", line)
	}
}

// TestCaptureD1OutcomeSkipsOnEmptyRunID is the degradation contract gate.go's usage text
// promises: a write with no --run-id (the normal case for anything that did not originate
// from a recall call) must cost nothing, not just "cost nothing visible on stdout".
func TestCaptureD1OutcomeSkipsOnEmptyRunID(t *testing.T) {
	root := t.TempDir()
	captureD1Outcome(datasetCfg(true), root, "", true)
	if _, err := os.Stat(filepath.Join(root, dataset.D1OutcomePath)); !os.IsNotExist(err) {
		t.Fatalf("expected no d1-outcomes.jsonl, got err = %v", err)
	}
}

func TestCaptureD1OutcomeSkipsWhenD1Disabled(t *testing.T) {
	root := t.TempDir()
	cfg := &config.Config{Dataset: config.Dataset{Enabled: true, Capture: []string{"d4"}}}
	captureD1Outcome(cfg, root, "abc123", true)
	if _, err := os.Stat(filepath.Join(root, dataset.D1OutcomePath)); !os.IsNotExist(err) {
		t.Fatalf("expected no d1-outcomes.jsonl, got err = %v", err)
	}
}

func acceptedDraft(t *testing.T) *vault.Note {
	t.Helper()
	src := "---\nslug: kafka-rebalancing\nstack: [kafka]\n" +
		"sources:\n  - url: https://example.test/a\n---\n\nthe accepted body\n"
	path := filepath.Join(t.TempDir(), "draft.md")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := vault.Load(path, "notes/concept/kafka-rebalancing.md")
	if err != nil {
		t.Fatal(err)
	}
	n.Body = []byte("the accepted body")
	return n
}

// d5Profile carries one free-text field (code_style) that D5 must not pick up, alongside
// the fixed-shape ones it must.
const d5Profile = "---\nprimary_language: java\nframeworks: [spring-boot, quarkus]\n" +
	"seniority: senior\nexplain_style: mechanism-first\n" +
	"code_style:\n  java: \"acme-internal house rules\"\n---\n\n# Profile\n"

func readOneLine(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", filepath.Base(path), err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) != 1 {
		t.Fatalf("%s: got %d lines, want 1", filepath.Base(path), len(lines))
	}
	return lines[0]
}
