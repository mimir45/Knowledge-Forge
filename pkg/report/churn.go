package report

import (
	"fmt"
	"strings"
	"time"

	"knowledge-forge/pkg/gitsig"
)

// ChurnInput is what churn.md renders from. The stats are over the *vault's* history, not
// a code repository's: §B.4's churn.md answers "which notes keep being rewritten".
// moc/codebase.md is where code churn lives, and the two must not be conflated — they are
// the same measurement over different repositories and they mean different things.
type ChurnInput struct {
	Stats  *gitsig.Stats
	Slugs  map[string]string // vault-relative path -> slug
	Months int               // window the history was read over; 0 means all of it
	Now    time.Time
}

// RenderChurn produces churn.md — volatile knowledge.
//
// A note rewritten nine times is not a bad note. It is a note about something that keeps
// moving, and that is a signal about the system rather than the writing: the areas of the
// codebase whose notes churn hardest are the areas whose behaviour is least settled.
func RenderChurn(in ChurnInput) []byte {
	var b strings.Builder
	header(&b, "Churn", churnSummary(in), in.Now)
	writeTopChurn(&b, in)
	writeCoupled(&b, in)
	return []byte(b.String())
}

func churnSummary(in ChurnInput) string {
	window := "all history"
	if in.Months > 0 {
		window = fmt.Sprintf("the last %d months", in.Months)
	}
	return fmt.Sprintf("**%d commits over %s**, touching %d %s.\n",
		in.Stats.Commits, window, len(in.Stats.Churn),
		plural(len(in.Stats.Churn), "file", "files"))
}

func writeTopChurn(b *strings.Builder, in ChurnInput) {
	top := gitsig.TopChurn(in.Stats, 25)
	b.WriteString("\n## Most rewritten\n\n")
	if len(top) == 0 {
		b.WriteString("_no history yet_\n")
		return
	}
	for _, f := range top {
		owner, share := in.Stats.Owner(f.File)
		fmt.Fprintf(b, "- **%dx** — %s", f.Count, note(in.Slugs[f.File], f.File))
		if owner != "" {
			fmt.Fprintf(b, " · %s (%.0f%%)", owner, share*100)
		}
		b.WriteString("\n")
	}
}

// writeCoupled shows notes that keep being edited in the same commit. Two notes that
// always change together are usually one idea split across two files, or a pair that
// should link to each other and does not.
func writeCoupled(b *strings.Builder, in ChurnInput) {
	pairs := gitsig.TopCoupled(in.Stats, 2, 15)
	b.WriteString("\n## Always edited together\n\n")
	if len(pairs) == 0 {
		b.WriteString("_no pair has changed together more than once_\n")
		return
	}
	b.WriteString("Usually one idea split across two files, or a pair that should link " +
		"to each other.\n\n")
	for _, p := range pairs {
		fmt.Fprintf(b, "- **%dx** — %s · %s\n", p.Count,
			note(in.Slugs[p.A], p.A), note(in.Slugs[p.B], p.B))
	}
}
