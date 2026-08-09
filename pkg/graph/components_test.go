package graph

import (
	"reflect"
	"testing"
)

// Two islands: a triangle reachable from a root, and a pair reachable from nothing. The
// pair is what graph-health.md exists to surface — neither note is an orphan, both have
// inbound links, and there is still no way to reach them.
func islands() []Node {
	return []Node{
		{Rel: "index.md", Outbound: []string{"notes/concept/a.md"}},
		{Rel: "notes/concept/a.md", Outbound: []string{"notes/concept/b.md"}},
		{Rel: "notes/concept/b.md", Outbound: []string{"notes/concept/a.md"}},
		{Rel: "notes/concept/x.md", Outbound: []string{"notes/concept/y.md"}},
		{Rel: "notes/concept/y.md", Outbound: []string{"notes/concept/x.md"}},
	}
}

func TestComponentsPartitionTheVault(t *testing.T) {
	comps := Components(islands())
	if len(comps) != 2 {
		t.Fatalf("components = %d, want 2", len(comps))
	}
	want := []string{"index.md", "notes/concept/a.md", "notes/concept/b.md"}
	if !reflect.DeepEqual(comps[0].Members, want) {
		t.Errorf("largest = %v, want %v", comps[0].Members, want)
	}
	if comps[1].Size() != 2 {
		t.Errorf("second component = %v, want the x/y pair", comps[1].Members)
	}
}

// A link is followed in both directions. b.md links only back to a.md, and it must still
// land in the component a.md's root reaches — a reader follows backlinks.
func TestLinksAreUndirected(t *testing.T) {
	comps := Components([]Node{
		{Rel: "index.md"},
		{Rel: "notes/concept/a.md", Outbound: []string{"index.md"}},
	})
	if len(comps) != 1 {
		t.Errorf("components = %d, want 1 — the link was followed only forward", len(comps))
	}
}

// The cluster with no way in is the finding. It must be distinguishable from the one that
// has a root, and neither of its members is an orphan.
func TestUnreachableClusterHasNoRoots(t *testing.T) {
	nodes := islands()
	comps := Build(nodes).WithRoots(Components(nodes))
	if len(comps[0].Roots) == 0 {
		t.Errorf("largest component has no root, want index.md")
	}
	if len(comps[1].Roots) != 0 {
		t.Errorf("x/y component roots = %v, want none", comps[1].Roots)
	}
	if got := Build(nodes).Orphans(nodes); len(got) != 0 {
		t.Errorf("orphans = %v; the unreachable pair is not an orphan, which is the point", got)
	}
}

// A wikilink to a note that does not exist must not invent a member.
func TestDanglingLinksDoNotCreateMembers(t *testing.T) {
	comps := Components([]Node{{Rel: "a.md", Outbound: []string{"gone.md"}}})
	if len(comps) != 1 || len(comps[0].Members) != 1 {
		t.Errorf("components = %+v, want a.md alone", comps)
	}
}

func TestLargestShare(t *testing.T) {
	if got := LargestShare(Components(islands()), 5); got != 0.6 {
		t.Errorf("share = %v, want 0.6", got)
	}
	if got := LargestShare(nil, 0); got != 0 {
		t.Errorf("share of an empty vault = %v, want 0", got)
	}
}

// Reports are committed, so the partition must not depend on map iteration order.
func TestComponentOrderIsDeterministic(t *testing.T) {
	first := Components(islands())
	for i := 0; i < 20; i++ {
		if !reflect.DeepEqual(Components(islands()), first) {
			t.Fatalf("component order changed between runs")
		}
	}
}
