package dataset

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

// loadTier reads a tier's file into its own struct type and boxes the result.
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
	case D6Tag:
		return loadD6(root)
	}
	return nil, fmt.Errorf("tier %q has no reader", t.Tag)
}

// loadD1 reads D1's own pairs plus the separate outcome file.
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

// joinD1Outcomes matches each pair to the outcome sharing its RunID, if any.
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

// stampOf is the field --since filters on.
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

// roundTripAll is the second of the two fail-closed layers.
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
	case D6Pair:
		return reparse[D6Pair](b)
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
