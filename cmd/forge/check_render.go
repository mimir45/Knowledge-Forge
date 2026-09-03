package main

import (
	"path/filepath"

	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/report"
)

func (d *checkData) coverage() ([]byte, error) {
	return report.RenderCoverage(report.CoverageInput{
		Entries: d.entries, Vocabulary: d.values("stack"), Types: d.values("type"),
		Now: d.now,
	}), nil
}

// values reads an enum out of references/schema.yaml. coverage.md cannot name an
// absence without it.
func (d *checkData) values(field string) []string {
	if d.schema == nil {
		return nil
	}
	f, ok := d.schema.Fields[field]
	if !ok {
		return nil
	}
	return f.Values
}

// staleness reads d.askCounts, loaded from .forge/log.jsonl by loadAskLog.
func (d *checkData) staleness() ([]byte, error) {
	return report.RenderStaleness(report.StalenessInput{
		Entries: d.entries, Asks: d.askCounts, Now: d.now,
	}), nil
}

func (d *checkData) duplicates() ([]byte, error) {
	return report.RenderDuplicates(report.DuplicatesInput{
		Pairs: d.pairs, Threshold: d.cfg.duplicateThreshold(), Compared: d.candidates,
		Slugs: d.slugs, Types: d.types, Now: d.now,
	}), nil
}

// orphans counts against the graph's population, not the contract's.
func (d *checkData) orphans() ([]byte, error) {
	return report.RenderOrphans(report.OrphansInput{
		Orphans: d.graph.Orphans(d.nodes), Total: len(d.nodes),
		Slugs: d.slugs, Types: d.types, Now: d.now,
	}), nil
}

// gaps reads d.askList, loaded from .forge/log.jsonl by loadAskLog.
func (d *checkData) gaps() ([]byte, error) {
	return report.RenderGaps(report.GapsInput{Asks: d.askList, Now: d.now}), nil
}

func (d *checkData) graphHealth() ([]byte, error) {
	return report.RenderGraphHealth(report.GraphHealthInput{
		Components: d.comps, Hubs: d.hubs(), Total: len(d.nodes),
		Slugs: d.slugs, Now: d.now,
	}), nil
}

func (d *checkData) hubs() []string {
	var out []string
	for _, n := range d.nodes {
		if d.graph.IsRoot(n.Rel) {
			out = append(out, n.Rel)
		}
	}
	return out
}

func (d *checkData) churn() ([]byte, error) {
	if d.churnErr != nil {
		return nil, d.churnErr
	}
	return report.RenderChurn(report.ChurnInput{
		Stats: d.churnStats, Slugs: d.slugs, Months: d.cfg.months, Now: d.now,
	}), nil
}

func (d *checkData) deadlinks() ([]byte, error) {
	return report.RenderDeadlinks(report.DeadlinksInput{
		Citations: d.citations, FirstParty: d.firstParty, Slugs: d.slugs, Now: d.now,
	}), nil
}

func (d *checkData) drift() ([]byte, error) {
	if d.repoErr != nil {
		return nil, d.repoErr
	}
	return report.RenderDrift(report.DriftInput{
		Findings: d.findings, Slugs: d.slugs, Now: d.now,
	}), nil
}

// cost surfaces d.budgetErr the same way churn and drift surface theirs.
func (d *checkData) cost() ([]byte, error) {
	if d.budgetErr != nil {
		return nil, d.budgetErr
	}
	return report.RenderCost(d.budget), nil
}

// codebase concatenates one rendered section per repository.
func (d *checkData) codebase() ([]byte, error) {
	if d.repoErr != nil {
		return nil, d.repoErr
	}
	if d.codeErr != nil {
		return nil, d.codeErr
	}
	var out []byte
	for i, in := range d.code {
		if i > 0 {
			out = append(out, "\n---\n\n"...)
		}
		out = append(out, report.RenderCodebase(in)...)
	}
	return out, nil
}

// weekly is the one report in this file that is not a pure function of d.
func (d *checkData) weekly() ([]byte, error) {
	store := report.OpenWeeklyStore(filepath.Join(d.root, ".forge"))
	key := report.WeekKey(d.now)
	prev := store.Prev(key)
	stats := d.vaultStats()

	md := report.RenderWeekly(d.weeklyInput(stats, prev))
	store.Record(key, stats)
	store.Prune(12) // .forge/ is not a place we want unbounded weekly history
	return md, store.Save()
}

func (d *checkData) vaultStats() report.VaultStats {
	return report.VaultStats{
		Notes:   len(d.notes),
		HitRate: report.HitRate(d.askList),
		Orphans: len(d.graph.Orphans(d.nodes)),
		Drift:   report.AffectedByDrift(d.findings, drift.Broken, drift.Suspect),
	}
}

func (d *checkData) weeklyInput(stats report.VaultStats, prev *report.VaultStats) report.WeeklyInput {
	week, year := d.now.ISOWeek()
	return report.WeeklyInput{
		Week: week, Year: year,
		Broken: d.findings, Uncovered: d.allUncovered(), UncoveredDays: d.cfg.days,
		DuplicatePairs: d.pairs, DeadCitations: d.citations,
		StaleCount: report.CountOverdue(d.entries, d.now), MergeCandidates: len(d.pairs),
		OrphanCount: stats.Orphans, Stats: stats, Prev: prev,
		Asks: d.askList, Slugs: d.slugs, Now: d.now,
	}
}

// allUncovered flattens moc/codebase.md's per-repo sections into one ranked pool — Act
// now ranks across repos, unlike codebase.md which ranks within each one.
func (d *checkData) allUncovered() []report.Uncovered {
	var out []report.Uncovered
	for _, c := range d.code {
		out = append(out, c.Uncovered...)
	}
	return out
}
