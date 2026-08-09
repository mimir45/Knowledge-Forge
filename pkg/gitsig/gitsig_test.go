package gitsig

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A tiny repository with a known shape: a.go and b.go always change together (3 commits),
// c.go churns alone, and one commit is wide enough to be capped out of coupling.
func repo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	git(t, dir, "config", "user.email", "ada@example.com")
	git(t, dir, "config", "user.name", "Ada")
	commit(t, dir, "Ada", "a.go", "b.go")
	commit(t, dir, "Ada", "a.go", "b.go")
	commit(t, dir, "Grace", "a.go", "b.go")
	commit(t, dir, "Grace", "c.go")
	commit(t, dir, "Grace", wide(CouplingCap+1)...)
	return dir
}

func wide(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "gen/f" + string(rune('a'+i)) + ".go"
	}
	return out
}

func commit(t *testing.T, dir, author string, files ...string) {
	t.Helper()
	for _, f := range files {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(time.Now().String()+f), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git(t, dir, "add", "-A")
	git(t, dir, "-c", "user.name="+author, "-c", "user.email="+author+"@example.com",
		"commit", "-q", "-m", "touch "+strings.Join(files, " "))
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestLogReadsEveryCommitAndItsFiles(t *testing.T) {
	commits, err := Log(repo(t), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 5 {
		t.Fatalf("commits = %d, want 5", len(commits))
	}
	if got := commits[0].Files; len(got) != CouplingCap+1 { // git log is newest-first
		t.Errorf("newest commit touched %d files, want %d", len(got), CouplingCap+1)
	}
	if commits[0].SHA == "" || commits[0].When.IsZero() {
		t.Errorf("commit = %+v, want a sha and a date", commits[0])
	}
}

func TestChurnAndCoupling(t *testing.T) {
	commits, err := Log(repo(t), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	s := Analyze(commits)
	for _, c := range []struct {
		file string
		want int
	}{{"a.go", 3}, {"b.go", 3}, {"c.go", 1}} {
		if got := s.Churn[c.file]; got != c.want {
			t.Errorf("churn[%s] = %d, want %d", c.file, got, c.want)
		}
	}
	if got := s.Coupled[[2]string{"a.go", "b.go"}]; got != 3 {
		t.Errorf("coupling a.go/b.go = %d, want 3", got)
	}
}

// A commit wide enough to be a reformat or a dependency bump must not dominate coupling: it
// alone would contribute more pairs than the entire real history.
func TestWideCommitIsCappedOutOfCoupling(t *testing.T) {
	s := Analyze(mustLog(t, repo(t)))
	for p := range s.Coupled {
		if strings.HasPrefix(p[0], "gen/") {
			t.Fatalf("pair %v came from the capped commit", p)
		}
	}
	if s.Churn["gen/fa.go"] != 1 { // but it still counts as churn, where its size is honest
		t.Errorf("capped commit did not count toward churn")
	}
}

func TestOwnerIsTheAuthorOfMostCommits(t *testing.T) {
	s := Analyze(mustLog(t, repo(t)))
	owner, share := s.Owner("a.go")
	if owner != "Ada" || share < 0.66 || share > 0.67 {
		t.Errorf("owner(a.go) = %q %.2f, want Ada with 2 of 3", owner, share)
	}
	if o, sh := s.Owner("nothing.go"); o != "" || sh != 0 {
		t.Errorf("owner of an unknown file = %q %.2f, want empty", o, sh)
	}
}

// Reports are committed to the vault, so identical history must render identical bytes —
// no map iteration order may leak into the ranking.
func TestRankingIsDeterministic(t *testing.T) {
	s := Analyze(mustLog(t, repo(t)))
	first, firstC := TopChurn(s, 5), TopCoupled(s, 2, 5)
	for i := 0; i < 20; i++ {
		if got := TopChurn(s, 5); got[0] != first[0] || got[len(got)-1] != first[len(first)-1] {
			t.Fatalf("churn ranking changed between runs: %+v vs %+v", got, first)
		}
		if got := TopCoupled(s, 2, 5); len(got) != len(firstC) || got[0] != firstC[0] {
			t.Fatalf("coupling ranking changed between runs: %+v vs %+v", got, firstC)
		}
	}
}

// One shared commit is a coincidence. The minimum exists so churn.md does not list every
// pair of files that ever appeared in the same commit.
func TestCouplingMinimumFiltersCoincidences(t *testing.T) {
	s := Analyze(mustLog(t, repo(t)))
	if got := TopCoupled(s, 4, 0); len(got) != 0 {
		t.Errorf("pairs at min=4 = %+v, want none", got)
	}
	if got := TopCoupled(s, 3, 0); len(got) != 1 || got[0].A != "a.go" {
		t.Errorf("pairs at min=3 = %+v, want just a.go/b.go", got)
	}
}

func mustLog(t *testing.T, dir string) []Commit {
	t.Helper()
	c, err := Log(dir, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func BenchmarkAnalyze(b *testing.B) {
	commits := make([]Commit, 2000)
	for i := range commits {
		commits[i] = Commit{Author: "Ada", Files: []string{"a.go", "b.go", "c.go"}}
	}
	b.ReportAllocs()
	for b.Loop() {
		Analyze(commits)
	}
}
