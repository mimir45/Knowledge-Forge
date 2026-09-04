// Package sentinel edits managed blocks of text inside files that forge does not
// otherwise own — a code repo's CLAUDE.md fragments and.
package sentinel

import "fmt"

// Style is one comment syntax. Close is empty for line comments (Go/Java "//", Python
// "#"); Markdown's "<!--"/"-->" needs both ends on the marker line.
type Style struct {
	Open, Close string
}

var (
	Markdown = Style{Open: "<!--", Close: "-->"}
	Slash    = Style{Open: "//"}
	Hash     = Style{Open: "#"}
)

func (s Style) begin(id string) string { return s.marker("begin", id) }
func (s Style) end(id string) string   { return s.marker("end", id) }

func (s Style) marker(which, id string) string {
	if s.Close == "" {
		return fmt.Sprintf("%s forge:%s:%s", s.Open, id, which)
	}
	return fmt.Sprintf("%s forge:%s:%s %s", s.Open, id, which, s.Close)
}
