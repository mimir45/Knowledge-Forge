package dataset

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// repo is a throwaway git repository with controllable commit dates. Everything D3 decides
// is a function of (commit, tree), so the dates have to be inputs rather than wall clock.
type repo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) *repo {
	t.Helper()
	r := &repo{t: t, dir: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.email", "test@example.com")
	r.git("config", "user.name", "Test")
	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()
	return r.gitAt("2026-01-01T00:00:00Z", args...)
}

func (r *repo) gitAt(when string, args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", r.dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE="+when, "GIT_COMMITTER_DATE="+when,
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func (r *repo) write(rel, body string) {
	r.t.Helper()
	abs := filepath.Join(r.dir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

func (r *repo) commit(when, msg string, paths ...string) {
	r.t.Helper()
	r.gitAt(when, append([]string{"add", "-A"}, paths...)...)
	r.gitAt(when, "commit", "-q", "-m", msg)
}

func note(origin, title, body string) string {
	return "---\ntitle: " + title + "\ntype: concept\norigin: " + origin +
		"\nstack:\n  - go\n---\n\n" + body + "\n"
}

const (
	day0 = "2026-03-01T12:00:00Z"
	day1 = "2026-03-02T12:00:00Z"
	day9 = "2026-03-10T12:00:00Z"
)

// seed lays down one forge-generated note and one imported one, then returns the repo.
func seed(t *testing.T) *repo {
	r := newRepo(t)
	r.write("notes/concept/goroutines.md", note("ask", "Goroutines", "generated text"))
	r.write("notes/concept/legacy.md", note("import", "Legacy", "migrated text"))
	r.commit(day0, "seed")
	return r
}

func capture(t *testing.T, r *repo) []Pair {
	t.Helper()
	pairs, err := Capture(Git{Dir: r.dir}, Options{Commit: "HEAD"})
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	return pairs
}

// TestCapturesHumanEditOfAGeneratedNote is the whole point of D3: a note forge wrote,
// corrected by a human the next day, becomes one preference pair.
func TestCapturesHumanEditOfAGeneratedNote(t *testing.T) {
	r := seed(t)
	r.write("notes/concept/goroutines.md", note("ask", "Goroutines", "corrected text"))
	r.commit(day1, "fix the goroutines note")

	pairs := capture(t, r)
	if len(pairs) != 1 {
		t.Fatalf("got %d pairs, want 1: %+v", len(pairs), pairs)
	}
	p := pairs[0]
	if !strings.Contains(p.Generated, "generated text") {
		t.Errorf("generated side is wrong: %q", p.Generated)
	}
	if !strings.Contains(p.Preferred, "corrected text") {
		t.Errorf("preferred side is wrong: %q", p.Preferred)
	}
	if p.Topic != "Goroutines" || p.Origin != "ask" || p.Kind != D3Kind {
		t.Errorf("metadata wrong: %+v", p)
	}
}

// TestSkipsImportedNotes: the Phase 1 migration stamped origin: import on all 91
// pre-existing notes.
func TestSkipsImportedNotes(t *testing.T) {
	r := seed(t)
	r.write("notes/concept/legacy.md", note("import", "Legacy", "hand-edited text"))
	r.commit(day1, "edit a migrated note")

	if pairs := capture(t, r); len(pairs) != 0 {
		t.Errorf("got %d pairs from an imported note, want 0: %+v", len(pairs), pairs)
	}
}

// TestSkipsEditsOutsideTheWindow: an edit nine days on is a revisit, not a correction of
// what the model produced. The window is set at seven days.
func TestSkipsEditsOutsideTheWindow(t *testing.T) {
	r := seed(t)
	r.write("notes/concept/goroutines.md", note("ask", "Goroutines", "much later text"))
	r.commit(day9, "revisit")

	if pairs := capture(t, r); len(pairs) != 0 {
		t.Errorf("got %d pairs outside the window, want 0: %+v", len(pairs), pairs)
	}
}

// TestSkipsEditsThatPredateGeneration: clock skew on a synced vault, or grafted
// history, can date the edit commit before the add commit.
func TestSkipsEditsThatPredateGeneration(t *testing.T) {
	r := newRepo(t)
	r.write("notes/concept/t.md", note("ask", "T", "generated text"))
	r.commit(day9, "seed with a future clock")
	r.write("notes/concept/t.md", note("ask", "T", "corrected text"))
	r.commit(day1, "edit, dated earlier than the note was born")

	if pairs := capture(t, r); len(pairs) != 0 {
		t.Errorf("got %d pairs from a backwards clock, want 0: %+v", len(pairs), pairs)
	}
}

// TestFollowsRenameAndEditInOneCommit: the realistic human move is a single commit that
// retitles the note, renames the file to match, and rewrites the body.
func TestFollowsRenameAndEditInOneCommit(t *testing.T) {
	r := seed(t)
	r.git("mv", "notes/concept/goroutines.md", "notes/concept/goroutine-basics.md")
	r.write("notes/concept/goroutine-basics.md", note("ask", "Goroutines", "corrected text"))
	r.commit(day1, "retitle and correct")

	pairs := capture(t, r)
	if len(pairs) != 1 {
		t.Fatalf("got %d pairs from a rename+edit commit, want 1: %+v", len(pairs), pairs)
	}
	if !strings.Contains(pairs[0].Generated, "generated text") {
		t.Errorf("generated side lost: %q", pairs[0].Generated)
	}
}

// TestSkipsTheCommitThatCreatedTheNote: generation and edit cannot be the same commit,
// or every new note would pair with itself.
func TestSkipsTheCommitThatCreatedTheNote(t *testing.T) {
	r := seed(t)
	if pairs := capture(t, r); len(pairs) != 0 {
		t.Errorf("got %d pairs from the creating commit, want 0", len(pairs))
	}
}

// TestSkipsForgeAuthoredCommits guards the Phase 4 failure mode: once forge-librarian
// commits notes it wrote, those commits would otherwise enter D3 as.
func TestSkipsForgeAuthoredCommits(t *testing.T) {
	r := seed(t)
	r.write("notes/concept/goroutines.md", note("ask", "Goroutines", "rewritten by forge"))
	r.commit(day1, "regenerate\n\n"+ForgeTrailer+": true")

	if pairs := capture(t, r); len(pairs) != 0 {
		t.Errorf("got %d pairs from a forge commit, want 0: %+v", len(pairs), pairs)
	}
}

// TestIgnoresNonNotePaths: raw/ and sources/ are inputs, _index.md is generated. An edit
// to any of them is not a correction of a note.
func TestIgnoresNonNotePaths(t *testing.T) {
	r := seed(t)
	r.write("raw/daily/2026-03-01.md", note("ask", "Daily", "v1"))
	r.commit(day0, "add a raw capture")
	r.write("raw/daily/2026-03-01.md", note("ask", "Daily", "v2"))
	r.commit(day1, "edit the raw capture")

	if pairs := capture(t, r); len(pairs) != 0 {
		t.Errorf("got %d pairs from raw/, want 0: %+v", len(pairs), pairs)
	}
}

// TestFollowsRenames: --follow is what keeps the generated side readable after a note
// moves. Without it the pre-rename history is invisible and the note looks newly born.
func TestFollowsRenames(t *testing.T) {
	r := seed(t)
	r.git("mv", "notes/concept/goroutines.md", "notes/concept/goroutine-basics.md")
	r.commit(day1, "rename")
	r.write("notes/concept/goroutine-basics.md", note("ask", "Goroutines", "corrected text"))
	r.commit(day1, "correct after the rename")

	pairs := capture(t, r)
	if len(pairs) != 1 {
		t.Fatalf("got %d pairs across a rename, want 1: %+v", len(pairs), pairs)
	}
	if !strings.Contains(pairs[0].Generated, "generated text") {
		t.Errorf("generated side lost across the rename: %q", pairs[0].Generated)
	}
}

// TestWindowIsGitAnchoredNotFrontmatter: `created:` is a mutable field that --fix
// backfills.
func TestWindowIsGitAnchoredNotFrontmatter(t *testing.T) {
	r := newRepo(t)
	stale := "---\ntitle: T\ntype: concept\norigin: ask\ncreated: 2020-01-01\n---\n\nv1\n"
	r.write("notes/concept/t.md", stale)
	r.commit(day0, "seed")
	r.write("notes/concept/t.md", strings.Replace(stale, "v1", "v2", 1))
	r.commit(day1, "edit")

	if pairs := capture(t, r); len(pairs) != 1 {
		t.Errorf("got %d pairs, want 1: a stale created: must not shrink the window", len(pairs))
	}
}

// TestAppendIsIdempotent: the hook can fire twice on one commit — amend, rebase, a manual
// rerun — and the dataset is append-only, so the second run must add nothing.
func TestAppendIsIdempotent(t *testing.T) {
	r := seed(t)
	r.write("notes/concept/goroutines.md", note("ask", "Goroutines", "corrected text"))
	r.commit(day1, "fix")
	path := filepath.Join(t.TempDir(), D3Path)

	first, err := Append(path, capture(t, r))
	if err != nil || first != 1 {
		t.Fatalf("first append = %d, %v; want 1, nil", first, err)
	}
	second, err := Append(path, capture(t, r))
	if err != nil || second != 0 {
		t.Fatalf("second append = %d, %v; want 0, nil", second, err)
	}
	if lines := countLines(t, path); lines != 1 {
		t.Errorf("dataset has %d lines after two runs, want 1", lines)
	}
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return len(strings.Split(strings.TrimSpace(string(b)), "\n"))
}

// TestNoPairsOnTodaysVault is the honest statement of what this hook does right now:
// every note in the real vault carries origin: import.
func TestNoPairsOnTodaysVault(t *testing.T) {
	r := newRepo(t)
	for _, name := range []string{"a", "b", "c"} {
		r.write("notes/concept/"+name+".md", note("import", name, "migrated"))
	}
	r.commit(day0, "Phase 1: migrate the vault")
	r.write("notes/concept/a.md", note("import", "a", "edited by hand"))
	r.commit(day1, "tidy")

	if pairs := capture(t, r); len(pairs) != 0 {
		t.Errorf("got %d pairs, want 0 on an all-import vault: %+v", len(pairs), pairs)
	}
}

func TestWindowDefaultsToSevenDays(t *testing.T) {
	if got := window(Options{}); got != 7*24*time.Hour {
		t.Errorf("default window = %v, want 168h", got)
	}
}
