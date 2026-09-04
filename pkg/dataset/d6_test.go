package dataset

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
)

// seedD6Vault writes one repo's code index cache plus one note citing it.
func seedD6Vault(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	ix := codeindex.Index{Repo: "payment-svc", Commit: "abc123", Extractor: codeindex.Extractor,
		Files: map[string]codeindex.File{
			"service/Payment.java": {Path: "service/Payment.java", Lang: "java",
				Symbols: []codeindex.Symbol{{Name: "PaymentService", Kind: "class", Start: 1, End: 40}}},
		}}
	cache := filepath.Join(root, ".forge", "code-index-payment-svc.json")
	if err := codeindex.Save(cache, ix); err != nil {
		t.Fatal(err)
	}
	note := "---\ntype: concept\n---\n\n" +
		"Cites `service/Payment.java` directly, and the `PaymentService` class by name.\n"
	notePath := filepath.Join(root, "notes", "concept", "payment.md")
	if err := os.MkdirAll(filepath.Dir(notePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(notePath, []byte(note), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestD6RendersFromCodeIndexAndCitations is D6's slice of the exhaustiveness check
// TestEveryTierRendersInEveryDefinedFormat runs for D1-D5: it proves loadTier, idOf.
func TestD6RendersFromCodeIndexAndCitations(t *testing.T) {
	root := seedD6Vault(t)
	out := filepath.Join(t.TempDir(), "export")
	rep, body := exportTo(t, root, ExportOptions{Set: D6Tag, Format: FormatSFT, Out: out})
	if rep.Records != 2 {
		t.Fatalf("got %d records, want 2 (one path citation, one symbol citation): %s",
			rep.Records, body)
	}
	if !strings.Contains(body, `path: service/Payment.java`) ||
		!strings.Contains(body, `PaymentService`) {
		t.Errorf("body missing an expected pair:\n%s", body)
	}
	sheet := readFile(t, filepath.Join(out, rep.Datasheet))
	if !strings.Contains(sheet, "## Limitations") || !strings.Contains(sheet, "Derived, not captured") {
		t.Errorf("datasheet missing D6's limitations:\n%s", sheet)
	}
}

// TestD6DedupesRepeatedCitations pins pairsFromNotes' seen-set: citing the same symbol
// twice (once in code_refs, once in body prose) must not double its weight in the corpus.
func TestD6DedupesRepeatedCitations(t *testing.T) {
	root := seedD6Vault(t)
	note := "---\ntype: concept\ncode_refs:\n  - \"payment-svc:service/Payment.java\"\n---\n\n" +
		"Also cites `service/Payment.java` again in prose.\n"
	if err := os.WriteFile(filepath.Join(root, "notes", "concept", "payment.md"),
		[]byte(note), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "export")
	rep, _ := exportTo(t, root, ExportOptions{Set: D6Tag, Format: FormatSFT, Out: out})
	if rep.Records != 1 {
		t.Errorf("got %d records for one citation repeated twice, want 1 (deduped)", rep.Records)
	}
}

// TestD6RefusesSince and TestD6RefusesAnonymize pin refuseDerivedOptions: both are exit
// before a record is read, both leave --out untouched.
func TestD6RefusesSince(t *testing.T) {
	root := seedD6Vault(t)
	out := filepath.Join(t.TempDir(), "export")
	since, err := time.Parse("2006-01-02", "2026-01-01")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Export(root, ExportOptions{Set: D6Tag, Format: FormatSFT, Since: since, Out: out})
	assertRefusedDerived(t, err, out, root)
}

func TestD6RefusesAnonymize(t *testing.T) {
	root := seedD6Vault(t)
	out := filepath.Join(t.TempDir(), "export")
	_, err := Export(root, ExportOptions{Set: D6Tag, Format: FormatSFT, Anonymize: true, Out: out})
	assertRefusedDerived(t, err, out, root)
}

func assertRefusedDerived(t *testing.T, err error, out, root string) {
	t.Helper()
	var bad UsageError
	if err == nil || !errors.As(err, &bad) {
		t.Fatalf("want a UsageError, got %v", err)
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Errorf("--out exists after a refused request: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, ExportLogPath)); !os.IsNotExist(statErr) {
		t.Errorf("exports.jsonl gained a line for a request refused before any record was read")
	}
}

// TestD6FailsClosedOnUnreadableCodeIndex pins loadIndexes' fail-closed rule: a cache
// Load cannot read.
func TestD6FailsClosedOnUnreadableCodeIndex(t *testing.T) {
	root := seedD6Vault(t)
	stale := codeindex.Index{Repo: "other-svc", Commit: "x", Extractor: codeindex.Extractor - 1,
		Files: map[string]codeindex.File{"a.java": {Path: "a.java", Lang: "java",
			Symbols: []codeindex.Symbol{{Name: "A"}}}}}
	if err := codeindex.Save(filepath.Join(root, ".forge", "code-index-other-svc.json"), stale); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "export")
	_, err := Export(root, ExportOptions{Set: D6Tag, Format: FormatSFT, Out: out})
	if err == nil {
		t.Fatal("Export succeeded over a stale-Extractor code index, want failure")
	}
	if !strings.Contains(err.Error(), "code-index-other-svc.json") {
		t.Errorf("error must name the unreadable cache, got: %v", err)
	}
}
