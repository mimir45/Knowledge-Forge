package sentinel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpsertCreatesFileAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CLAUDE.md")
	if err := Upsert(path, "logback", Markdown, "Relevant notes: [[a]]"); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "<!-- forge:logback:begin -->\nRelevant notes: [[a]]\n<!-- forge:logback:end -->\n"
	if string(first) != want {
		t.Errorf("got %q, want %q", first, want)
	}
	if err := Upsert(path, "logback", Markdown, "Relevant notes: [[a]]"); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Errorf("second Upsert changed the file:\n got %q\nwant %q", second, first)
	}
}

func TestUpsertReplacesInPlaceAndKeepsSurroundingContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CLAUDE.md")
	os.WriteFile(path, []byte("# Module notes\n\nHand-written prose stays here.\n"), 0o644)
	if err := Upsert(path, "logback", Markdown, "Notes: [[a]]"); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(path, "logback", Markdown, "Notes: [[a]] · [[b]]"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	want := "# Module notes\n\nHand-written prose stays here.\n\n" +
		"<!-- forge:logback:begin -->\nNotes: [[a]] · [[b]]\n<!-- forge:logback:end -->\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpsertBeforeAnchorsANewBlockAndLeavesAnExistingOneInPlace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "App.java")
	os.WriteFile(path, []byte("package app;\n\npublic class App {\n}\n"), 0o644)
	if err := UpsertBefore(path, "App.App", Slash, "forge: [[a]]", 3); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	want := "package app;\n\n// forge:App.App:begin\n// forge: [[a]]\n// forge:App.App:end\n" +
		"public class App {\n}\n"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	// A drifted anchorLine must not move a block that already exists.
	if err := UpsertBefore(path, "App.App", Slash, "forge: [[a]] · [[b]]", 1); err != nil {
		t.Fatal(err)
	}
	got, _ = os.ReadFile(path)
	want = "package app;\n\n// forge:App.App:begin\n// forge: [[a]] · [[b]]\n// forge:App.App:end\n" +
		"public class App {\n}\n"
	if string(got) != want {
		t.Errorf("existing block moved on a stale anchor: got %q, want %q", got, want)
	}
}

func TestTwoDistinctIDsCoexistInOneFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "App.java")
	os.WriteFile(path, []byte("package app;\n"), 0o644)
	UpsertBefore(path, "App.foo", Slash, "forge: [[a]]", 1)
	UpsertBefore(path, "App.bar", Slash, "forge: [[b]]", 1)
	if err := Remove(path, "App.foo", Slash); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	want := "// forge:App.bar:begin\n// forge: [[b]]\n// forge:App.bar:end\npackage app;\n"
	if string(got) != want {
		t.Errorf("removing App.foo disturbed App.bar: got %q, want %q", got, want)
	}
}

func TestRemoveRestoresTheFileByteForByteWhenTheBlockWasAppended(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CLAUDE.md")
	before := "# Module notes\n\nHand-written prose stays here.\n"
	os.WriteFile(path, []byte(before), 0o644)
	if err := Upsert(path, "logback", Markdown, "Notes: [[a]]"); err != nil {
		t.Fatal(err)
	}
	if err := Remove(path, "logback", Markdown); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != before {
		t.Errorf("got %q, want the pre-Upsert file %q", got, before)
	}
}

func TestRemoveIsANoOpWithoutAFileOrAnUnwrittenID(t *testing.T) {
	dir := t.TempDir()
	if err := Remove(filepath.Join(dir, "missing.md"), "logback", Markdown); err != nil {
		t.Fatalf("missing file: %v", err)
	}
	path := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(path, []byte("# Notes\n"), 0o644)
	if err := Remove(path, "logback", Markdown); err != nil {
		t.Fatalf("unwritten id: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "# Notes\n" {
		t.Errorf("no-op Remove changed the file: %q", got)
	}
}
