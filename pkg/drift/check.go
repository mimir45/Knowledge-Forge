package drift

import (
	"fmt"

	"knowledge-forge/pkg/codeindex"
	"knowledge-forge/pkg/coderef"
)

// checkPath walks ADDENDUM §B.6's ladder for a citation that named a file, in the order
// the addendum states it: file gone, then symbol gone, then line moved, then body
// changed. The order matters — a deleted file would otherwise report as a moved line.
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

// checkSymbol is the case the addendum cares most about: a citation that names a
// declaration. A rename or a removal is BROKEN and demotes the note; a line number that
// moved while the symbol stayed put is repaired silently.
func checkSymbol(f Finding, n Note, symbol string, now codeindex.File, src Source) Finding {
	s, ok := now.Lookup(symbol)
	if !ok {
		f.Verdict = Broken
		f.Reason = fmt.Sprintf("symbol %s is no longer declared in this file", symbol)
		return f
	}
	return symbolVerdict(f, n, symbol, s, src)
}

// symbolVerdict is the tail of the ladder shared by both citation shapes: the symbol is
// present, so the only questions left are whether its body moved on and whether the line
// number the note recorded still points at it.
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

// checkSymbolOnly handles a citation that named a class with no path — the shape most of
// the vault's references take. The symbol table answers where it lives, so a note citing
// `OrderConsumer` gets a real verdict rather than a shrug.
func checkSymbolOnly(f Finding, n Note, ref coderef.Ref, src Source, opts Opts) Finding {
	if repo, path, s, ok := src.Find(ref.Symbol, ""); ok {
		f.Repo, f.Path = repo, path
		return symbolVerdict(f, n, ref.Symbol, s, src)
	}
	return absentSymbol(f, n, ref, src, opts)
}

// absentSymbol decides what "no repository declares this name" means, and the honest
// answer depends on the budget. On the hook path a name we cannot find is more often a
// library type than a deletion, so it is SKIPPED. Only the weekly pass pays to look at
// the verified-era tree, and only a name that *was* declared there is BROKEN.
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
// stood when the note was last verified. Both sides come from git, so the comparison is
// a pure function of tree state and a revert flips it back without an undo log.
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
// is a single comparison. Names and body hashes only: reordering declarations is a
// change, moving them down the file is not.
func fileHash(f codeindex.File) string {
	h := ""
	for _, s := range f.Symbols {
		h += s.Name + ":" + s.BodyHash + ";"
	}
	return h
}
