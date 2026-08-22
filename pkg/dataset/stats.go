package dataset

import "time"

// TierStats is one tier's accumulated volume. Err carries a read failure rather than
// aborting the whole report: dataset-stats is a read-only look at five independent files
// and one torn tier should not hide the other four. The strict reader's message already
// names the file and line, so it is worth printing as-is.
type TierStats struct {
	Tag      string
	Kind     string
	Count    int
	From, To time.Time
	Err      string
}

// Stats reports every tier's volume and date range. It reuses the export path's strict
// reader on purpose — a line dataset-stats counts but export would refuse is a number
// that lies about what you can actually export.
func Stats(vaultRoot string) []TierStats {
	out := make([]TierStats, 0, len(Tiers()))
	for _, t := range Tiers() {
		out = append(out, statsFor(vaultRoot, t))
	}
	return out
}

func statsFor(vaultRoot string, t Tier) TierStats {
	s := TierStats{Tag: t.Tag, Kind: t.Kind}
	recs, err := loadTier(vaultRoot, t)
	if err != nil {
		s.Err = err.Error()
		return s
	}
	s.Count = len(recs)
	s.From, s.To = span(recs)
	return s
}
