// Package codeindex turns source files into a symbol table drift can compare against.
//
// It is the only package in the tree that needs cgo, because go-tree-sitter is a C
// library. The cgo surface is confined here by build tag: with CGO_ENABLED=1 the
// tree-sitter parser is compiled in; with CGO_ENABLED=0 the package still builds and
// Parse returns ErrUnavailable, so `CGO_ENABLED=0 go build ./...` keeps working and
// every downstream package — pkg/drift in particular — stays pure Go and stays
// cross-compilable. Release builds set CGO_ENABLED=1.
//
// Language coverage is Java and TypeScript. STACK §10 left this open ("start with Java
// + Kotlin only") and AUDIT §7 settled it with a count: the vault's reference corpus is
// 42 Java files, 308 TypeScript, and *zero* Kotlin. Shipping a Kotlin grammar would add
// a C dependency for a language nothing cites.
package codeindex

import (
	"errors"
	"hash/fnv"
	"path/filepath"
	"strings"
)

// ErrUnavailable is returned by Parse when the binary was built without cgo.
var ErrUnavailable = errors.New("codeindex: built without cgo; rebuild with CGO_ENABLED=1")

// Symbol is one declaration. Start and End are 1-based inclusive line numbers, which is
// what a note's `file:line` citation speaks and what line-number auto-repair rewrites.
type Symbol struct {
	Name     string `json:"name"` // qualified: "OrderConsumer.receive"
	Kind     string `json:"kind"` // class, interface, enum, record, method, function
	Start    int    `json:"start"`
	End      int    `json:"end"`
	BodyHash string `json:"body_hash"` // whitespace-normalized hash of the body
}

// File is one parsed source file.
type File struct {
	Path    string   `json:"path"` // repo-relative, forward slashes
	Lang    string   `json:"lang"`
	Symbols []Symbol `json:"symbols"`
}

// Extractor doubles as this cache's format version (BACKLOG B-013): Load rejects any
// stamp but this, treating a mismatch as a cache miss rather than a bad parse. Bump it
// whenever declKinds or kindOf changes (an older extractor holds fewer symbols, and a
// missing symbol is exactly what drift reads as BROKEN) — but also whenever Symbol or
// File's serialized shape changes, even if extraction logic itself doesn't. Go's
// json.Unmarshal is lenient about missing/added fields, so a shape change alone would
// otherwise unmarshal an old cache "successfully" into a struct it was never written for.
const Extractor = 2

// Index is the serialized form written to .forge/code-index.json.
type Index struct {
	Repo      string            `json:"repo"`
	Commit    string            `json:"commit"`
	Extractor int               `json:"extractor"`
	Files     map[string]File   `json:"files"`
	Deps      map[string]string `json:"deps"` // declared dependency -> version
}

// Lang maps a file extension to a supported grammar, or "" for a file this package
// does not parse.
func Lang(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".java":
		return "java"
	case ".ts", ".tsx", ".js", ".jsx":
		return "typescript"
	}
	return ""
}

// Lookup finds a symbol by exact qualified name, then by its trailing member. Notes
// cite `receive` as often as `OrderConsumer.receive`, and demanding the qualified form
// would report a symbol that plainly exists as missing.
func (f File) Lookup(name string) (Symbol, bool) {
	for _, s := range f.Symbols {
		if s.Name == name {
			return s, true
		}
	}
	for _, s := range f.Symbols {
		if i := strings.LastIndex(s.Name, "."); i >= 0 && s.Name[i+1:] == name {
			return s, true
		}
	}
	return Symbol{}, false
}

// FindSymbol searches the whole index for a symbol, returning the file that declares
// it. A citation that names only a class has no path to resolve through, so this is
// how pkg/drift decides whether it still exists.
func (ix Index) FindSymbol(name string) (string, Symbol, bool) {
	for path, f := range ix.Files {
		if s, ok := f.Lookup(name); ok {
			return path, s, true
		}
	}
	return "", Symbol{}, false
}

// hashBody collapses every run of whitespace to one space before hashing, so a
// reformat, a re-indent or a line ending change is not reported as a body change.
// SUSPECT means "the behaviour this note describes may have moved"; gofmt does not.
func hashBody(src []byte) string {
	h := fnv.New64a()
	h.Write([]byte(strings.Join(strings.Fields(string(src)), " ")))
	return strings.ToLower(hexOf(h.Sum64()))
}

func hexOf(v uint64) string {
	const digits = "0123456789abcdef"
	buf := make([]byte, 16)
	for i := 15; i >= 0; i-- {
		buf[i] = digits[v&0xf]
		v >>= 4
	}
	return string(buf)
}
