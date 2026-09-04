package drift

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
	"github.com/mimir45/Knowledge-Forge/pkg/coderef"
)

// checkPath walks the original spec's ladder for a citation that named a file.
func checkPath(f Finding, n Note, ref coderef.Ref, src Source) Finding {
	head := src.Head(f.Repo)
	now, ok := src.At(f.Repo, f.Path, head)
	if !ok {
		f.Verdict, f.Reason = Broken, "file no longer exists at HEAD"
		return f
	}
	if ref.Symbol != "" {
		return checkSymbol(f, n, ref.Symbol, now, src)
	}
	return checkFileBody(f, n, now, src)
}

// checkSymbol is the case the addendum cares most about.
func checkSymbol(f Finding, n Note, symbol string, now codeindex.File, src Source) Finding {
	s, ok := now.Lookup(symbol)
	if !ok {
		f.Verdict = Broken
		f.Reason = fmt.Sprintf("symbol %s is no longer declared in this file", symbol)
		return f
	}
	return symbolVerdict(f, n, symbol, s, src)
}

// symbolVerdict is the tail of the ladder shared by both citation shapes.
func symbolVerdict(f Finding, n Note, symbol string, s codeindex.Symbol, src Source) Finding {
	f.NowLine = s.Start
	if changed, base := bodyChanged(f, n, symbol, s.BodyHash, src); changed {
		f.Verdict = Suspect
		f.Reason = fmt.Sprintf("body of %s changed since the note was verified (%s)", symbol, base)
		return f
	}
	if f.WasLine != 0 && f.WasLine != s.Start {
		f.Verdict = Repaired
		f.Reason = fmt.Sprintf("%s moved from line %d to %d", symbol, f.WasLine, s.Start)
		return f
	}
	f.Verdict, f.Reason = OK, "symbol present and unchanged"
	return f
}

// checkSymbolOnly handles a citation that named a class with no path.
func checkSymbolOnly(f Finding, n Note, ref coderef.Ref, src Source, opts Opts) Finding {
	if repo, path, s, ok := src.Find(ref.Symbol, ""); ok {
		f.Repo, f.Path = repo, path
		return symbolVerdict(f, n, ref.Symbol, s, src)
	}
	return absentSymbol(f, n, ref, src, opts)
}

// absentSymbol decides what "no repository declares this name" means.
func absentSymbol(f Finding, n Note, ref coderef.Ref, src Source, opts Opts) Finding {
	if !opts.Deep || n.Verified == "" {
		return skip(f, "no indexed repository declares "+ref.Symbol)
	}
	repo, path, _, was := src.Find(ref.Symbol, n.Verified)
	if !was {
		return skip(f, ref.Symbol+" was not declared here at "+n.Verified+" either")
	}
	f.Repo, f.Path = repo, path
	f.Verdict = Broken
	f.Reason = fmt.Sprintf("%s was declared at %s and is gone at HEAD", ref.Symbol, n.Verified)
	return f
}

// unresolvedPath decides what "no registered repository contains this path" means on a
// true full sweep.
func unresolvedPath(f Finding, n Note, ref coderef.Ref, src Source, opts Opts) Finding {
	if !opts.Deep || n.Verified == "" {
		return skip(f, "no registered repository contains this path")
	}
	res := src.ResolveAt(ref, n.Verified)
	if res.Status != coderef.Resolved {
		return skip(f, "no registered repository contained this path at "+n.Verified+" either")
	}
	f.Repo, f.Path = res.Ref.Repo, res.RepoPath
	f.Verdict = Broken
	f.Reason = fmt.Sprintf("no path matching this citation exists at HEAD; it resolved at %s",
		n.Verified)
	return f
}

// deletedInGate reports whether ref's basename matches a path the cheap gate reported
// deleted in the commit range this run is checking.
func deletedInGate(ref coderef.Ref, deleted map[string]string) (repo, path string, ok bool) {
	base := lastSegment(ref.Path)
	if base == "" {
		return "", "", false
	}
	var matches []string
	for p := range deleted {
		if strings.EqualFold(lastSegment(p), base) {
			matches = append(matches, p)
		}
	}
	if len(matches) == 0 {
		return "", "", false
	}
	sort.Strings(matches)
	return deleted[matches[0]], matches[0], true
}

func lastSegment(p string) string {
	seg := coderef.Segments(p)
	if len(seg) == 0 {
		return ""
	}
	return seg[len(seg)-1]
}

// checkFileBody is the weakest case: the note named a file and nothing inside it. There
// is no declaration to anchor to, so the verdict is about the file as a whole.
func checkFileBody(f Finding, n Note, now codeindex.File, src Source) Finding {
	base, ok := baseline(f, n, src)
	if !ok {
		f.Verdict, f.Reason = OK, "file present; no verified baseline to compare against"
		return f
	}
	if fileHash(base) != fileHash(now) {
		f.Verdict = Suspect
		f.Reason = "declarations in this file changed since the note was verified"
		return f
	}
	f.Verdict, f.Reason = OK, "file present and its declarations are unchanged"
	return f
}

// bodyChanged compares one symbol's body at HEAD against its body in the tree as it
// stood when the note was last verified.
func bodyChanged(f Finding, n Note, symbol, nowHash string, src Source) (bool, string) {
	base, ok := baseline(f, n, src)
	if !ok {
		return false, ""
	}
	was, found := base.Lookup(symbol)
	if !found {
		return false, "" // newly introduced since the note; not drift
	}
	return was.BodyHash != nowHash, n.Verified
}

func baseline(f Finding, n Note, src Source) (codeindex.File, bool) {
	if n.Verified == "" {
		return codeindex.File{}, false
	}
	rev := src.RevBefore(f.Repo, n.Verified)
	if rev == "" {
		return codeindex.File{}, false
	}
	return src.At(f.Repo, f.Path, rev)
}

// fileHash folds a file's declarations into one value so "did anything in here change"
// is a single comparison.
func fileHash(f codeindex.File) string {
	h := ""
	for _, s := range f.Symbols {
		h += s.Name + ":" + s.BodyHash + ";"
	}
	return h
}
