package report

import (
	"fmt"
	"strings"
	"time"
)

// Report is one rendered file on its way to <vault>/reports/.
type Report struct {
	Name string // file name, e.g. "drift.md"
	Body []byte
}

// header opens every report the same way.
//
// The date is a date and never a timestamp, for the same reason RenderIndex's is: these
// nine files land in a git repository the user commits. A clock in the header would make
// every weekly run a nine-file diff of nothing, and a diff that is always noise is a diff
// nobody reads. Two runs on one day must produce identical bytes.
func header(b *strings.Builder, title, subtitle string, now time.Time) {
	fmt.Fprintf(b, "# %s — %s\n\n", title, now.Format("2006-01-02"))
	if subtitle != "" {
		fmt.Fprintf(b, "%s\n", subtitle)
	}
}

// note renders a wikilink, falling back to the path when a note has no slug — a note that
// failed the contract still has to be nameable in the report that says so.
func note(slug, rel string) string {
	if slug == "" {
		return "`" + rel + "`"
	}
	return "[[" + slug + "]]"
}

// empty writes the "nothing to report" state. It is a sentence rather than a blank section
// because an empty section reads like a bug, and half of these reports are supposed to be
// empty on a healthy vault.
func empty(b *strings.Builder, what string) {
	fmt.Fprintf(b, "\n_%s_\n", what)
}

// plural is for headings that would otherwise read "1 notes".
func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// days returns whole days between two instants, rounded down.
func days(from, to time.Time) int {
	if from.IsZero() {
		return 0
	}
	return int(to.Sub(from).Hours() / 24)
}
