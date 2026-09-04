package recall

import "time"

// IsStale reports whether a note has aged past its freshness window.
func IsStale(d Doc, now time.Time) bool {
	if d.FreshnessDays <= 0 {
		return false
	}
	when, ok := parseDate(d.Verified)
	if !ok {
		if when, ok = parseDate(d.Updated); !ok {
			return true // undatable: cannot vouch for it, so do not answer from it
		}
	}
	return now.Sub(when) > time.Duration(d.FreshnessDays)*24*time.Hour
}

// parseDate accepts the schema's `YYYY-MM-DD` and the RFC3339 timestamps Obsidian
// plugins sometimes write instead.
func parseDate(s string) (time.Time, bool) {
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
