package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// WeekKey formats t's ISO week as "YYYY-Www", zero-padded so plain string comparison
// sorts weeks correctly — including across a year boundary, where ISOWeek's own returned
// year (not t.Year()) is what keys a week that spans two calendar years.
func WeekKey(t time.Time) string {
	y, w := t.ISOWeek()
	return fmt.Sprintf("%04d-W%02d", y, w)
}

// WeeklyStore persists one VaultStats snapshot per ISO week under .forge/, so weekly.md
// can show a week-over-week delta without recomputing history. Mirrors
// pkg/drift/demotions.go's Store/OpenStore/Save shape.
type WeeklyStore struct {
	Weeks map[string]VaultStats `json:"weeks"`

	dir   string
	dirty bool
}

func OpenWeeklyStore(dir string) *WeeklyStore {
	s := &WeeklyStore{Weeks: map[string]VaultStats{}, dir: dir}
	b, err := os.ReadFile(filepath.Join(dir, "weekly-stats.json"))
	if err == nil {
		json.Unmarshal(b, s) //nolint:errcheck // a corrupt store loses one week's delta, never a verdict
	}
	if s.Weeks == nil {
		s.Weeks = map[string]VaultStats{}
	}
	return s
}

// Prev returns the stats for the most recent week strictly before key, or nil if none —
// a second run in the same week must neither zero the delta nor duplicate a snapshot, so
// "most recent prior week" is deliberately not "the last run".
func (s *WeeklyStore) Prev(key string) *VaultStats {
	best := ""
	for k := range s.Weeks {
		if k < key && k > best {
			best = k
		}
	}
	if best == "" {
		return nil
	}
	v := s.Weeks[best]
	return &v
}

// Record stores this week's snapshot, overwriting any prior run the same week — a second
// run today must update today's numbers, not stack a duplicate entry.
func (s *WeeklyStore) Record(key string, stats VaultStats) {
	s.Weeks[key] = stats
	s.dirty = true
}

// Prune keeps only the most recent n weeks by key order — .forge/ is not a place we want
// unbounded history to accumulate.
func (s *WeeklyStore) Prune(n int) {
	if len(s.Weeks) <= n {
		return
	}
	keys := make([]string, 0, len(s.Weeks))
	for k := range s.Weeks {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys[:len(keys)-n] {
		delete(s.Weeks, k)
	}
	s.dirty = true
}

func (s *WeeklyStore) Save() error {
	if !s.dirty {
		return nil
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "weekly-stats.json"), b, 0o644)
}
