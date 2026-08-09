package coderef

import "testing"

// The span classifier is the whole reason the unresolved count means anything: if it
// admits `mvn test` and `spring.datasource.url` as citations, every report downstream
// reads as a vault full of broken code references when it is a vault full of prose.
func TestParseSpanClassification(t *testing.T) {
	cases := []struct {
		in   string
		ok   bool
		kind Kind
		path string
		sym  string
		line int
	}{
		{in: "common-domain/valueobject/Money.java", ok: true, kind: KindPath, path: "common-domain/valueobject/Money.java"},
		{in: "src/app/page.tsx:42", ok: true, kind: KindPath, path: "src/app/page.tsx", line: 42},
		{in: "OrderConsumer.java#L17", ok: true, kind: KindPath, path: "OrderConsumer.java", line: 17},
		{in: "OrderConsumer", ok: true, kind: KindSymbol, sym: "OrderConsumer"},
		{in: "Money.add()", ok: true, kind: KindSymbol, sym: "Money.add"},
		{in: "mvn test", ok: false},
		{in: "spring.datasource.url", ok: false},
		{in: "docs/README.md", ok: false},
		{in: "application.yml", ok: false},
		{in: "id", ok: false},
	}
	for _, c := range cases {
		r, ok := parseSpan(c.in)
		if ok != c.ok {
			t.Errorf("%q: ok=%v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok && (r.Kind != c.kind || r.Path != c.path || r.Symbol != c.sym || r.Line != c.line) {
			t.Errorf("%q: %+v, want kind=%s path=%s sym=%s line=%d",
				c.in, r, c.kind, c.path, c.sym, c.line)
		}
	}
}

func TestFromBodyDedupesAndIgnoresFences(t *testing.T) {
	body := []byte("see `Money.java` and `Money.java` again\n\n" +
		"```java\nclass Illustration { }\n```\n")
	refs := FromBody("notes/concept/a.md", body)
	if len(refs) != 1 || refs[0].Path != "Money.java" || refs[0].Note != "notes/concept/a.md" {
		t.Fatalf("refs = %+v, want one Money.java citing the note", refs)
	}
}

func TestFromFrontmatterCanonicalForm(t *testing.T) {
	refs := FromFrontmatter("n.md", []string{
		"MeterReadingsService:src/main/java/app/ReadingController.java:88#create",
		"leprecoin:src/app/page.tsx",
		"nonsense-without-a-colon",
	})
	if len(refs) != 2 {
		t.Fatalf("refs = %+v, want 2", refs)
	}
	got := refs[0]
	if got.Repo != "MeterReadingsService" || got.Line != 88 || got.Symbol != "create" ||
		got.Path != "src/main/java/app/ReadingController.java" {
		t.Errorf("canonical ref = %+v", got)
	}
}
