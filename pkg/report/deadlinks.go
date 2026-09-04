package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mimir45/Knowledge-Forge/pkg/linkcheck"
)

// Citation ties a checked URL back to the notes that cite it. One URL is often cited by
// several notes, and the report is only actionable if it names them.
type Citation struct {
	linkcheck.Status
	Notes []string // vault-relative paths, sorted
}

// DeadlinksInput is what deadlinks.md renders from.
type DeadlinksInput struct {
	Citations  []Citation
	FirstParty int
	Slugs      map[string]string
	Now        time.Time
}

// RenderDeadlinks produces deadlinks.md — citations that have rotted.
func RenderDeadlinks(in DeadlinksInput) []byte {
	dead := withVerdict(in.Citations, linkcheck.Dead)
	unreachable := withVerdict(in.Citations, linkcheck.Unreachable)
	var b strings.Builder
	header(&b, "Dead links", deadlinksSummary(in, dead, unreachable), in.Now)
	writeCitations(&b, in, dead, "Dead — the server answered and said no",
		"These are gone. Find a replacement source or drop the claim they support.")
	writeCitations(&b, in, unreachable, "Unreachable — no answer",
		"Not evidence of anything yet. Re-run when the network is known good.")
	return []byte(b.String())
}

func withVerdict(cs []Citation, v linkcheck.Verdict) []Citation {
	var out []Citation
	for _, c := range cs {
		if c.Verdict == v {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].URL < out[j].URL })
	return out
}

func deadlinksSummary(in DeadlinksInput, dead, unreachable []Citation) string {
	alive := len(in.Citations) - len(dead) - len(unreachable)
	return fmt.Sprintf("**%d dead, %d unreachable, %d alive** of %d cited %s.%s\n\n"+
		"Unreachable is counted apart from dead on purpose: no answer is not the same as "+
		"an answer of no, and one offline run must not read as a rotted vault.\n",
		len(dead), len(unreachable), alive, len(in.Citations),
		plural(len(in.Citations), "URL", "URLs"), firstPartyNote(in.FirstParty))
}

func firstPartyNote(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf(" A further %d %s first-party — a vault-relative path, which this "+
		"report cannot and should not check over HTTP.",
		n, plural(n, "citation is", "citations are"))
}

func writeCitations(b *strings.Builder, in DeadlinksInput, cs []Citation, title, lede string) {
	fmt.Fprintf(b, "\n## %s — %d\n", title, len(cs))
	if len(cs) == 0 {
		empty(b, "none")
		return
	}
	fmt.Fprintf(b, "\n%s\n\n", lede)
	for _, c := range cs {
		writeCitation(b, in, c)
	}
}

func writeCitation(b *strings.Builder, in DeadlinksInput, c Citation) {
	fmt.Fprintf(b, "- %s — %s\n", c.URL, detailOf(c))
	for _, rel := range c.Notes {
		fmt.Fprintf(b, "  - cited by %s\n", note(in.Slugs[rel], rel))
	}
}

func detailOf(c Citation) string {
	if c.Code != 0 {
		return fmt.Sprintf("HTTP %d", c.Code)
	}
	if c.Detail != "" {
		return c.Detail
	}
	return string(c.Verdict)
}
