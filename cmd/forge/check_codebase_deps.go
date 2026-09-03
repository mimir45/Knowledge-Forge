package main

import (
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/mimir45/Knowledge-Forge/pkg/codeindex"
)

// dependsOn folds codeindex.File.Imports into the directory-level dependency graph
// CodeGroup.DependsOn needs.
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

// sortedDeps is where "deduped and sorted" (the plan's own words) actually happens.
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

// resolveTSImport handles a relative specifier two ways.
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

// resolveJavaImport has no source root to anchor from.
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

// matchSuffix picks the lexicographically first match so an ambiguous suffix.
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
