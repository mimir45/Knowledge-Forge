package drift

import (
	"math/rand"
	"sort"
	"testing"

	"knowledge-forge/pkg/codeindex"
)

// TestNameMapOrdersDeclarationsWithinOneFile is the bug drift.md showed on the real vault:
// its verdict count flipped between 9 and 10 notes across runs on an unchanged tree. One
// Java file declares both `Order.Builder` and `OrderItem.Builder`, the short-name hits tied
// on (repo, path), and sort.Slice settled the tie by map iteration order. Find must pick
// the same declaration every time or a verdict is not a function of the tree.
func TestNameMapOrdersDeclarationsWithinOneFile(t *testing.T) {
	want := []loc{
		{"a", "src/Order.java", codeindex.Symbol{Name: "Order.Builder", Start: 10}},
		{"a", "src/Order.java", codeindex.Symbol{Name: "OrderItem.Builder", Start: 90}},
		{"b", "src/Other.java", codeindex.Symbol{Name: "Other.Builder", Start: 5}},
	}
	for run := 0; run < 50; run++ {
		ls := append([]loc(nil), want...)
		rand.Shuffle(len(ls), func(i, j int) { ls[i], ls[j] = ls[j], ls[i] })
		m := &nameMap{full: map[string][]loc{}, short: map[string][]loc{"Builder": ls}}
		m.sort()
		if got := m.short["Builder"]; !sameOrder(got, want) {
			t.Fatalf("run %d: %+v, want %+v", run, got, want)
		}
	}
}

func sameOrder(got, want []loc) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// TestLessLocIsATotalOrder: a comparator that reports two distinct declarations equal is
// exactly what let an unstable sort choose between them.
func TestLessLocIsATotalOrder(t *testing.T) {
	ls := []loc{
		{"a", "p.java", codeindex.Symbol{Name: "X", Start: 1}},
		{"a", "p.java", codeindex.Symbol{Name: "Y", Start: 1}},
		{"a", "p.java", codeindex.Symbol{Name: "Z", Start: 2}},
	}
	sort.Slice(ls, func(i, j int) bool { return lessLoc(ls[i], ls[j]) })
	for i := 0; i+1 < len(ls); i++ {
		if !lessLoc(ls[i], ls[i+1]) {
			t.Errorf("lessLoc reports %+v and %+v equal", ls[i], ls[i+1])
		}
	}
}
