package vault

import "testing"

// The classification of generated output is the one that fails silently. forge check
// writes nine reports and a MOC into the vault it just measured.
func TestGeneratedOutputIsNotACountedNote(t *testing.T) {
	for _, rel := range []string{"reports/duplicates.md", "reports/orphans.md"} {
		if IsContentNote(rel) {
			t.Errorf("IsContentNote(%q) = true; a report would count itself, and its wikilinks "+
				"would give its own subjects inbound links", rel)
		}
	}
}

// moc/ is a content note but never a contract note: `type` admits seven values and none
// is "moc".
func TestMOCIsAGraphNodeButNotContractBound(t *testing.T) {
	const rel = "moc/codebase.md"
	if !IsContentNote(rel) {
		t.Errorf("IsContentNote(%q) = false; a MOC's outbound links are what de-orphan the "+
			"notes it points at", rel)
	}
	if IsContractNote(rel) {
		t.Errorf("IsContractNote(%q) = true; there is no type: value it could satisfy", rel)
	}
}

// A weekly rollup is dated output, not a stable map: counting it in the next run's graph
// would move the very duplicate/orphan/drift counts it quotes inside itself.
func TestWeeklyRollupIsNotACountedNote(t *testing.T) {
	const rel = "moc/weekly/2026-W33.md"
	if IsContentNote(rel) {
		t.Errorf("IsContentNote(%q) = true; a weekly rollup would count itself on the run "+
			"after it was written", rel)
	}
}

func TestOrdinaryNotesAreStillBothThings(t *testing.T) {
	const rel = "notes/concept/soft-delete.md"
	if !IsContentNote(rel) || !IsContractNote(rel) {
		t.Errorf("%q: content=%v contract=%v, want both true",
			rel, IsContentNote(rel), IsContractNote(rel))
	}
}
