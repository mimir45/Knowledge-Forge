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

	"knowledge-forge/pkg/recall"
)

// calibrationVault is the corpus recall-spec.md §3.1 is measured against.
//
// It is examples/vault, not the live vault at ~/Documents/Base, for a reason B-008's
// sizing pass found the hard way: §3.1's original numbers were measured on a 91-note
// corpus that has drifted since, so the "before" column could never be re-derived and an
// "after" compared against it would prove nothing. A git-tracked corpus makes both
// columns products of the same run, and makes the table reproducible by anyone.
const calibrationVault = "../../examples/vault"

// calibrationDocs pins the corpus size. A vault edit that changes it must fail here
// rather than silently re-base the golden.
const calibrationDocs = 92

// calibrationNow fixes the clock. Decide branches on Candidate.Stale above the answer
// threshold (ANSWER_FROM_VAULT vs UPDATE(refresh)), so a wall-clock run would flip a
// verdict as the corpus ages and the golden would rot on a calendar, not on a change.
var calibrationNow = time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)

// calibrationQueries are recall-spec.md §3.1's nine adjacent-topic queries, verbatim and
// in the order the table prints them. Adjacent-topic means a closely related note exists
// and extending it is the right move — the band where the verdict is hardest.
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

// TestCalibration is the harness §3.1 never had: until B-008's 2026-08-21 sizing pass,
// the nine queries existed only as prose and the table was hand-transcribed, which is
// why it went stale unnoticed. Run with -update to re-record.
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

// calibrationTable renders one row per query unconditionally. Iterating results instead
// would drop the empty ones: nonZero (rank.go:40-47) truncates a candidate list at the
// first zero score, so a query whose topic left the corpus returns [] and would vanish
// from the table silently. An empty row is the record of a corpus difference.
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

// calibrationRow scores one query. The Top-1 slug column is the one §3.1's original table
// lacked: it recorded what the winner scored but never which note won, so a fix that
// changed the winner while holding the score would have read as a no-op.
//
// The Neighbours column is B-033's. The verdict says a note gets created; this column
// says whether it gets created linked or orphaned, which is the whole of what the
// neighbour floor controls. A floor change must move this column and nothing else —
// score and verdict are functions of Rank and Decide, which the floor does not enter.
func calibrationRow(docs []recall.Doc, question string) string {
	q := recall.Query{Question: question}
	res := recall.DefaultThresholds.Result(q, recall.Rank(q, docs, calibrationNow))
	slug, score := "(none)", "—"
	if len(res.Candidates) > 0 {
		slug = res.Candidates[0].Slug
		score = fmt.Sprintf("%.3f", res.TopScore)
	}
	return fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
		question, slug, score, res.Verdict, neighbourCell(res.Neighbours))
}

// neighbourCell renders the neighbour set as "n: slug, slug". The slugs travel with the
// count because a floor that emits three links is only defensible if they are the right
// three; a bare count would hide a floor that admits noise.
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

// calibrationCorpus stages the vault in a temp dir and loads it. Copying is not
// fastidiousness: loadDocs opens a SQLite cache under <root>/.forge and writes re-parsed
// rows back, so scoring examples/vault in place would mutate a git-tracked directory and
// make the golden depend on whether the cache happened to be warm.
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

// stageVault copies a vault tree into a temp dir. It is fixtureCopy generalised over its
// source; fixtureCopy stays as it is because e2e_test.go's callers name testdata/vault
// through it and the B-002 warning in its doc comment is about that fixture specifically.
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
