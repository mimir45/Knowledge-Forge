package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"knowledge-forge/pkg/recall"
)

// This is B-038 question (b)'s Phase 4: does an alternative tie-break cost any new file
// open, and how does its sort cost compare at a corpus larger than the real vault?
//
// TestDocOpenCountIdenticalAcrossStrategies checks the load-bearing claim directly
// (a counter, not a read of the code): Recency/Verified read fields already populated by
// frontmatterScore, and DocFreqTieBreak builds its specificity map once, at construction
// — before the returned closure runs — so reusing one constructed TieBreak across a sort
// opens exactly the files PathTieBreak would.
func TestDocOpenCountIdenticalAcrossStrategies(t *testing.T) {
	docs := calibrationCorpus(t)
	strategies := tieBreakPrecisionStrategies(docs)
	q := recall.Query{Question: calibrationQueries[0]}
	var base int
	for i, s := range strategies {
		wrapped, counts := countingDocs(docs)
		recall.RankPoolWithTieBreak(q, wrapped, calibrationNow, recall.BodyPassSize, s.tb)
		total := sumCounts(counts)
		if i == 0 {
			base = total
			continue
		}
		if total != base {
			t.Errorf("%s opened %d docs, want %d (same as %s)", s.name, total, base, strategies[0].name)
		}
	}
}

// countingDocs wraps each doc's LoadBody with a per-path counter, everything else
// unchanged. A fresh n/orig per iteration (both declared inside the loop body) means the
// closures never alias each other.
func countingDocs(docs []recall.Doc) ([]recall.Doc, map[string]*int) {
	counts := map[string]*int{}
	out := make([]recall.Doc, len(docs))
	for i, d := range docs {
		n, orig := new(int), d.LoadBody
		counts[d.Rel] = n
		d.LoadBody = func() []byte {
			*n++
			if orig == nil {
				return nil
			}
			return orig()
		}
		out[i] = d
	}
	return out, counts
}

func sumCounts(counts map[string]*int) int {
	total := 0
	for _, n := range counts {
		total += *n
	}
	return total
}

// BenchmarkTieBreakStrategies compares sort cost at a corpus larger than the real vault
// (92 docs). Cloning is valid for *timing* only — NOT for re-measuring saturation rate or
// F1: it multiplies every tag/stack's document frequency by the clone factor uniformly,
// shifting every idf() value uniformly rather than reflecting a real larger vault's own
// tag distribution. See BACKLOG.md's B-038 closing section for the measurements that
// matter; this benchmark is cost only.
func BenchmarkTieBreakStrategies(b *testing.B) {
	docs := scaledBenchCorpus(b, 10)
	q := recall.Query{Question: calibrationQueries[0]}
	for _, s := range tieBreakPrecisionStrategies(docs) {
		b.Run(s.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				recall.RankPoolWithTieBreak(q, docs, calibrationNow, recall.BodyPassSize, s.tb)
			}
		})
	}
}

// scaledBenchCorpus clones examples/vault's docs `factor` times with suffixed
// Rel/Slug, staged into a temp dir the same way calibrationCorpus stages the original
// (loadDocs writes a SQLite cache under <root>/.forge).
func scaledBenchCorpus(b *testing.B, factor int) []recall.Doc {
	b.Helper()
	dst := filepath.Join(b.TempDir(), "vault")
	if err := copyTree(calibrationVault, dst); err != nil {
		b.Fatalf("staging %s: %v", calibrationVault, err)
	}
	docs, err := loadDocs(dst)
	if err != nil {
		b.Fatalf("loading %s: %v", dst, err)
	}
	return cloneDocs(docs, factor)
}

// copyTree is stageVault's copy loop, duplicated rather than reused because stageVault
// takes *testing.T and this file only has a *testing.B.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		return copyFile(p, filepath.Join(dst, rel))
	})
}

// cloneDocs repeats the corpus `factor` times with a suffixed Rel/Slug per copy, so a
// larger N exercises the same sort/comparator cost without changing per-doc content.
func cloneDocs(docs []recall.Doc, factor int) []recall.Doc {
	out := make([]recall.Doc, 0, len(docs)*factor)
	for c := 0; c < factor; c++ {
		suffix := "-" + string(rune('a'+c))
		for _, d := range docs {
			clone := d
			clone.Rel, clone.Slug = d.Rel+suffix, d.Slug+suffix
			out = append(out, clone)
		}
	}
	return out
}
