package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"knowledge-forge/pkg/report"
)

// writeStatsAskLog writes .forge/log.jsonl with one "ask" line per (topic, count) pair —
// the same event shape countAskTopics (check_asklog.go) scans for. Named distinctly from
// check_asklog_test.go's own writeAskLog (different signature, same package).
func writeStatsAskLog(t *testing.T, root string, topics map[string]int) {
	t.Helper()
	dir := filepath.Join(root, ".forge")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	for topic, n := range topics {
		for i := 0; i < n; i++ {
			line, _ := json.Marshal(struct {
				Event string `json:"event"`
				Topic string `json:"topic"`
			}{"ask", topic})
			buf.Write(line)
			buf.WriteByte('\n')
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "log.jsonl"), buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRunStatsHappyPath: "hibernate" (a real fixture slug) resolves written, an unknown
// topic asked twice surfaces as a gap.
func TestRunStatsHappyPath(t *testing.T) {
	root := fixtureCopy(t)
	writeStatsAskLog(t, root, map[string]int{"hibernate": 3, "totally-unknown-topic": 2})

	var out bytes.Buffer
	if code := runStats(root, &out); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	got := out.String()

	if !strings.Contains(got, "Hit rate:") {
		t.Errorf("missing hit-rate line:\n%s", got)
	}
	if !strings.Contains(got, "hibernate") || !strings.Contains(got, "(written)") {
		t.Errorf("missing written topic:\n%s", got)
	}
	if !strings.Contains(got, "totally-unknown-topic") {
		t.Errorf("missing gap topic in most-asked:\n%s", got)
	}
	gapsSection := got[strings.Index(got, "Gaps ("):]
	if !strings.Contains(gapsSection, "totally-unknown-topic") || !strings.Contains(gapsSection, "2x") {
		t.Errorf("gap not reported:\n%s", got)
	}
}

// TestRunStatsNoLogGracefulEmpty: a vault with no .forge/log.jsonl at all must render an
// empty-but-valid report, never an error.
func TestRunStatsNoLogGracefulEmpty(t *testing.T) {
	root := fixtureCopy(t)

	var out bytes.Buffer
	if code := runStats(root, &out); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	got := out.String()
	if !strings.Contains(got, "none recorded yet") {
		t.Errorf("want empty-topics message, got:\n%s", got)
	}
	if !strings.Contains(got, "no weekly snapshots yet") {
		t.Errorf("want empty-trend message, got:\n%s", got)
	}
}

// TestRunStatsWeeklyTrend seeds .forge/weekly-stats.json directly via WeeklyStore so the
// trend section renders a real row.
func TestRunStatsWeeklyTrend(t *testing.T) {
	root := fixtureCopy(t)
	if err := os.MkdirAll(filepath.Join(root, ".forge"), 0o755); err != nil {
		t.Fatal(err)
	}
	store := report.OpenWeeklyStore(filepath.Join(root, ".forge"))
	key := report.WeekKey(time.Now())
	store.Record(key, report.VaultStats{Notes: 13, HitRate: 42.5, Orphans: 3, Drift: 1})
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if code := runStats(root, &out); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	got := out.String()
	if !strings.Contains(got, key) || !strings.Contains(got, "13") {
		t.Errorf("missing weekly trend row:\n%s", got)
	}
}

// TestCmdStatsExitsNonzeroOnMissingVault: unlike the fail-silent hook subcommands, forge
// stats is a direct user command — a bad --vault must surface a nonzero exit via
// vaultOrExit's existing error path, not swallow it.
func TestCmdStatsExitsNonzeroOnMissingVault(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "does-not-exist")
	if code := cmdStats([]string{"--vault", bad}); code == 0 {
		t.Errorf("exit = 0, want nonzero for a missing vault")
	}
}
