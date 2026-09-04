package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/dataset"
)

func seedD1(t *testing.T, lines ...string) string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, dataset.D1Path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := ""
	for _, l := range lines {
		body += l + "\n"
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

const goodD1 = `{"kind":"d1-routing","q_hash":"a","topic":"t","decision":"CREATE_NEW",` +
	`"recall_top_score":0.2,"candidates":1,"captured_at":"2026-08-01T12:00:00Z"}`

// TestExportExitCodesDistinguishUsageFromFailure.
func TestExportExitCodesDistinguishUsageFromFailure(t *testing.T) {
	root := seedD1(t, goodD1)
	out := filepath.Join(t.TempDir(), "export")
	bad := dataset.ExportOptions{Set: "d1", Format: dataset.FormatDPO, Out: out}
	if code := runExportDataset(root, bad); code != 2 {
		t.Errorf("undefined format combination exited %d, want 2", code)
	}

	torn := seedD1(t, goodD1, `{"kind":"d1-routing",`)
	if code := runExportDataset(torn, dataset.ExportOptions{Set: "d1",
		Format: dataset.FormatCSV, Anonymize: true, Out: out}); code != 3 {
		t.Errorf("torn capture file exited %d, want 3", code)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("--out exists after both failures: %v", err)
	}
}

func TestExportDatasetD6RefusesSinceAndAnonymize(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(t.TempDir(), "export")
	since := dataset.ExportOptions{Set: "d6", Format: dataset.FormatSFT, Out: out,
		Since: mustParseDate(t, "2026-01-01")}
	if code := runExportDataset(root, since); code != 2 {
		t.Errorf("d6 --since exited %d, want 2", code)
	}
	anon := dataset.ExportOptions{Set: "d6", Format: dataset.FormatSFT, Out: out, Anonymize: true}
	if code := runExportDataset(root, anon); code != 2 {
		t.Errorf("d6 --anonymize exited %d, want 2", code)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("--out exists after both refusals: %v", err)
	}
}

func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

func TestExportSucceedsAndWritesBothFiles(t *testing.T) {
	root := seedD1(t, goodD1)
	out := filepath.Join(t.TempDir(), "export")
	code := runExportDataset(root, dataset.ExportOptions{Set: "d1",
		Format: dataset.FormatCSV, Anonymize: true, Out: out})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	for _, name := range []string{"d1-csv.csv", "d1-csv-datasheet.md"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Errorf("missing %s: %v", name, err)
		}
	}
}

// TestAnonymizeChoice covers the split AUDIT 8.4 D-6 turns on: the config default, the
// two overrides, and the refusal to guess when both are given.
func TestAnonymizeChoice(t *testing.T) {
	on := &config.Config{Dataset: config.Dataset{AnonymizeOnExport: true}}
	off := &config.Config{Dataset: config.Dataset{AnonymizeOnExport: false}}
	cases := []struct {
		name     string
		cfg      *config.Config
		anon, no bool
		want     bool
		wantExit int
	}{
		{"config default on", on, false, false, true, 0},
		{"config default off", off, false, false, false, 0},
		{"--anonymize overrides an off default", off, true, false, true, 0},
		{"--no-anonymize overrides an on default", on, false, true, false, 0},
		{"both flags is a usage error", on, true, true, false, 2},
	}
	for _, c := range cases {
		got, code := anonymizeChoice(c.cfg, c.anon, c.no)
		if got != c.want || code != c.wantExit {
			t.Errorf("%s: got (%v, %d), want (%v, %d)", c.name, got, code, c.want, c.wantExit)
		}
	}
}

func TestParseSinceRejectsANonDate(t *testing.T) {
	if _, err := parseSince("last tuesday"); err == nil {
		t.Error("parseSince accepted a non-date")
	}
	if got, err := parseSince(""); err != nil || !got.IsZero() {
		t.Errorf("empty --since should mean no lower bound, got %v %v", got, err)
	}
}
