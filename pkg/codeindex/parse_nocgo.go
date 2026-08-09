//go:build !cgo

package codeindex

// The no-cgo build keeps the package compiling so `CGO_ENABLED=0 go build ./...` still
// covers the whole tree and every downstream package stays cross-compilable. It does
// not substitute a second, weaker extractor: two implementations of "what symbols does
// this file declare" would make drift's verdicts depend on how the binary was built,
// and drift's whole contract is that verdicts are a pure function of tree state.

// Available reports whether this build can parse source.
func Available() bool { return false }

// Parse always fails in a no-cgo build.
func Parse(path string, src []byte) (File, error) { return File{}, ErrUnavailable }
