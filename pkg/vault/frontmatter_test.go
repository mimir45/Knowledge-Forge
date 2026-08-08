package vault

import (
	"testing"
)

func TestSplitFrontmatter(t *testing.T) {
	cases := []struct {
		name, src, wantYAML, wantBody string
		wantErr                       error
	}{
		{"normal", "---\ntitle: x\n---\n# H\n", "title: x", "# H\n", nil},
		{"crlf", "---\r\ntitle: x\r\n---\r\n# H\r\n", "title: x", "# H\n", nil},
		{"empty body", "---\ntitle: x\n---\n", "title: x", "", nil},
		{"no block", "# H\n\nbody\n", "", "# H\n\nbody\n", ErrNoFrontmatter},
		{"unterminated", "---\ntitle: x\n", "", "---\ntitle: x\n", ErrNoFrontmatter},
		// A horizontal rule is not a fence: it has no opening delimiter on line 1.
		{"hr only", "text\n\n---\n\nmore\n", "", "text\n\n---\n\nmore\n", ErrNoFrontmatter},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			y, body, err := SplitFrontmatter([]byte(c.src))
			if err != c.wantErr {
				t.Fatalf("err = %v, want %v", err, c.wantErr)
			}
			if string(y) != c.wantYAML {
				t.Errorf("yaml = %q, want %q", y, c.wantYAML)
			}
			if string(body) != c.wantBody {
				t.Errorf("body = %q, want %q", body, c.wantBody)
			}
		})
	}
}

// TestSplitFrontmatterBodyIsByteIdentical is the property the migration depends on:
// body content is never reordered or rewritten, only frontmatter is.
func TestSplitFrontmatterBodyIsByteIdentical(t *testing.T) {
	body := "# Heading\n\n```yaml\n---\nnot: frontmatter\n---\n```\n\n- a\n-  b\n\n\n"
	_, got, err := SplitFrontmatter([]byte("---\ntitle: x\n---\n" + body))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Errorf("body mutated:\n got %q\nwant %q", got, body)
	}
}

func TestParseFrontmatterKeepsOrder(t *testing.T) {
	fm, err := ParseFrontmatter([]byte("type: concept\ntitle: x\nslug: y\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"type", "title", "slug"}
	for i, k := range want {
		if fm.Keys[i] != k {
			t.Fatalf("Keys = %v, want %v", fm.Keys, want)
		}
	}
}

func TestFrontmatterAccessors(t *testing.T) {
	src := "title: Kafka\ntags:\n  - java\n  - jpa\nstack: docker\nempty:\nrelated: []\n"
	fm, err := ParseFrontmatter([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if fm.Str("title") != "Kafka" {
		t.Errorf("Str(title) = %q", fm.Str("title"))
	}
	if got := fm.List("tags"); len(got) != 2 || got[0] != "java" || got[1] != "jpa" {
		t.Errorf("List(tags) = %v", got)
	}
	// A bare scalar `stack:` is how several real vault notes write a single value.
	if got := fm.List("stack"); len(got) != 1 || got[0] != "docker" {
		t.Errorf("List(stack) = %v, want [docker]", got)
	}
	if got := fm.List("empty"); got != nil {
		t.Errorf("List(empty) = %v, want nil", got)
	}
	if got := fm.List("related"); len(got) != 0 {
		t.Errorf("List(related) = %v, want empty", got)
	}
	if !fm.Has("empty") {
		t.Error("Has(empty) = false; a null-valued key was still written")
	}
	if fm.Has("absent") {
		t.Error("Has(absent) = true")
	}
}

func TestParseFrontmatterNonMapping(t *testing.T) {
	if _, err := ParseFrontmatter([]byte("- a\n- b\n")); err == nil {
		t.Error("a YAML sequence was accepted as frontmatter")
	}
}
