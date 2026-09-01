package main

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/mimir45/Knowledge-Forge/pkg/report"
)

// loadAskLog reads .forge/log.jsonl and returns the same ask counts in the two shapes
// staleness.md and gaps.md need: slug-keyed (StalenessInput.Asks) and topic-keyed
// ([]report.Ask). Written is resolved by exact match against slugs — DESIGN §14's topic
// field is already a slug-shaped label (recall's logAsk writes vault.Slug output), so no
// fuzzy matching is attempted. Tolerant of a missing file: every vault before its first
// telemetry-enabled `forge recall` run has none.
func loadAskLog(path string, slugs map[string]string) (bySlug map[string]int, asks []report.Ask) {
	counts := countAskTopics(path)
	known := knownSlugs(slugs)
	bySlug = map[string]int{}
	for topic, n := range counts {
		written := known[topic]
		if written {
			bySlug[topic] = n
		}
		asks = append(asks, report.Ask{Topic: topic, Count: n, Written: written})
	}
	return bySlug, asks
}

// countAskTopics tallies event:"ask" lines by topic, skipping anything that doesn't
// parse — a hand-edited or truncated log line must not abort the whole weekly run.
func countAskTopics(path string) map[string]int {
	counts := map[string]int{}
	f, err := os.Open(path)
	if err != nil {
		return counts
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var line struct{ Event, Topic string }
		if json.Unmarshal(sc.Bytes(), &line) != nil {
			continue
		}
		if line.Event == "ask" && line.Topic != "" {
			counts[line.Topic]++
		}
	}
	return counts
}

func knownSlugs(slugs map[string]string) map[string]bool {
	out := make(map[string]bool, len(slugs))
	for _, s := range slugs {
		out[s] = true
	}
	return out
}
