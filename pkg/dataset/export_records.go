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
		return boxed(readStrict[D1Pair](path))
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
