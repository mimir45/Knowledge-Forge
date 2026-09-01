package vault

import (
	"reflect"
	"testing"
)

func TestWikilinks(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{"plain", "see [[soft-delete]] and [[hibernate]]",
			[]string{"soft-delete", "hibernate"}},
		{"alias", "see [[soft-delete|the soft delete note]]", []string{"soft-delete"}},
		{"heading", "see [[hibernate#Soft Delete]]", []string{"hibernate"}},
		{"path qualified", "[[issues/testcontainers-docker-socket]]",
			[]string{"issues/testcontainers-docker-socket"}},
		{"with extension", "[[log.md]]", []string{"log.md"}},
		{"fenced block ignored", "```\n[[not-a-link]]\n```\n[[real]]", []string{"real"}},
		{"inline code ignored", "write `[[x]]` like this; then [[real]]", []string{"real"}},
		{"empty target", "[[]]", nil},
		{"none", "no links here", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Wikilinks([]byte(c.body))
			if len(got) == 0 && len(c.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

// TestResolveExtensionAgnostic guards against a resolver
// that blindly appends ".md" and reads the fixture's [[log.md]] as dangling.
func TestResolveExtensionAgnostic(t *testing.T) {
	ix := NewIndex([]string{"log.md", "index.md", "issues/hibernate-column-mismatch.md"})
	cases := []struct {
		link, want string
	}{
		{"log", "log.md"},
		{"log.md", "log.md"},
		{"./log.md", "log.md"},
		{"LOG", "log.md"},
		{"hibernate-column-mismatch", "issues/hibernate-column-mismatch.md"},
		{"issues/hibernate-column-mismatch", "issues/hibernate-column-mismatch.md"},
		{"issues/hibernate-column-mismatch.md", "issues/hibernate-column-mismatch.md"},
	}
	for _, c := range cases {
		got, ok := ix.Resolve(c.link)
		if !ok || got != c.want {
			t.Errorf("Resolve(%q) = %q,%v; want %q,true", c.link, got, ok, c.want)
		}
	}
	if _, ok := ix.Resolve("does-not-exist"); ok {
		t.Error("a missing target resolved")
	}
}

func TestAmbiguousBasename(t *testing.T) {
	ix := NewIndex([]string{
		"TIL/docker/testcontainers-docker-socket.md",
		"issues/testcontainers-docker-socket.md",
		"concepts/soft-delete.md",
	})
	if !ix.Ambiguous("testcontainers-docker-socket") {
		t.Error("the exact-basename collision was not flagged ambiguous")
	}
	if ix.Ambiguous("soft-delete") {
		t.Error("a unique basename was flagged ambiguous")
	}
	// A path-qualified link is never ambiguous — it names its directory.
	got, ok := ix.Resolve("issues/testcontainers-docker-socket")
	if !ok || got != "issues/testcontainers-docker-socket.md" {
		t.Errorf("path-qualified resolve = %q,%v", got, ok)
	}
}
