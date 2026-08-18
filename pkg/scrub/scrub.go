// Package scrub redacts personal and secret content from an Obsidian vault before it is
// published — Phase 6's `examples/vault/` and, later, Phase 6b's `--anonymize` both
// build on this. It never mutates its input: Scrub reads srcDir and writes a redacted
// copy to dstDir, and it fails closed — any note it cannot scrub and re-validate aborts
// the whole run before anything is written.
package scrub

import (
	"fmt"

	"knowledge-forge/pkg/vault"
)

// Report summarizes one Scrub run.
type Report struct {
	NotesTotal    int
	NotesWritten  int
	Redactions    int
	NoFrontmatter []string // vault-relative paths written body-only (no FM to scrub)
}

// file is one note's scrubbed output, held in memory until every note in the run has
// succeeded — that is what makes a failure leave dstDir untouched.
type file struct {
	rel  string
	data []byte
}

// Scrub walks srcDir as a vault, redacts every note, and writes the result under dstDir
// with the same relative layout. On any note's failure it returns an error and writes
// nothing at all.
func Scrub(srcDir, dstDir string) (Report, error) {
	schema, err := vault.LoadSchema()
	if err != nil {
		return Report{}, fmt.Errorf("scrub: load schema: %w", err)
	}
	rels, err := vault.Walk(srcDir)
	if err != nil {
		return Report{}, fmt.Errorf("scrub: walk %s: %w", srcDir, err)
	}
	files, rep, err := scrubAll(srcDir, rels, schema)
	if err != nil {
		return Report{}, err
	}
	if err := writeAll(dstDir, files); err != nil {
		return Report{}, err
	}
	return rep, nil
}

// scrubAll processes every note into memory. It stops at the first failure — a caller
// gets either a complete, consistent set of outputs or none at all.
func scrubAll(srcDir string, rels []string, schema *vault.Schema) ([]file, Report, error) {
	var files []file
	var rep Report
	for _, rel := range rels {
		data, noFM, n, err := scrubOne(srcDir, rel, schema)
		if err != nil {
			return nil, Report{}, fmt.Errorf("scrub: %s: %w", rel, err)
		}
		files = append(files, file{rel, data})
		rep.NotesTotal++
		rep.Redactions += n
		if noFM {
			rep.NoFrontmatter = append(rep.NoFrontmatter, rel)
		}
	}
	rep.NotesWritten = len(files)
	return files, rep, nil
}
