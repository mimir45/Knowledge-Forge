package engine

import "time"

// Ledger is the budget store this package needs — structurally matched by *store.Store
// so pkg/engine never imports pkg/store and stays a leaf. Spend/Remaining take an
// injected clock so a test can pin a day rather than racing time.Now().
type Ledger interface {
	Spend(tier string, usd float64, clock func() time.Time) error
	Remaining(tier string, capUSD float64, clock func() time.Time) (float64, error)
}
