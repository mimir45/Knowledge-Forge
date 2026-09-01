package graph

import (
	"fmt"
	"testing"
)

func TestInboundCounts(t *testing.T) {
	g := Build([]Node{
		{Rel: "concepts/a.md", Outbound: []string{"concepts/b.md", "concepts/c.md"}},
		{Rel: "concepts/b.md", Outbound: []string{"concepts/c.md"}},
		{Rel: "concepts/c.md"},
	})
	want := map[string]int{"concepts/a.md": 0, "concepts/b.md": 1, "concepts/c.md": 2}
	for rel, n := range want {
		if g.Inbound[rel] != n {
			t.Errorf("Inbound[%s] = %d, want %d", rel, g.Inbound[rel], n)
		}
	}
	if g.Outbound["concepts/a.md"] != 2 {
		t.Errorf("Outbound[a] = %d, want 2", g.Outbound["concepts/a.md"])
	}
}

// TestRootDetectionIsNotInboundCount guards against conflating root detection with
// inbound-link count.
// index.md has zero inbound links and must not be reported as an orphan; a buried note
// with zero inbound must be.
func TestRootDetectionIsNotInboundCount(t *testing.T) {
	g := Build([]Node{
		{Rel: "index.md", Outbound: []string{"concepts/a.md"}},
		{Rel: "moc/java.md", Outbound: []string{"concepts/a.md"}},
		{Rel: "concepts/a.md"},
		{Rel: "concepts/lonely.md"},
	})
	cases := []struct {
		rel    string
		orphan bool
	}{
		{"index.md", false},          // root by location and by name
		{"moc/java.md", false},       // root by location
		{"concepts/a.md", false},     // two inbound links
		{"concepts/lonely.md", true}, // zero inbound, not a root
	}
	for _, c := range cases {
		if got := g.IsOrphan(c.rel); got != c.orphan {
			t.Errorf("IsOrphan(%s) = %v, want %v", c.rel, got, c.orphan)
		}
	}
}

// TestMutualLinkPairIsStillClassified: the fixture's index.md and log.md link to each
// other, which a count-based rule would call non-orphans by accident. Both must be roots
// on structural grounds instead — the classification has to hold for the right reason.
func TestMutualLinkPairIsStillClassified(t *testing.T) {
	g := Build([]Node{
		{Rel: "index.md", Outbound: []string{"log.md"}},
		{Rel: "log.md", Outbound: []string{"index.md"}},
	})
	for _, rel := range []string{"index.md", "log.md"} {
		if !g.IsRoot(rel) {
			t.Errorf("IsRoot(%s) = false; expected a structural root", rel)
		}
	}
}

// TestHubFanoutIsARoot: a note linking to a tenth of the vault is an entry point even
// buried in a subdirectory with no inbound links.
func TestHubFanoutIsARoot(t *testing.T) {
	nodes := []Node{{Rel: "syntheses/hub.md"}}
	for i := 0; i < 60; i++ {
		rel := fmt.Sprintf("concepts/n%02d.md", i)
		nodes[0].Outbound = append(nodes[0].Outbound, rel)
		nodes = append(nodes, Node{Rel: rel})
	}
	g := Build(nodes)
	if !g.IsRoot("syntheses/hub.md") {
		t.Error("a 60-outbound note in a subdirectory was not classified as a hub root")
	}
	if g.IsOrphan("syntheses/hub.md") {
		t.Error("the hub was reported as an orphan")
	}
}

func TestOrphansPreservesInputOrder(t *testing.T) {
	g := Build([]Node{
		{Rel: "concepts/z.md"},
		{Rel: "concepts/a.md"},
		{Rel: "index.md"},
	})
	got := g.Orphans([]Node{{Rel: "concepts/z.md"}, {Rel: "concepts/a.md"}, {Rel: "index.md"}})
	if len(got) != 2 || got[0] != "concepts/z.md" || got[1] != "concepts/a.md" {
		t.Errorf("Orphans = %v, want [concepts/z.md concepts/a.md]", got)
	}
}

// TestUnresolvedTargetIsNeverReportedAsANote: the fixture carries a [[does-not-exist]]
// link. Even if a caller hands Build an unresolved target, every report Build produces
// iterates the node list, so a phantom can never appear in vault output.
func TestUnresolvedTargetIsNeverReportedAsANote(t *testing.T) {
	nodes := []Node{{Rel: "concepts/a.md", Outbound: []string{"does-not-exist.md"}}}
	g := Build(nodes)
	for _, rel := range g.Orphans(nodes) {
		if rel == "does-not-exist.md" {
			t.Error("an unresolved target was reported as an orphaned note")
		}
	}
	if g.Outbound["does-not-exist.md"] != 0 {
		t.Error("an unresolved target was given outbound state")
	}
}
