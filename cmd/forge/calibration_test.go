package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/recall"
)

// calibrationVault is the corpus recall-spec.md §3.1 is measured against. It is
// examples/vault, not a live vault: a moving corpus makes the golden non-reproducible.
const calibrationVault = "../../examples/vault"

// calibrationDocs pins the corpus size. A vault edit that changes it must fail here
// rather than silently re-base the golden.
const calibrationDocs = 91

// calibrationNow fixes the clock.
var calibrationNow = time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)

// calibrationQueries are recall-spec.md §3.1's nine adjacent-topic queries.
var calibrationQueries = []string{
	"Redis caching in Spring Boot",
	"Spring Boot 4 configuration properties binding",
	"Storybook interaction testing with play functions",
	"Java virtual threads with Spring Boot",
	"Keycloak token exchange between clients",
	"React Server Components data fetching",
	"Kafka consumers with Testcontainers",
	"Docker multi-stage build cache optimization",
	"JPA entity graph to avoid N+1",
}

var updateGolden = flag.Bool("update", false,
	"rewrite cmd/forge/testdata/calibration.golden from the current scorer")

const goldenPath = "testdata/calibration.golden"

// TestCalibration is the harness that makes the recall-spec.md calibration table
// generated rather than hand-transcribed: the nine queries used to exist only as prose.
func TestCalibration(t *testing.T) {
	got := calibrationTable(t)
	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", goldenPath)
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("%v — run: go test ./cmd/forge -run TestCalibration -update", err)
	}
	if got != string(want) {
		t.Errorf("calibration table changed\n--- want\n%s\n--- got\n%s", want, got)
	}
}

// calibrationTable renders one row per query unconditionally.
func calibrationTable(t *testing.T) string {
	docs := calibrationCorpus(t)
	var b strings.Builder
	b.WriteString("| Query | Top-1 slug | Top score | Verdict | Neighbours |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, q := range calibrationQueries {
		b.WriteString(calibrationRow(docs, q))
	}
	return b.String()
}

// calibrationRow scores one query.
func calibrationRow(docs []recall.Doc, question string) string {
	q := recall.Query{Question: question}
	res := recall.DefaultThresholds.ResultFrom(q, recall.RankPool(q, docs, calibrationNow))
	slug, score := "(none)", "—"
	if len(res.Candidates) > 0 {
		slug = res.Candidates[0].Slug
		score = fmt.Sprintf("%.3f", res.TopScore)
	}
	return fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
		question, slug, score, res.Verdict, neighbourCell(res.Neighbours))
}

// neighbourCell renders the neighbour set as "n: slug, slug".
func neighbourCell(ns []recall.Candidate) string {
	if len(ns) == 0 {
		return "0"
	}
	slugs := make([]string, len(ns))
	for i, n := range ns {
		slugs[i] = n.Slug
	}
	return fmt.Sprintf("%d: %s", len(ns), strings.Join(slugs, ", "))
}

// calibrationCorpus stages the vault in a temp dir and loads it.
func calibrationCorpus(t *testing.T) []recall.Doc {
	t.Helper()
	docs, err := loadDocs(stageVault(t, calibrationVault))
	if err != nil {
		t.Fatalf("loading %s: %v", calibrationVault, err)
	}
	if len(docs) != calibrationDocs {
		t.Fatalf("corpus is %d docs, want %d — the golden was measured against %d",
			len(docs), calibrationDocs, calibrationDocs)
	}
	return docs
}

// stageVault copies a vault tree into a temp dir.
func stageVault(t *testing.T, src string) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "vault")
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		return copyFile(p, filepath.Join(dst, rel))
	})
	if err != nil {
		t.Fatalf("staging %s: %v", src, err)
	}
	return dst
}
