//go:build !cgo

package codeindex

// The no-cgo build keeps the package compiling so `CGO_ENABLED=0 go build ./...` still
// covers the whole tree and every downstream package stays cross-compilable.

// Available reports whether this build can parse source.
func Available() bool { return false }

// Parse always fails in a no-cgo build.
func Parse(path string, src []byte) (File, error) { return File{}, ErrUnavailable }
