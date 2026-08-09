package store

import (
	"database/sql"
	"time"
)

// budgetSchemaSQL lives apart from schemaSQL so the one table Reset() must never touch
// stays a one-file diff (AUDIT §8.4 D-8) rather than a line buried in store.go's list.
const budgetSchemaSQL = `
CREATE TABLE IF NOT EXISTS budget (
  day  TEXT NOT NULL,
  tier TEXT NOT NULL,
  usd  REAL NOT NULL DEFAULT 0,
  PRIMARY KEY (day, tier)
);
`

func ensureBudgetSchema(db *sql.DB) error {
	_, err := db.Exec(budgetSchemaSQL)
	return err
}

// dayKey is the UTC calendar day the budget table keys spend by. A day boundary needs no
// reset job: it is just a primary-key value that does not exist yet.
func dayKey(clock func() time.Time) string {
	return clock().UTC().Format("2006-01-02")
}

// Spend books usd against tier for today (per clock), upserting onto any prior spend the
// same day. clock is injected so a test can pin a day rather than racing time.Now().
func (s *Store) Spend(tier string, usd float64, clock func() time.Time) error {
	_, err := s.DB.Exec(`INSERT INTO budget (day, tier, usd) VALUES (?, ?, ?)
		ON CONFLICT(day, tier) DO UPDATE SET usd = usd + excluded.usd`,
		dayKey(clock), tier, usd)
	return err
}

// Remaining is capUSD minus everything already spent on tier today. A tier with no rows
// yet has spent 0, so Remaining returns the full cap — no seeding required.
func (s *Store) Remaining(tier string, capUSD float64, clock func() time.Time) (float64, error) {
	var spent float64
	err := s.DB.QueryRow(`SELECT COALESCE(SUM(usd), 0) FROM budget WHERE day = ? AND tier = ?`,
		dayKey(clock), tier).Scan(&spent)
	if err != nil {
		return 0, err
	}
	return capUSD - spent, nil
}
