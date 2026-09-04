// Package coderef is the citation layer drift stands on.
package coderef

import "strings"

// Kind records how confident the extractor is about what it found, which resolution
// later needs.
type Kind string

const (
	KindPath   Kind = "path"   // path-shaped, with or without a line number
	KindSymbol Kind = "symbol" // a type or member name, no path
)

// Ref is one code citation as the note wrote it, before resolution.
type Ref struct {
	Raw    string `json:"raw"` // verbatim, for the report
	Kind   Kind   `json:"kind"`
	Path   string `json:"path,omitempty"`   // may be shorthand; forward slashes
	Symbol string `json:"symbol,omitempty"` // "Money", "Money.add", "" when unknown
	Line   int    `json:"line,omitempty"`   // 0 when the note cited no line
	Repo   string `json:"repo,omitempty"`   // set by the canonical form, or by Resolve
	Note   string `json:"note,omitempty"`   // vault-relative path of the citing note
}

// Status is the outcome of resolving a Ref against one repository's file list.
type Status string

const (
	Resolved   Status = "resolved"   // exactly one file matched
	Ambiguous  Status = "ambiguous"  // several files matched the shorthand
	Unresolved Status = "unresolved" // no file matched — NF-4's 14-of-19 case
)

// Resolution pairs a Ref with what the repository said about it.
type Resolution struct {
	Ref       Ref      `json:"ref"`
	Status    Status   `json:"status"`
	RepoPath  string   `json:"repo_path,omitempty"` // repo-relative, resolved
	Ambiguity []string `json:"ambiguity,omitempty"` // the competing matches
}

// Segments splits a citation path into its components, dropping the leading "./" and
// any empty segments a doubled slash leaves behind.
func Segments(p string) []string {
	var out []string
	for _, s := range strings.Split(strings.TrimPrefix(filepathSlash(p), "./"), "/") {
		if s != "" && s != "." {
			out = append(out, s)
		}
	}
	return out
}

func filepathSlash(p string) string { return strings.ReplaceAll(p, `\`, "/") }
