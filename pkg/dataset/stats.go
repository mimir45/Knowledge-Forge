package dataset

import "time"

// TierStats is one tier's accumulated volume.
type TierStats struct {
	Tag      string
	Kind     string
	Count    int
	From, To time.Time
	Err      string
}

// Stats reports every tier's volume and date range.
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
