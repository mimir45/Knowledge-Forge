package dataset

import (
	"strings"
	"testing"
)

// TestScanStrictRejectsAnOverlongLine is the case bufio.Scanner reports as a plain
// error and every tolerant reader in this tree silently treats as end-of-file.
func TestScanStrictRejectsAnOverlongLine(t *testing.T) {
	huge := `{"kind":"d2","draft":"` + strings.Repeat("x", exportLineCap+1) + `"}`
	_, err := scanStrict[D2Pair](strings.NewReader(huge), "d2.jsonl")
	if err == nil {
		t.Fatal("scanStrict accepted a line past the buffer cap, want an error")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("error should name the cause, got: %v", err)
	}
}

// TestScanStrictNamesTheBadLine: the reader deliberately refuses what jsonl.go tolerates,
// so it owes the user the one thing that makes the failure fixable by hand.
func TestScanStrictNamesTheBadLine(t *testing.T) {
	in := "{\"kind\":\"d2\"}\n{\"kind\":\"d2\"}\nnot json\n{\"kind\":\"d2\"}\n"
	_, err := scanStrict[D2Pair](strings.NewReader(in), "d2.jsonl")
	if err == nil || !strings.Contains(err.Error(), "d2.jsonl:3") {
		t.Fatalf("got %v, want an error naming d2.jsonl:3", err)
	}
}

func TestScanStrictReadsEveryGoodLine(t *testing.T) {
	in := "{\"kind\":\"d2\",\"draft\":\"a\"}\n{\"kind\":\"d2\",\"draft\":\"b\"}\n"
	recs, err := scanStrict[D2Pair](strings.NewReader(in), "d2.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 || recs[1].Draft != "b" {
		t.Errorf("got %+v, want two records ending in b", recs)
	}
}

// TestReadStrictTreatsAMissingFileAsEmpty: a tier that has never captured is a zero-row
// export, not a failure — dataset-stats depends on this to report all five tiers.
func TestReadStrictTreatsAMissingFileAsEmpty(t *testing.T) {
	recs, err := readStrict[D2Pair](t.TempDir() + "/nothing.jsonl")
	if err != nil || recs != nil {
		t.Errorf("got %v, %v; want nil, nil", recs, err)
	}
}
