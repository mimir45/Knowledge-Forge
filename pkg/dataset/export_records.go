package dataset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

// loadTier reads a tier's file into its own struct type and boxes the result. The type
// switch is what makes the strict reader's generic parameter concrete; a tier added to
// Tiers() without a case here fails loudly at export rather than exporting nothing.
func loadTier(root string, t Tier) ([]any, error) {
	path := filepath.Join(root, t.Path)
	switch t.Tag {
	case D1Tag:
		return loadD1(root, path)
	case D2Tag:
		return boxed(readStrict[D2Pair](path))
	case D3Tag:
		return boxed(readStrict[Pair](path))
	case D4Tag:
		return boxed(readStrict[D4Pair](path))
	case D5Tag:
		return boxed(readStrict[D5Pair](path))
	}
	return nil, fmt.Errorf("tier %q has no reader", t.Tag)
}

// loadD1 reads D1's own pairs plus the separate outcome file BACKLOG B-035 added
// (d1_outcome.go) and joins them by RunID before boxing. The join happens here, not in
// render, so `since`, anonymizeAll and roundTripAll all see the already-joined shape —
// Outcome is an export-time-only field and never round-trips back into the capture file.
func loadD1(root, path string) ([]any, error) {
	pairs, err := readStrict[D1Pair](path)
	if err != nil {
		return nil, err
	}
	outs, err := readStrict[D1Outcome](filepath.Join(root, D1OutcomePath))
	if err != nil {
		return nil, err
	}
	return boxed(joinD1Outcomes(pairs, outs), nil)
}

// joinD1Outcomes matches each pair to the outcome sharing its RunID, if any. A pair with
// no RunID (captured before B-035) and a pair whose gate call never received --run-id
// both stay unjoined — Outcome is left nil, which the renderers read as "no outcome
// recorded", not "not published".
//
// Two outcomes can legitimately share a RunID: a quarantine followed by a fixed retry
// that passes --run-id back through the same --previous-draft repair loop (see
// SKILL.md's Stage 4). The map assignment below is last-wins in append order — outs is
// read straight off the JSONL file, so "last" means "most recently appended" — which is
// deliberately the retry's outcome, not the original quarantine's. That is the final
// disposition of the routing decision; recording "quarantined" forever because the first
// gate call happened to fail would bias the label in exactly the direction a repair loop
// exists to fix.
func joinD1Outcomes(pairs []D1Pair, outs []D1Outcome) []D1Pair {
	if len(outs) == 0 {
		return pairs
	}
	published := make(map[string]bool, len(outs))
	for _, o := range outs {
		published[o.RunID] = o.Published // last write for a RunID wins, see doc comment
	}
	for i := range pairs {
		if pairs[i].RunID == "" {
			continue
		}
		if pub, ok := published[pairs[i].RunID]; ok {
			v := pub
			pairs[i].Outcome = &v
		}
	}
	return pairs
}

func boxed[T any](recs []T, err error) ([]any, error) {
	if err != nil {
		return nil, err
	}
	out := make([]any, len(recs))
	for i, r := range recs {
		out[i] = r
	}
	return out, nil
}

// stampOf is the field --since filters on. D3 uses EditedAt rather than a capture time:
// the pair is dated by when the human made the correction, which is the event, not by
// when the hook happened to harvest it.
func stampOf(rec any) time.Time {
	switch p := rec.(type) {
	case D1Pair:
		return p.CapturedAt
	case D2Pair:
		return p.CapturedAt
	case Pair:
		return p.EditedAt
	case D4Pair:
		return p.CapturedAt
	case D5Pair:
		return p.CapturedAt
	}
	return time.Time{}
}

// roundTripAll is the second of the two fail-closed layers, and it is worth being precise
// about what it proves. Re-decoding every redacted record into its own struct and getting
// byte-identical JSON back proves the redaction did not corrupt the record's shape — that
// no replacement landed inside a JSON escape and turned a field into something that
// silently parses as another. It proves nothing about whether a secret escaped; the
// seeded-fixture test is the only thing that does, which is why that test is the D-6
// regression guard rather than this check.
func roundTripAll(recs []any) error {
	for i, r := range recs {
		if err := roundTrip(r); err != nil {
			return fmt.Errorf("record %d failed re-validation after redaction: %w", i+1, err)
		}
	}
	return nil
}

func roundTrip(rec any) error {
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	switch rec.(type) {
	case D1Pair:
		return reparse[D1Pair](b)
	case D2Pair:
		return reparse[D2Pair](b)
	case Pair:
		return reparse[Pair](b)
	case D4Pair:
		return reparse[D4Pair](b)
	case D5Pair:
		return reparse[D5Pair](b)
	}
	return fmt.Errorf("unhandled record type %T", rec)
}

func reparse[T any](b []byte) error {
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	again, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if !bytes.Equal(b, again) {
		return fmt.Errorf("re-encoding changed the record:\n  before %s\n  after  %s", b, again)
	}
	return nil
}
