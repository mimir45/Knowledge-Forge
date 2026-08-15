// Package drift answers one question per note: does the code this note cites still say
// what the note says it says?
//
// Two design rules from ADDENDUM §B.6 shape everything here.
//
// Git-anchored. Drift runs from post-commit / post-merge / post-checkout hooks against
// `--since-commit <sha>`, never on file save and never against the uncommitted working
// tree. `git diff --name-only <sha>..HEAD` is the cheap gate; symbol comparison touches
// only the files in that set, and the vault is never re-scanned.
//
// Verdicts are a pure function of (note refs, tree state). Nothing here consults a
// stored hash from a previous run, which is why a revert restores a demoted note
// automatically: at the restored HEAD the symbol is back, so the verdict is OK again,
// with no undo log to replay. The only thing .forge/ persists is the confidence value a
// note held *before* demotion — a restore target, never a verdict input.
package drift

import (
	"strings"

	"knowledge-forge/pkg/codeindex"
	"knowledge-forge/pkg/coderef"
)

// Verdict is the outcome for one citation.
type Verdict string

const (
	OK       Verdict = "ok"
	Repaired Verdict = "repaired" // the line moved but the symbol is intact
	Suspect  Verdict = "suspect"  // the code is still there and no longer says the same thing
	Broken   Verdict = "broken"   // the file or the symbol is gone
	Skipped  Verdict = "skipped"  // nothing on this machine can adjudicate it
)

// Finding is one citation's verdict, with enough context for drift.md to be actionable
// without the reader opening the note.
type Finding struct {
	Note    string  `json:"note"`
	Ref     string  `json:"ref"` // as the note wrote it
	Repo    string  `json:"repo,omitempty"`
	Path    string  `json:"path,omitempty"` // repo-relative, resolved
	Symbol  string  `json:"symbol,omitempty"`
	Verdict Verdict `json:"verdict"`
	Reason  string  `json:"reason"`
	WasLine int     `json:"was_line,omitempty"`
	NowLine int     `json:"now_line,omitempty"`
}

// Demoting reports whether this verdict costs the note its confidence. Only BROKEN
// does: SUSPECT flags and pins, it does not demote, because "the body changed" is a
// prompt to re-read, not evidence the note is wrong.
func (f Finding) Demoting() bool { return f.Verdict == Broken }

// Source is what drift needs from a repository. Keeping it an interface is what lets
// this package stay pure Go while pkg/codeindex carries the cgo: the concrete
// implementation is wired in cmd/forge, and drift itself never imports a C library.
type Source interface {
	// At returns the symbol table for one file at one revision. ok is false when the
	// file does not exist at that revision.
	At(repo, path, rev string) (codeindex.File, bool)
	// RevBefore returns the last commit on the current branch at or before a date in
	// YYYY-MM-DD form, or "" when the repository has no such commit.
	RevBefore(repo, date string) string
	// Head returns the revision drift is evaluating.
	Head(repo string) string
	// Find locates a declaration by name across every registered repository. asOf is ""
	// for HEAD, or a YYYY-MM-DD date, in which case each repository is searched at its
	// last commit on or before it. A citation that names a class and no file — the shape
	// most of this vault's references take — has nothing else to resolve through.
	Find(name, asOf string) (repo, path string, sym codeindex.Symbol, ok bool)
	// ResolveAt resolves a path-shaped citation against one repository's file list as it
	// stood at asOf ("" for HEAD, else a YYYY-MM-DD date). Mirrors Find's asOf contract.
	ResolveAt(ref coderef.Ref, asOf string) coderef.Resolution
}

// Opts carries the one behavioural switch, which exists because the two callers have
// budgets two orders of magnitude apart. `forge drift` runs on the git-hook path with
// 100ms and checks symbol-only citations against HEAD alone. `forge check` runs weekly
// with 10s and can afford to re-index a repository at the note's verified-era revision
// to prove a missing symbol was once present — the difference between "this class is
// gone" and "this class is java.util.ArrayList".
type Opts struct{ Deep bool }

// Note is the vault side of the input: a note's citations plus the date its claims were
// last verified, which is the baseline SUSPECT is measured against.
type Note struct {
	Rel      string
	Verified string // YYYY-MM-DD; may be empty
	Refs     []coderef.Ref
}

// Check evaluates every citation of every note. Changed is the output of the cheap gate
// — the repo-relative paths touched between the anchor sha and HEAD. A nil Changed
// means "evaluate everything", which is what `forge check` does weekly and what the
// hook path must never do.
func Check(notes []Note, rg *coderef.Registry, src Source, changed map[string]bool,
	opts Opts) []Finding {

	var out []Finding
	for _, n := range notes {
		for _, ref := range n.Refs {
			if f, ok := checkRef(n, ref, rg, src, changed, opts); ok {
				out = append(out, f)
			}
		}
	}
	return out
}

func checkRef(n Note, ref coderef.Ref, rg *coderef.Registry, src Source,
	changed map[string]bool, opts Opts) (Finding, bool) {

	f := Finding{Note: n.Rel, Ref: ref.Raw, Symbol: ref.Symbol, WasLine: ref.Line}
	if ref.Kind == coderef.KindSymbol {
		return checkSymbolOnly(f, n, ref, src, opts), true
	}
	res := rg.Resolve(ref)
	switch res.Status {
	case coderef.Unresolved:
		return unresolvedPath(f, n, ref, src, changed, opts), true
	case coderef.Ambiguous:
		return skip(f, "matches "+strings.Join(res.Ambiguity, ", ")), true
	}
	f.Repo, f.Path = res.Ref.Repo, res.RepoPath
	if changed != nil && !changed[f.Path] {
		return Finding{}, false // outside the gate: the file did not move, so nothing can have
	}
	return checkPath(f, n, ref, src), true
}

func skip(f Finding, reason string) Finding {
	f.Verdict, f.Reason = Skipped, reason
	return f
}
