package main

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/config"
	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/engine"
	"github.com/mimir45/Knowledge-Forge/pkg/gitsig"
	"github.com/mimir45/Knowledge-Forge/pkg/graph"
	"github.com/mimir45/Knowledge-Forge/pkg/report"
	"github.com/mimir45/Knowledge-Forge/pkg/similarity"
	"github.com/mimir45/Knowledge-Forge/pkg/store"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
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

	askCounts map[string]int
	askList   []report.Ask

	// repoErr fails drift.md and the codebase map together — both read the same registry.
	// codeErr fails only the map, which is the cgo-only one.
	repoErr error
	codeErr error

	budget    report.CostInput
	budgetErr error
}

func collectVault(cfg checkCfg, root string) (*checkData, error) {
	notes, err := loadNotes(root)
	if err != nil {
		return nil, err
	}
	d := &checkData{cfg: cfg, root: root, now: time.Now(), notes: notes}
	d.schema, _ = vault.LoadSchema()
	d.slugs, d.types = slugMap(notes), typeMap(notes)
	d.askCounts, d.askList = loadAskLog(filepath.Join(root, ".forge", "log.jsonl"), d.slugs)
	d.buildGraph()
	d.similar()
	d.churnStats, d.churnErr = vaultHistory(root, cfg.months, d.now)
	d.links()
	d.driftAndCode()
	d.budgetErr = d.budgetSnapshot()
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

// budgetSnapshot fills d.budget from cfg.Engines.Budget, the SQLite budget table, and
// cfg.Pipeline — cost.md's whole job is showing what the config chain and today's spend
// jointly decide, so this is the one collector that reads config and cache directly.
// A bare checkCfg (as the collectVault tests build) has no config to read; that degrades
// to an empty report rather than a nil-pointer panic, the same posture duplicateThreshold
// takes toward a zero threshold.
func (d *checkData) budgetSnapshot() error {
	d.budget = report.CostInput{Now: d.now, StageEngine: map[string]string{}}
	if d.cfg.config == nil {
		return nil
	}
	cfg := d.cfg.config
	d.budget.OnExhausted = cfg.Engines.Budget.OnExhausted
	d.budget.CapPerDay = map[string]float64{
		"api": cfg.Engines.Budget.APIUSDPerDay, "advisor": cfg.Engines.Budget.AdvisorUSDPerDay,
	}
	st, err := store.Open(d.root)
	if err != nil {
		return err
	}
	defer st.Close()
	if d.budget.SpentToday, err = spentToday(st, d.budget.CapPerDay, time.Now); err != nil {
		return err
	}
	d.budget.StageEngine = stageEngines(cfg, st, time.Now)
	d.budget.QueuedNotes = countQueued(d.notes)
	return nil
}

// spentToday recovers spend from Remaining rather than a second SQL query: Remaining
// already computes cap-minus-spent, so cap-minus-Remaining is spent, for any cap
// including the unmetered zero value.
func spentToday(l engine.Ledger, caps map[string]float64, clock func() time.Time) (map[string]float64, error) {
	out := make(map[string]float64, len(caps))
	for tier, cap := range caps {
		remaining, err := l.Remaining(tier, cap, clock)
		if err != nil {
			return nil, err
		}
		out[tier] = cap - remaining
	}
	return out, nil
}

// stageEngines is cost.md's "what would today actually run" section — the same Resolve
// call forge engine select makes, once per pipeline stage.
func stageEngines(cfg *config.Config, l engine.Ledger, clock func() time.Time) map[string]string {
	out := make(map[string]string, len(cfg.Pipeline))
	for stage := range cfg.Pipeline {
		if name, _, err := engine.Resolve(cfg, l, clock, stage); err == nil {
			out[stage] = name
		}
	}
	return out
}

// countQueued reads the same pending_advisor flag queueNote (engine_run.go) writes.
func countQueued(notes []*vault.Note) int {
	n := 0
	for _, note := range notes {
		if isQueued(note) {
			n++
		}
	}
	return n
}

// isQueued is countQueued's per-note predicate, shared with check_drain.go's dispatch
// loop so the two never drift apart on what "queued" means.
func isQueued(n *vault.Note) bool {
	return n.FM != nil && strings.EqualFold(n.FM.Str("pending_advisor"), "true")
}
