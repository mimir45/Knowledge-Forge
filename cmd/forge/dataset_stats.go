package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"knowledge-forge/pkg/dataset"
)

const datasetStatsUsage = `usage: forge dataset-stats [--vault DIR]

Reports how much training data each of ADDENDUM D.1's five capture tiers has
accumulated, and what that volume is honestly enough for. Zero model calls; a tier that
has captured nothing is a zero row, not an error.

The second half is deliberately unexciting. ADDENDUM D.2 exists because these pitches
usually overclaim, and the whole value of the section is that it will tell you 180 pairs
are enough for a small routing adapter and nothing else.
`

func cmdDatasetStats(args []string) int {
	fs := flag.NewFlagSet("forge dataset-stats", flag.ContinueOnError)
	vaultDir := fs.String("vault", "", "vault root; defaults to config vault_path, then .")
	fs.Usage = func() { fmt.Fprint(os.Stderr, datasetStatsUsage); fs.PrintDefaults() }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	root, code := vaultOrExit("dataset-stats", *vaultDir)
	if code != 0 {
		return code
	}
	return runDatasetStats(root, os.Stdout)
}

// runDatasetStats writes to an io.Writer so it is testable without a stdout pipe — the
// same split cmd/forge/stats.go uses, and the same single tabwriter across sections.
func runDatasetStats(root string, w io.Writer) int {
	stats := dataset.Stats(root)
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	writeTierCounts(tw, stats)
	writeAdequacy(tw, stats)
	tw.Flush()
	return 0
}

func writeTierCounts(tw *tabwriter.Writer, stats []dataset.TierStats) {
	fmt.Fprintf(tw, "Captured pairs:\n")
	fmt.Fprintf(tw, "  set\tkind\tpairs\tfirst\tlast\n")
	for _, s := range stats {
		if s.Err != "" {
			fmt.Fprintf(tw, "  %s\t%s\tunreadable\t%s\t\n", s.Tag, s.Kind, s.Err)
			continue
		}
		fmt.Fprintf(tw, "  %s\t%s\t%d\t%s\t%s\n", s.Tag, s.Kind, s.Count, day(s.From), day(s.To))
	}
}

func day(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("2006-01-02")
}

func writeAdequacy(tw *tabwriter.Writer, stats []dataset.TierStats) {
	fmt.Fprintf(tw, "\nWhat this is enough for (ADDENDUM D.2):\n")
	for _, s := range stats {
		if s.Err == "" {
			fmt.Fprintf(tw, "  %s\t%d\t%s\n", s.Tag, s.Count, adequacy(s.Tag, s.Count))
		}
	}
	fmt.Fprintf(tw, "\n  Any tier makes an evaluation set long before it can be trained on.\n")
	fmt.Fprintf(tw, "  The honest sequencing is eval sets, then routing, then style, then\n")
	fmt.Fprintf(tw, "  drafting, then advisor distillation — in that order.\n")
}

// adequacy is ADDENDUM D.2's table read back one tier at a time. Every string here is
// bounded above deliberately: the section's whole purpose is to stop someone spending a
// month fine-tuning on 200 examples and concluding fine-tuning does not work.
func adequacy(tag string, n int) string {
	switch tag {
	case "d1":
		return band(n, "an eval set for the router; too few to train on",
			"a 1-3B routing LoRA, and nothing else", "a routing LoRA comfortably; still not a drafting model")
	case "d5":
		return band(n, "an eval set for note style; too few to train on",
			"a style adapter that mimics your note voice", "a style adapter, and half of a drafting corpus")
	case "d3":
		return dpoBand(n, "the scarcest data here — keep accumulating")
	case "d2":
		return dpoBand(n, "advisor critiques; distillation needs 10k+ pairs")
	case "d4":
		return band(n, "too few to train on", "SFT repair examples",
			"repair examples that fold into a 7-8B drafting LoRA")
	}
	return ""
}

// band applies D.2's 100 / 1000 thresholds, which are the two that actually change the
// answer: below 100 nothing trains, 100-500 is the small-adapter window, and 1k-5k is
// where a 7-8B drafting LoRA becomes credible.
func band(n int, under100, under1000, over string) string {
	switch {
	case n < 100:
		return under100
	case n < 1000:
		return under1000
	}
	return over
}

// dpoBand exists because D2 and D3 share one destination — DPO advisor distillation at
// 10k+ — and that number is far enough above the others to deserve its own ladder.
func dpoBand(n int, low string) string {
	switch {
	case n < 100:
		return low
	case n < 1000:
		return "an eval set, and a preference sample worth inspecting by hand"
	case n < 10000:
		return "a 7-8B drafting LoRA in your stack and voice, for that task only"
	}
	return "DPO on d2+d3: distilling advisor judgement is now realistic"
}
