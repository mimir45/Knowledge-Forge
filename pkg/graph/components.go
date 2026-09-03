package graph

import "sort"

// Component is one connected group of notes.
type Component struct {
	Members []string // sorted, so a report generated twice is byte-identical
	Roots   []string // the entry points inside it, sorted; empty means no way in
}

// Size reports how many notes the component holds.
func (c Component) Size() int { return len(c.Members) }

// Components partitions the vault into connected groups, treating links as undirected.
func Components(nodes []Node) []Component {
	uf := newUnionFind(nodes)
	for _, n := range nodes {
		for _, t := range n.Outbound {
			uf.union(n.Rel, t)
		}
	}
	return uf.collect(nodes)
}

// LargestShare returns the fraction of notes living in the biggest component. It is the one
// number that summarises the partition: 0.9 is a wiki, 0.2 is a folder.
func LargestShare(comps []Component, total int) float64 {
	if total == 0 || len(comps) == 0 {
		return 0
	}
	best := 0
	for _, c := range comps {
		best = max(best, c.Size())
	}
	return float64(best) / float64(total)
}

// unionFind is the standard disjoint-set, keyed by vault-relative path.
type unionFind struct {
	parent map[string]string
}

func newUnionFind(nodes []Node) *unionFind {
	uf := &unionFind{parent: make(map[string]string, len(nodes))}
	for _, n := range nodes {
		uf.parent[n.Rel] = n.Rel
	}
	return uf
}

// find resolves a path to its set representative, flattening as it goes.
func (u *unionFind) find(x string) (string, bool) {
	root, ok := u.parent[x]
	if !ok {
		return "", false
	}
	for root != u.parent[root] {
		u.parent[root] = u.parent[u.parent[root]]
		root = u.parent[root]
	}
	u.parent[x] = root
	return root, true
}

func (u *unionFind) union(a, b string) {
	ra, okA := u.find(a)
	rb, okB := u.find(b)
	if !okA || !okB || ra == rb {
		return
	}
	// Smaller path wins so the representative does not depend on insertion order.
	if rb < ra {
		ra, rb = rb, ra
	}
	u.parent[rb] = ra
}

// collect groups members by representative and sorts everything: components by size
// descending then by first member, and members within a component by path.
func (u *unionFind) collect(nodes []Node) []Component {
	byRoot := map[string][]string{}
	for _, n := range nodes {
		if r, ok := u.find(n.Rel); ok {
			byRoot[r] = append(byRoot[r], n.Rel)
		}
	}
	out := make([]Component, 0, len(byRoot))
	for _, members := range byRoot {
		sort.Strings(members)
		out = append(out, Component{Members: members})
	}
	sortComponents(out)
	return out
}

func sortComponents(cs []Component) {
	sort.Slice(cs, func(i, j int) bool {
		if len(cs[i].Members) != len(cs[j].Members) {
			return len(cs[i].Members) > len(cs[j].Members)
		}
		return cs[i].Members[0] < cs[j].Members[0]
	})
}

// WithRoots fills in each component's entry points.
func (g *Graph) WithRoots(comps []Component) []Component {
	out := make([]Component, len(comps))
	for i, c := range comps {
		out[i] = Component{Members: c.Members}
		for _, m := range c.Members {
			if g.roots[m] {
				out[i].Roots = append(out[i].Roots, m)
			}
		}
	}
	return out
}
