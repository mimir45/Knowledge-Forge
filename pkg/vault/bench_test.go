package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// benchNote is a note of roughly the real vault's shape: a full schema frontmatter block
// and about 3 KB of body carrying the wikilinks and code spans the graph and drift passes
// both read back out.
func benchNote() []byte {
	var b strings.Builder
	b.WriteString("---\ntitle: Spring Boot 4 OpenAPI codegen\ntype: concept\n" +
		"status: published\nstack:\n  - java\n  - spring-boot\n  - maven\n" +
		"tags:\n  - codegen\n  - openapi\nverified: 2026-04-14\nfreshness_days: 365\n" +
		"origin: import\nconfidence: high\nsources:\n  - url: https://example.test/a\n" +
		"    kind: official\n---\n\n# Spring Boot 4 OpenAPI codegen\n\n")
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&b, "Paragraph %d referring to [[some-other-note]] and the `ApiDelegate` "+
			"type, plus `src/main/java/Foo.java` for good measure.\n\n", i)
	}
	return []byte(b.String())
}

// BenchmarkParseNote is the frontmatter half: fence split plus YAML. It is separate from
// the link scan because the two are paid by different callers — `forge validate` never
// walks the body, and folding them into one number hid that the scan is the larger cost.
func BenchmarkParseNote(b *testing.B) {
	src := benchNote()
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		yamlSrc, _, err := SplitFrontmatter(src)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := ParseFrontmatter(yamlSrc); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWikilinks is the body half, and the one the graph reports pay per note.
func BenchmarkWikilinks(b *testing.B) {
	_, body, err := SplitFrontmatter(benchNote())
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(body)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Wikilinks(body)
	}
}

// BenchmarkLoadNote is what `forge index` actually pays per note — read, stat, split,
// YAML — so the 200 ms index budget is roughly 91x this number plus the walk.
func BenchmarkLoadNote(b *testing.B) {
	dir := b.TempDir()
	abs := filepath.Join(dir, "note.md")
	if err := os.WriteFile(abs, benchNote(), 0o644); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Load(abs, "notes/concept/note.md"); err != nil {
			b.Fatal(err)
		}
	}
}
