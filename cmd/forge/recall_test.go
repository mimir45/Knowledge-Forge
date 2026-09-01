package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
)

// TestRecallLogsAskWhenTelemetryEnabled and its disabled counterpart pin the gate that
// belongs in the caller, not pkg/telemetry itself (writer.go's doc comment).
func TestRecallLogsAskWhenTelemetryEnabled(t *testing.T) {
	root := fixtureCopy(t)
	cfg := &config.Config{Telemetry: config.Telemetry{Enabled: true}}
	if code := runRecall(root, "kafka rebalancing", "", false, thresholdsFrom(cfg), cfg); code != 0 {
		t.Fatalf("runRecall exit = %d", code)
	}
	if _, err := os.Stat(filepath.Join(root, ".forge", "log.jsonl")); err != nil {
		t.Fatalf("expected log.jsonl to exist: %v", err)
	}
}

// TestRecallEmitsRunID pins BACKLOG B-035's envelope addition: every forge recall call
// prints a non-empty run_id on stdout, distinct from call to call, so a caller can thread
// it back through forge gate --run-id without the two ever colliding.
func TestRecallEmitsRunID(t *testing.T) {
	root := fixtureCopy(t)
	cfg := &config.Config{}
	out1 := captureStdout(t, func() { runRecall(root, "kafka rebalancing", "", false, thresholdsFrom(cfg), cfg) })
	out2 := captureStdout(t, func() { runRecall(root, "kafka rebalancing", "", false, thresholdsFrom(cfg), cfg) })

	var env1, env2 struct {
		RunID string `json:"run_id"`
	}
	if err := json.Unmarshal([]byte(out1), &env1); err != nil {
		t.Fatalf("decoding first envelope: %v\n%s", err, out1)
	}
	if err := json.Unmarshal([]byte(out2), &env2); err != nil {
		t.Fatalf("decoding second envelope: %v\n%s", err, out2)
	}
	if env1.RunID == "" {
		t.Error("run_id is empty")
	}
	if env1.RunID == env2.RunID {
		t.Errorf("two calls minted the same run_id: %s", env1.RunID)
	}
	// The envelope must still carry every field recall-spec.md §4 already documents.
	for _, want := range []string{`"verdict"`, `"top_score"`, `"candidates"`, `"neighbours"`} {
		if !strings.Contains(out1, want) {
			t.Errorf("envelope missing %s:\n%s", want, out1)
		}
	}
}

func TestRecallSkipsLogWhenTelemetryDisabled(t *testing.T) {
	root := fixtureCopy(t)
	cfg := &config.Config{Telemetry: config.Telemetry{Enabled: false}}
	if code := runRecall(root, "kafka rebalancing", "", false, thresholdsFrom(cfg), cfg); code != 0 {
		t.Fatalf("runRecall exit = %d", code)
	}
	if _, err := os.Stat(filepath.Join(root, ".forge", "log.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("expected no log.jsonl, got err = %v", err)
	}
}
