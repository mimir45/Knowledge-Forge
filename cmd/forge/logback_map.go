package main

import (
	"fmt"
	"path/filepath"

	"knowledge-forge/pkg/coderef"
	"knowledge-forge/pkg/drift"
	"knowledge-forge/pkg/gitsig"
	"knowledge-forge/pkg/report"
	"knowledge-forge/pkg/vault"
)

// buildGroups is check_codebase.go's codebases/oneCodebase pipeline minus churn and
// owners: logback has no --days window and nothing it renders uses either, so this
// loads the vault once and skips the gitsig.Log a churn report would need.
// gitsig.Analyze(nil) yields a valid, empty Stats — groupsOf reads it for Churn/Owner
// data that stays zero/"" throughout, which is fine since RenderKnowledgeMap and the
// CLAUDE.md fragments never look at either field.
func buildGroups(vaultRoot string, r drift.Repo, rg *coderef.Registry, src symbolFinder) ([]report.CodeGroup, error) {
	notes, err := loadNotes(vaultRoot)
	if err != nil {
		return nil, err
	}
	scanned, err := coderef.ScanRepo(r.Name, r.Root, "HEAD")
	if err != nil {
		return nil, err
	}
	cited := citedPathsFree(notes, slugsOf(notes), rg, src)
	ix := src.Index(r.Name, "HEAD")
	return groupsOf(scanned.Files, gitsig.Analyze(nil), cited[r.Name], dependsOn(ix, scanned.Files)), nil
}

// citedPathsFree is check_codebase.go's citedPaths without the *checkData receiver — the
// same join, over an explicit notes/slugs pair instead of the weekly pass's cached
// fields, so forge logback does not need to build a checkData just to reuse this logic.
func citedPathsFree(notes []*vault.Note, slugs map[string]string, rg *coderef.Registry,
	src symbolFinder) map[string]map[string][]string {

	out := map[string]map[string][]string{}
	for _, n := range driftNotes(notes) {
		for _, ref := range n.Refs {
			if repo, p, ok := locate(ref, rg, src); ok {
				if out[repo] == nil {
					out[repo] = map[string][]string{}
				}
				out[repo][p] = append(out[repo][p], slugs[n.Rel])
			}
		}
	}
	return out
}

func slugsOf(notes []*vault.Note) map[string]string {
	out := make(map[string]string, len(notes))
	for _, n := range notes {
		out[n.Rel] = slugOf(n)
	}
	return out
}

// writeKnowledgeMap renders docs/knowledge-map.md into the repo root. --dry-run reports
// what would change without touching disk, following writeReport's own changed/unchanged
// vocabulary so forge logback's output reads like forge check's.
func writeKnowledgeMap(r drift.Repo, groups []report.CodeGroup, dryRun bool) bool {
	path := filepath.Join(r.Root, "docs", "knowledge-map.md")
	md := report.RenderKnowledgeMap(groups)
	if dryRun {
		fmt.Printf("%s: docs/knowledge-map.md would be written (dry run)\n", r.Name)
		return true
	}
	changed, err := writeReport(path, md)
	if err != nil {
		fmt.Printf("%s: docs/knowledge-map.md: %v\n", r.Name, err)
		return false
	}
	fmt.Printf("%s: docs/knowledge-map.md written%s\n", r.Name, unchangedNote(changed))
	return true
}
