package report

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// CostInput is what cost.md renders from. pkg/report stays config-free, matching every
// other report here.
type CostInput struct {
	SpentToday map[string]float64 // tier ("api"/"advisor") -> USD spent today
	// CapPerDay: pkg/engine's availableMetered treats a cap of 0 as exhausted (remaining =
	// cap - spent <= 0), not unmetered.
	CapPerDay   map[string]float64 // tier -> configured daily cap
	OnExhausted string             // "degrade" | "queue" | "stop"
	StageEngine map[string]string  // pipeline stage -> the engine that would win today
	QueuedNotes int                // notes carrying pending_advisor: true
	Now         time.Time
}

// RenderCost produces cost.md — where the money goes, and what today's config would do
// about a stage that runs out of it.
func RenderCost(in CostInput) []byte {
	var b strings.Builder
	header(&b, "Cost", costSummary(in), in.Now)
	writeSpend(&b, in)
	writeStageEngine(&b, in)
	writeQueue(&b, in)
	return []byte(b.String())
}

func costSummary(in CostInput) string {
	total := 0.0
	for _, usd := range in.SpentToday {
		total += usd
	}
	return fmt.Sprintf("**$%.2f spent today** across metered tiers. `on_exhausted: %s`.\n",
		total, in.OnExhausted)
}

func writeSpend(b *strings.Builder, in CostInput) {
	b.WriteString("\n## Spend today\n\n")
	tiers := sortedKeys2(in.SpentToday, in.CapPerDay)
	if len(tiers) == 0 {
		empty(b, "no metered tier is configured")
		return
	}
	for _, t := range tiers {
		writeTierSpend(b, t, in.SpentToday[t], in.CapPerDay[t])
	}
}

// writeTierSpend prints one tier's line. pkg/engine's availableMetered reads cap 0 as
// "budget exhausted for today" (remaining <= 0).
func writeTierSpend(b *strings.Builder, tier string, spent, cap float64) {
	if cap == 0 {
		fmt.Fprintf(b, "- **%s** — $%.2f spent, cap $0.00: unavailable, not routed here\n", tier, spent)
		return
	}
	fmt.Fprintf(b, "- **%s** — $%.2f of $%.2f\n", tier, spent, cap)
}

// writeStageEngine is the only place in these nine.
func writeStageEngine(b *strings.Builder, in CostInput) {
	b.WriteString("\n## Per-stage engine\n\n")
	stages := make([]string, 0, len(in.StageEngine))
	for s := range in.StageEngine {
		stages = append(stages, s)
	}
	sort.Strings(stages)
	for _, s := range stages {
		fmt.Fprintf(b, "- **%s** — %s\n", s, in.StageEngine[s])
	}
}

func writeQueue(b *strings.Builder, in CostInput) {
	b.WriteString("\n## Queue\n\n")
	if in.QueuedNotes == 0 {
		b.WriteString("_nothing waiting on the advisor queue_\n")
		return
	}
	fmt.Fprintf(b, "%d %s waiting on `on_exhausted: queue` — drained by the weekly checker.\n",
		in.QueuedNotes, plural(in.QueuedNotes, "note", "notes"))
}

// sortedKeys2 unions two maps' keys (a tier may have spend with no cap, or a cap with no
// spend yet) and returns them sorted, so the two never silently disagree on which tiers exist.
func sortedKeys2(a, b map[string]float64) []string {
	seen := map[string]bool{}
	var out []string
	for k := range a {
		seen[k] = true
	}
	for k := range b {
		seen[k] = true
	}
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
