package main

import (
	"knowledge-forge/pkg/report"
	"knowledge-forge/pkg/similarity"
)

func (d *checkData) coverage() ([]byte, error) {
	return report.RenderCoverage(report.CoverageInput{
		Entries: d.entries, Vocabulary: d.values("stack"), Types: d.values("type"),
		Now: d.now,
	}), nil
}

// values reads an enum out of references/schema.yaml. coverage.md cannot name an absence
// without it: counting the stacks that appear in notes says what has been written about,
// and only subtracting from the full vocabulary says what has not.
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

// staleness passes a nil ask log, which is the truth today: every note in this vault is
// origin: import, so nothing has been asked for and the report says so and falls back to
// days overdue. Phase 4's capture log is what fills this in.
func (d *checkData) staleness() ([]byte, error) {
	return report.RenderStaleness(report.StalenessInput{
		Entries: d.entries, Asks: nil, Now: d.now,
	}), nil
}

func (d *checkData) duplicates() ([]byte, error) {
	return report.RenderDuplicates(report.DuplicatesInput{
		Pairs: d.pairs, Threshold: similarity.DuplicateThreshold, Compared: d.candidates,
		Slugs: d.slugs, Types: d.types, Now: d.now,
	}), nil
}

// orphans counts against the graph's population, not the contract's. An orphan is a note
// nothing links to, and moc/ pages link — excluding them from the denominator would count
// notes the maps do reach.
func (d *checkData) orphans() ([]byte, error) {
	return report.RenderOrphans(report.OrphansInput{
		Orphans: d.graph.Orphans(d.nodes), Total: len(d.nodes),
		Slugs: d.slugs, Types: d.types, Now: d.now,
	}), nil
}

// gaps has no ask log to read, so it reports the honest empty state: not "no gaps" but
// "no data yet". The renderer keeps those two apart; see reports/gaps.md.
func (d *checkData) gaps() ([]byte, error) {
	return report.RenderGaps(report.GapsInput{Asks: nil, Now: d.now}), nil
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

// codebase concatenates one rendered section per repository. RenderCodebase names a single
// repo because a map of two codebases is two maps; joining them here keeps moc/codebase.md
// a single entry point without teaching the renderer about a list it cannot rank across.
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
