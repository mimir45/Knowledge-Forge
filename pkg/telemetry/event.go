// Package telemetry writes DESIGN §14's ask-event log — the sole source `pkg/report`'s
// Gaps and Staleness sections read for how often something was asked. The invariant is
// narrower than "don't log secrets": it logs a topic label and a hash, never the raw
// question, code, or file contents (see QHash).
package telemetry

import "time"

// Event mirrors DESIGN §14's schema field-for-field. json tags match the doc's example
// line exactly so `.forge/log.jsonl` is diffable against that example by eye.
type Event struct {
	TS             time.Time `json:"ts"`
	Event          string    `json:"event"`
	QHash          string    `json:"q_hash"`
	Topic          string    `json:"topic"`
	Stack          []string  `json:"stack"`
	Decision       string    `json:"decision"`
	RecallTopScore float64   `json:"recall_top_score"`
	DurationMS     int64     `json:"duration_ms"`
	Sources        int       `json:"sources"`
	Project        string    `json:"project"`
	// RunID is BACKLOG B-035's correlation key, added after DESIGN §14 shipped — an
	// addition to the schema, not a rename, so an older log line simply decodes with it
	// empty. See NewRunID and cmd/forge/recall.go's runRecall.
	RunID string `json:"run_id,omitempty"`
}
