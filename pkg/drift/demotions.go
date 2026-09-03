package drift

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Demotion is what .forge/ remembers about a note drift knocked down: the confidence the
// note held before, and the sha at which it lost it.
//
// This file is a restore target, never an input to a verdict. Check never reads it, which
// is exactly why a revert restores a note on tree state alone — and why deleting .forge/
// costs at most the memory of what a note's confidence used to be, not a wrong answer.
type Demotion struct {
	Note       string `json:"note"` // vault-relative path
	Slug       string `json:"slug"`
	Confidence string `json:"confidence"` // the value held *before* demotion
	Repo       string `json:"repo"`
	SHA        string `json:"sha"` // HEAD at the moment of demotion
	Ref        string `json:"ref"`
	At         string `json:"at"` // YYYY-MM-DD
}

// Store persists demotions under .forge/. It is keyed by slug — the original spec says
// "slug+sha keyed", and the sha rides in the record rather than the key so a restore is
// one lookup instead of a scan: at restore time the demoting sha is precisely what the
// caller does not know.
type Store struct {
	Notes map[string]Demotion `json:"notes"`

	dir   string
	dirty bool
}

func OpenStore(dir string) *Store {
	s := &Store{Notes: map[string]Demotion{}, dir: dir}
	b, err := os.ReadFile(filepath.Join(dir, "demotions.json"))
	if err == nil {
		json.Unmarshal(b, s) //nolint:errcheck // a corrupt store loses restore targets, never verdicts
	}
	if s.Notes == nil {
		s.Notes = map[string]Demotion{}
	}
	return s
}

// Record remembers a pre-demotion confidence, once. A note that stays broken across ten
// commits must not have its remembered confidence overwritten with "low" on the second.
func (s *Store) Record(d Demotion) bool {
	if _, seen := s.Notes[d.Slug]; seen {
		return false
	}
	d.At = time.Now().Format("2006-01-02")
	s.Notes[d.Slug] = d
	s.dirty = true
	return true
}

// Take returns and forgets a note's demotion — the restore half of rollback symmetry.
func (s *Store) Take(slug string) (Demotion, bool) {
	d, ok := s.Notes[slug]
	if ok {
		delete(s.Notes, slug)
		s.dirty = true
	}
	return d, ok
}

func (s *Store) Save() error {
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
	return os.WriteFile(filepath.Join(s.dir, "demotions.json"), b, 0o644)
}

// Log appends one line to .forge/drift.log. Demotion history lives here rather than in
// note bodies: §B.6 is explicit that a note's prose must not churn with every verdict.
func (s *Store) Log(line string) {
	if os.MkdirAll(s.dir, 0o755) != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(s.dir, "drift.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(time.Now().Format("2006-01-02T15:04:05") + " " + line + "\n") //nolint:errcheck // drift.log is best-effort history; a failed append must never change a verdict
}
