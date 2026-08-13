package report

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Ask is a topic that was asked about, with how often and whether a note answers it now.
type Ask struct {
	Topic   string
	Count   int
	Written bool
}

// GapsInput is what gaps.md renders from. Asks comes from the capture log in `.forge/`.
type GapsInput struct {
	Asks []Ask
	Now  time.Time
}

// minAsks is §B.4's threshold: asked at least twice and never written. Once is a passing
// curiosity; twice is a topic you keep needing and do not have.
const minAsks = 2

// RenderGaps produces gaps.md — the personal curriculum.
//
// Asks comes from `.forge/log.jsonl`, written by `forge recall` when telemetry is
// enabled (pkg/telemetry) — never by the D3 capture hook, which harvests written notes,
// not questions asked. On a vault with telemetry off, or one that predates it, the log is
// empty and the report says so rather than inventing a substitute: filling this page from
// some other signal — unlinked mentions, say — would produce a plausible list of topics
// that nobody actually asked about, which is worse than an empty page because it cannot
// be told apart from a real one.
func RenderGaps(in GapsInput) []byte {
	gaps := unwritten(in.Asks)
	var b strings.Builder
	header(&b, "Gaps", gapsSummary(in, gaps), in.Now)
	if len(gaps) == 0 {
		writeNoAskLog(&b, in)
		return []byte(b.String())
	}
	b.WriteString("\n")
	for _, a := range head(gaps, 30) {
		fmt.Fprintf(&b, "- **%s** — asked %dx\n", a.Topic, a.Count)
	}
	return []byte(b.String())
}

// HitRate is asks resolved by an existing note, as a percentage of all asks recorded —
// weekly.md's Vault section reuses this rather than redefining "resolved" a second way.
func HitRate(asks []Ask) float64 {
	var written, total int
	for _, a := range asks {
		total += a.Count
		if a.Written {
			written += a.Count
		}
	}
	if total == 0 {
		return 0
	}
	return float64(written) / float64(total) * 100
}

func unwritten(asks []Ask) []Ask {
	var out []Ask
	for _, a := range asks {
		if !a.Written && a.Count >= minAsks {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Topic < out[j].Topic
	})
	return out
}

func gapsSummary(in GapsInput, gaps []Ask) string {
	if len(in.Asks) == 0 {
		return "**No ask log yet.**\n"
	}
	return fmt.Sprintf("**%d %s asked %d+ times with no note.**\n",
		len(gaps), plural(len(gaps), "topic", "topics"), minAsks)
}

// writeNoAskLog distinguishes the two empty states, which mean opposite things: no data
// versus data showing nothing missing.
func writeNoAskLog(b *strings.Builder, in GapsInput) {
	if len(in.Asks) > 0 {
		empty(b, "every topic asked about more than once has a note")
		return
	}
	b.WriteString("\nThis report is filled by the capture log in `.forge/`, which has " +
		"nothing in it yet — every note in the vault is `origin: import`, so no ask has " +
		"been recorded. It starts paying off once questions are being captured.\n\n" +
		"For where the wiki is thin today, see `coverage.md` (stacks with no notes) and " +
		"`../moc/codebase.md` (code that churns with nothing written about it).\n")
}
