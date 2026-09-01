// Package graph analyses the note link graph.
package graph

import (
	"path"
	"strings"
)

// Node is one note's position in the graph.
type Node struct {
	Rel      string
	Outbound []string // resolved vault-relative paths
}

// Graph holds inbound counts and root classification for a vault.
type Graph struct {
	Inbound  map[string]int
	Outbound map[string]int
	roots    map[string]bool
}

// hubFraction: a note linking to at least this share of the vault is a hub, and a hub
// with no inbound links is an entry point rather than an orphan.
const hubFraction = 0.10

// Build computes the graph. Roots are classified here, not by the caller, because
// getting it wrong — treating a legitimate root like index.md as an orphan just because
// it has zero inbound links — is a specific, easy-to-reintroduce failure.
func Build(nodes []Node) *Graph {
	g := &Graph{
		Inbound:  make(map[string]int, len(nodes)),
		Outbound: make(map[string]int, len(nodes)),
		roots:    map[string]bool{},
	}
	for _, n := range nodes {
		g.Inbound[n.Rel] += 0
		g.Outbound[n.Rel] = len(n.Outbound)
		for _, t := range n.Outbound {
			g.Inbound[t]++
		}
	}
	g.classifyRoots(nodes)
	return g
}

// classifyRoots marks the notes that are entry points into the graph.
//
// Root detection deliberately does NOT reduce to an inbound-link count. In the fixture
// vault index.md and log.md link to each other, so a count-based rule calls them
// non-orphans by accident; in the real vault index.md has genuinely zero inbound and the
// same rule would report the vault's front door as an orphan. Structure decides instead:
// location, conventional name, or hub-scale fan-out.
func (g *Graph) classifyRoots(nodes []Node) {
	threshold := int(float64(len(nodes)) * hubFraction)
	for _, n := range nodes {
		if isRootLocation(n.Rel) || isRootName(n.Rel) || len(n.Outbound) >= max(threshold, 5) {
			g.roots[n.Rel] = true
		}
	}
}

func isRootLocation(rel string) bool {
	return strings.HasPrefix(rel, "moc/") || !strings.Contains(rel, "/")
}

func isRootName(rel string) bool {
	switch strings.TrimSuffix(strings.ToLower(path.Base(rel)), ".md") {
	case "index", "_index", "log", "readme", "00-index", "home":
		return true
	}
	return false
}

// IsRoot reports whether a note is an entry point.
func (g *Graph) IsRoot(rel string) bool { return g.roots[rel] }

// IsOrphan reports whether a note has no inbound links and is not a root. This is the
// only definition of orphan the rest of the codebase should use.
func (g *Graph) IsOrphan(rel string) bool {
	return g.Inbound[rel] == 0 && !g.roots[rel]
}

// Orphans returns every orphaned note, in the order given to Build.
func (g *Graph) Orphans(nodes []Node) []string {
	var out []string
	for _, n := range nodes {
		if g.IsOrphan(n.Rel) {
			out = append(out, n.Rel)
		}
	}
	return out
}
