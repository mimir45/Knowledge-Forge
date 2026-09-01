package main

import (
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
)

// dependsOn folds codeindex.File.Imports into the directory-level dependency graph
// CodeGroup.DependsOn needs. An import is per file and names a package or a module path,
// neither of which is the "module" this report groups by (groupsOf's module is a
// directory) — resolution and folding happen together here, one directory-shaped answer
// per import, rather than as two separate steps: neither language resolver below can ever
// name an exact file with confidence (Java has no known source root to anchor from; a
// TypeScript import may name a directory's barrel index directly), so "which directory"
// is the only question either can honestly answer.
//
// Third-party imports — an unresolvable dotted Java name, a bare TypeScript specifier that
// would resolve through node_modules — have no matching directory in this repo and are
// silently dropped: never invent a dependency-graph node for them.
func dependsOn(ix codeindex.Index, files []string) map[string][]string {
	dirs, fileSet := knownDirs(files), fileSetOf(files)
	out := map[string]map[string]bool{}
	for p, f := range ix.Files {
		from := path.Dir(p)
		for _, imp := range f.Imports {
			dep, ok := resolveImport(imp, f.Lang, from, files, dirs, fileSet)
			if ok && dep != from {
				addDep(out, from, dep)
			}
		}
	}
	return sortedDeps(out)
}

func addDep(out map[string]map[string]bool, from, dep string) {
	if out[from] == nil {
		out[from] = map[string]bool{}
	}
	out[from][dep] = true
}

// sortedDeps is where "deduped and sorted" (the plan's own words) actually happens: a
// map's own iteration order is not a property of the tree, and moc/codebase.md has to
// render the same "Depends on:" line every run (2b's determinism standard).
func sortedDeps(out map[string]map[string]bool) map[string][]string {
	result := make(map[string][]string, len(out))
	for dir, set := range out {
		result[dir] = slices.Sorted(maps.Keys(set))
	}
	return result
}

func knownDirs(files []string) map[string]bool {
	out := make(map[string]bool, len(files))
	for _, f := range files {
		out[path.Dir(f)] = true
	}
	return out
}

func fileSetOf(files []string) map[string]bool {
	out := make(map[string]bool, len(files))
	for _, f := range files {
		out[f] = true
	}
	return out
}

func resolveImport(imp, lang, from string, files []string, dirs, fileSet map[string]bool) (string, bool) {
	switch lang {
	case "typescript":
		return resolveTSImport(imp, from, fileSet)
	case "java":
		return resolveJavaImport(imp, files, dirs)
	default:
		return "", false
	}
}

var tsExts = []string{".ts", ".tsx", ".js", ".jsx"}

// resolveTSImport handles a relative specifier two ways: `./Thing` naming a sibling file
// (dependency = that file's own directory) and `./sub` naming a subdirectory's barrel
// index (dependency = the subdirectory itself). A bare specifier ("react", a tsconfig
// path alias) has no relative anchor to resolve from and is dropped — this package knows
// nothing about node_modules or tsconfig, and guessing would invent a node for a
// dependency that may not even be first-party.
func resolveTSImport(imp, from string, fileSet map[string]bool) (string, bool) {
	if !strings.HasPrefix(imp, ".") {
		return "", false
	}
	joined := path.Join(from, imp)
	for _, ext := range tsExts {
		if fileSet[joined+ext] {
			return path.Dir(joined), true
		}
		if fileSet[joined+"/index"+ext] {
			return joined, true
		}
	}
	return "", false
}

// resolveJavaImport has no source root to anchor from (Maven's src/main/java, or nothing
// at all), so it matches by suffix against the files this repo actually has. The full
// dotted name is tried as a file first — the common case, `import a.b.C` naming class C —
// then trimmed one dotted segment at a time, because a static member import
// (`import static a.b.C.FIELD`) and a wildcard import (`import a.b.*`, arriving here with
// the star already stripped by the extractor) both carry a trailing segment that is not
// part of any file's path. Once no segment remains, the untouched import is tried once
// more as a directory itself — the wildcard case, where it was a package name all along.
func resolveJavaImport(imp string, files []string, dirs map[string]bool) (string, bool) {
	for p := imp; p != ""; p = trimLastSegment(p) {
		if f, ok := matchSuffix(strings.ReplaceAll(p, ".", "/")+".java", files); ok {
			return path.Dir(f), true
		}
	}
	return matchSuffix(strings.ReplaceAll(imp, ".", "/"), keysOf(dirs))
}

func trimLastSegment(p string) string {
	if i := strings.LastIndex(p, "."); i >= 0 {
		return p[:i]
	}
	return ""
}

// matchSuffix picks the lexicographically first match so an ambiguous suffix — two
// modules sharing a package fragment under different source roots — resolves the same
// way on every run — a project-wide determinism rule — rather than however the caller's
// slice happened to be ordered.
func matchSuffix(suffix string, candidates []string) (string, bool) {
	var best string
	for _, c := range candidates {
		if (c == suffix || strings.HasSuffix(c, "/"+suffix)) && (best == "" || c < best) {
			best = c
		}
	}
	return best, best != ""
}

func keysOf(m map[string]bool) []string {
	return slices.Sorted(maps.Keys(m))
}
