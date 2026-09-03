package main

import (
	"fmt"
	"path/filepath"

	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
	"github.com/mimir45/Knowledge-Forge/pkg/drift"
	"github.com/mimir45/Knowledge-Forge/pkg/gitsig"
	"github.com/mimir45/Knowledge-Forge/pkg/report"
	"github.com/mimir45/Knowledge-Forge/pkg/vault"
)

// buildGroups is check_codebase.go's codebases/oneCodebase pipeline minus churn and
// owners: logback has no --days window and nothing it renders uses either.
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

// citedPathsFree is check_codebase.go's citedPaths without the *checkData receiver.
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

// writeKnowledgeMap renders docs/knowledge-map.md into the repo root.
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
