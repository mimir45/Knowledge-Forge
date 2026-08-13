package main

import (
	"os"
	"path/filepath"
	"testing"

	"knowledge-forge/pkg/config"
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
