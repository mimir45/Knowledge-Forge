package main

import (
	"sort"
	"strings"
	"time"

	"knowledge-forge/pkg/drift"
	"knowledge-forge/pkg/gitsig"
	"knowledge-forge/pkg/graph"
	"knowledge-forge/pkg/report"
	"knowledge-forge/pkg/similarity"
	"knowledge-forge/pkg/vault"
)

// checkData is the vault collected once, in the two shapes the reports need.
//
// notes and entries are deliberately different sets and must not be interchanged.
// notes are content notes — everything that is a node in the link graph, moc/ included,
// because a map of content genuinely de-orphans what it points at. entries are contract
// notes — the subset the schema judges, moc/ and the hubs excluded, because `type:` has
// no meaning for a map and counting one would move the validity denominator. Feed notes
// to the graph reports and entries to the coverage and staleness ones; the crisp check
// after a run is that the two counts differ by exactly the maps and hubs.
type checkData struct {
	cfg  checkCfg
	root string
	now  time.Time

	notes   []*vault.Note
	entries []report.Entry
	slugs   map[string]string
	types   map[string]string
	schema  *vault.Schema

	nodes []graph.Node
	graph *graph.Graph
	comps []graph.Component

	pairs      []similarity.Pair
	candidates int

	churnStats *gitsig.Stats
	churnErr   error

	citations  []report.Citation
	firstParty int
	findings   []drift.Finding
	code       []report.CodebaseInput

	// repoErr fails drift.md and the codebase map together — both read the same registry.
	// codeErr fails only the map, which is the cgo-only one.
	repoErr error
	codeErr error
}

func collectVault(cfg checkCfg, root string) (*checkData, error) {
	notes, err := loadNotes(root)
	if err != nil {
		return nil, err
	}
	d := &checkData{cfg: cfg, root: root, now: time.Now(), notes: notes}
	d.schema, _ = vault.LoadSchema()
	d.slugs, d.types = slugMap(notes), typeMap(notes)
	d.buildGraph()
	d.similar()
	d.churnStats, d.churnErr = vaultHistory(root, cfg.months, d.now)
	d.links()
	d.driftAndCode()
	return d, nil
}

func (d *checkData) buildGraph() {
	rels := make([]string, 0, len(d.notes))
	for _, n := range d.notes {
		rels = append(rels, n.Rel)
	}
	ix := vault.NewIndex(rels)
	d.nodes = nodesOf(ix, d.notes)
	d.graph = graph.Build(d.nodes)
	d.comps = d.graph.WithRoots(graph.Components(d.nodes))
	d.entries = entries(d.notes, d.graph)
}

func slugMap(notes []*vault.Note) map[string]string {
	out := make(map[string]string, len(notes))
	for _, n := range notes {
		out[n.Rel] = slugOf(n)
	}
	return out
}

func typeMap(notes []*vault.Note) map[string]string {
	out := make(map[string]string, len(notes))
	for _, n := range notes {
		out[n.Rel] = typeOf(n.Rel)
	}
	return out
}

// typeOf reads the note type off the directory rather than the frontmatter.
//
// The two disagree in this vault — 31 of 91 notes fail the contract and several fail on
// `type:` itself — and the directory is what the migration actually enforced. Duplicate
// scoping in particular has to agree with how the notes are grouped on disk, or a pair
// gets compared against a group it is not in.
func typeOf(rel string) string {
	rest, ok := strings.CutPrefix(rel, "notes/")
	if !ok {
		if i := strings.Index(rel, "/"); i > 0 {
			return rel[:i]
		}
		return ""
	}
	if i := strings.Index(rest, "/"); i > 0 {
		return rest[:i]
	}
	return ""
}

// similar indexes bodies only. Frontmatter is boilerplate by construction — the same
// eleven keys in the same order in every note — and including it would score every pair
// of well-formed notes as partly identical.
func (d *checkData) similar() {
	ix := similarity.NewIndex()
	for _, n := range d.notes {
		ix.Add(n.Rel, typeOf(n.Rel), string(n.Body))
	}
	d.candidates = ix.Candidates()
	d.pairs = ix.Pairs(d.cfg.duplicateThreshold())
}

// duplicateThreshold resolves the config value against the package constant. The
// fallback is not decoration: collectVault is called from tests with a bare checkCfg,
// and a zero threshold would report every pair in the vault as a duplicate.
func (c checkCfg) duplicateThreshold() float64 {
	if c.dupThreshold > 0 {
		return c.dupThreshold
	}
	return similarity.DuplicateThreshold
}

// vaultHistory reads the *vault's* commits, not a code repository's. ADDENDUM section
// B.4's churn.md asks which notes keep being rewritten; code churn is moc/codebase.md's
// subject and the two must not be conflated.
func vaultHistory(root string, months int, now time.Time) (*gitsig.Stats, error) {
	var since time.Time
	if months > 0 {
		since = now.AddDate(0, -months, 0)
	}
	commits, err := gitsig.Log(root, since)
	if err != nil {
		return nil, err
	}
	return gitsig.Analyze(commits), nil
}

func sortedStrings(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
