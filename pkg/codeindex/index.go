// Package codeindex turns source files into a symbol table drift can compare against.
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
	// Imports is each import/re-export this file declares, in the source language's own
	// shape — a dotted Java package+class name.
	Imports []string `json:"imports,omitempty"`
}

// Extractor doubles as this cache's format version: Load rejects any stamp but this,
// treating a mismatch as a cache miss rather than a bad parse.
const Extractor = 3

// Index is the serialized form Save writes. One Index describes one repo at one commit;
// the file it lands in is the caller's choice, not this package's (see Save).
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

// Lookup finds a symbol by exact qualified name, then by its trailing member.
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
// it.
func (ix Index) FindSymbol(name string) (string, Symbol, bool) {
	for path, f := range ix.Files {
		if s, ok := f.Lookup(name); ok {
			return path, s, true
		}
	}
	return "", Symbol{}, false
}

// hashBody collapses every run of whitespace to one space before hashing.
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
