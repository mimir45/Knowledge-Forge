package report

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/graph"
)

// OrphansInput is what orphans.md renders from.
type OrphansInput struct {
	Orphans []string // vault-relative paths, from graph.Orphans
	Total   int      // notes in the graph, for the share
	Slugs   map[string]string
	Types   map[string]string
	Now     time.Time
}

// RenderOrphans produces orphans.md — notes nothing links to.
//
// An orphan is not a bad note. It is a note the vault cannot surface: it will never appear
// in a backlink pane, never be reached by following a thread, and so exists only for
// whoever remembers its name. The fix is a link from somewhere, not a rewrite.
func RenderOrphans(in OrphansInput) []byte {
	var b strings.Builder
	header(&b, "Orphans", orphansSummary(in), in.Now)
	if len(in.Orphans) == 0 {
		empty(&b, "every note is reachable")
		return []byte(b.String())
	}
	groups := byType(in.Orphans, in.Types)
	for _, t := range slices.Sorted(maps.Keys(groups)) {
		fmt.Fprintf(&b, "\n## %s\n\n", t)
		for _, rel := range groups[t] {
			fmt.Fprintf(&b, "- %s\n", note(in.Slugs[rel], rel))
		}
	}
	return []byte(b.String())
}

func orphansSummary(in OrphansInput) string {
	var share float64
	if in.Total > 0 {
		share = float64(len(in.Orphans)) / float64(in.Total) * 100
	}
	return fmt.Sprintf("**%d of %d notes (%.0f%%) have no inbound links.**\n\n"+
		"Roots are excluded: a map of content or a vault front door is an entry point, "+
		"not an orphan.\n", len(in.Orphans), in.Total, share)
}

func byType(rels []string, types map[string]string) map[string][]string {
	out := map[string][]string{}
	for _, rel := range rels {
		t := types[rel]
		if t == "" {
			t = "untyped"
		}
		out[t] = append(out[t], rel)
	}
	return out
}

// GraphHealthInput is what graph-health.md renders from.
type GraphHealthInput struct {
	Components []graph.Component // sorted largest first, roots filled in
	Hubs       []string          // notes classified as roots
	Total      int
	Slugs      map[string]string
	Now        time.Time
}

// RenderGraphHealth produces graph-health.md — is this one wiki or fifteen notebooks?
//
// The number that answers it is the largest component's share. One big component with a
// tail of pairs is a wiki; fifteen components of equal size is fifteen unrelated notebooks
// that happen to share a folder, and no amount of per-note quality fixes that.
func RenderGraphHealth(in GraphHealthInput) []byte {
	var b strings.Builder
	header(&b, "Graph health", graphSummary(in), in.Now)
	writeIslands(&b, in)
	writeHubs(&b, in)
	return []byte(b.String())
}

func graphSummary(in GraphHealthInput) string {
	share := graph.LargestShare(in.Components, in.Total) * 100
	return fmt.Sprintf("**%d %s over %d notes**; the largest holds %.0f%% of them.\n",
		len(in.Components), plural(len(in.Components), "component", "components"),
		in.Total, share)
}

// writeIslands names the components with no root. Those are the real finding: none of
// their members is an orphan — they link to each other — and there is still no way to
// reach them from the vault's front door.
func writeIslands(b *strings.Builder, in GraphHealthInput) {
	islands := rootless(in.Components)
	fmt.Fprintf(b, "\n## Unreachable clusters — %d\n", len(islands))
	if len(islands) == 0 {
		empty(b, "every cluster has a way in")
		return
	}
	b.WriteString("\nNotes here link to each other, so none is an orphan, and nothing " +
		"links in from outside.\n\n")
	for _, c := range head(islands, 15) {
		fmt.Fprintf(b, "- **%d notes** — %s\n", c.Size(), joinNotes(c.Members, in.Slugs, 4))
	}
}

func rootless(cs []graph.Component) []graph.Component {
	var out []graph.Component
	for _, c := range cs {
		if len(c.Roots) == 0 && c.Size() > 1 {
			out = append(out, c)
		}
	}
	return out
}

func writeHubs(b *strings.Builder, in GraphHealthInput) {
	fmt.Fprintf(b, "\n## Entry points — %d\n\n", len(in.Hubs))
	if len(in.Hubs) == 0 {
		b.WriteString("_none; every note is a leaf, so there is no way to browse this vault_\n")
		return
	}
	for _, rel := range head(in.Hubs, 20) {
		fmt.Fprintf(b, "- %s\n", note(in.Slugs[rel], rel))
	}
}

func joinNotes(rels []string, slugs map[string]string, n int) string {
	var parts []string
	for _, rel := range head(rels, n) {
		parts = append(parts, note(slugs[rel], rel))
	}
	if len(rels) > n {
		parts = append(parts, fmt.Sprintf("_+%d more_", len(rels)-n))
	}
	return strings.Join(parts, " · ")
}
