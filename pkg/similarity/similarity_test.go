package similarity

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const noteA = `Spring Boot resolves a bean by type first and by name only when the type is
ambiguous. Marking one candidate @Primary settles the ambiguity without a qualifier, and a
@Qualifier on the injection point overrides @Primary when both are present.`

// The near-duplicate: same claims, a few words moved. This is the shape the vault actually
// produces — two notes written months apart about one behaviour — not a copy-paste.
const noteBNear = `Spring Boot resolves a bean by type first and by name only when the type
is ambiguous. Marking a candidate @Primary settles the ambiguity without a qualifier, and a
@Qualifier at the injection point overrides @Primary when both are present.`

const noteC = `Testcontainers starts a real Postgres in Docker for the test run. Reuse is off
by default, so each class pays container startup unless withReuse(true) is set and
testcontainers.reuse.enable=true is present in ~/.testcontainers.properties.`

func TestEstimateTracksJaccard(t *testing.T) {
	for _, c := range []struct {
		name     string
		a, b     string
		min, max float64
	}{
		{"identical", noteA, noteA, 1.0, 1.0},
		// Two moved words out of ~45 distinct ones, so at w=1 these are nearly the same bag.
		{"near duplicate", noteA, noteBNear, 0.85, 1.0},
		// Not zero: prose shares "a", "the", "is". The separation that matters is this gap,
		// and it is wide.
		{"unrelated", noteA, noteC, 0.0, 0.20},
	} {
		got := Estimate(Sign(c.a), Sign(c.b))
		if got < c.min || got > c.max {
			t.Errorf("%s: estimate = %.3f, want in [%.2f, %.2f]", c.name, got, c.min, c.max)
		}
	}
}

// The estimate is a sketch, so it is only useful if it stays near the truth. Compare it to
// the exact Jaccard over the same shingle sets.
func TestEstimateApproximatesExactJaccard(t *testing.T) {
	for _, c := range [][2]string{{noteA, noteBNear}, {noteA, noteC}, {noteBNear, noteC}} {
		exact, est := exactJaccard(c[0], c[1]), Estimate(Sign(c[0]), Sign(c[1]))
		if math.Abs(exact-est) > 0.12 { // ~1.3 standard errors at 128 hashes
			t.Errorf("estimate %.3f vs exact %.3f", est, exact)
		}
	}
}

func TestPairsFindsNearDuplicateAndNotTheRest(t *testing.T) {
	ix := NewIndex()
	ix.Add("a.md", "concept", noteA)
	ix.Add("b.md", "concept", noteBNear)
	ix.Add("c.md", "concept", noteC)
	pairs := ix.Pairs(DuplicateThreshold)
	if len(pairs) != 1 {
		t.Fatalf("pairs = %+v, want exactly the a/b pair", pairs)
	}
	if pairs[0].A != "a.md" || pairs[0].B != "b.md" {
		t.Errorf("pair = %+v, want a.md/b.md", pairs[0])
	}
}

// Cross-type pairs must never be scored, however similar the text. In the real vault the
// top-scoring pairs are all cross-type and none of them is a duplicate: a decision note and
// the pitfall note that caused it share almost all their vocabulary by design.
func TestGroupsAreNeverComparedAcrossTypes(t *testing.T) {
	ix := NewIndex()
	ix.Add("a.md", "concept", noteA)
	ix.Add("b.md", "decision", noteA) // identical text, different type
	if pairs := ix.Pairs(DuplicateThreshold); len(pairs) != 0 {
		t.Errorf("pairs = %+v, want none across types", pairs)
	}
}

// The end-to-end acceptance criterion: the fixture vault's deliberate F7 near-duplicate must
// appear in the report. It is the pair testdata/README.md exists to make the duplicate gate
// catch, and it is the pair two earlier band tunings silently dropped — Pairs returned
// nothing while Estimate agreed it was a duplicate.
func TestFixtureNearDuplicateIsNominated(t *testing.T) {
	ix := NewIndex()
	dir := filepath.Join("..", "..", "testdata", "vault", "concepts")
	for _, name := range []string{"soft-delete.md", "soft-deletion.md", "hibernate.md"} {
		ix.Add(name, "concepts", bodyOf(t, filepath.Join(dir, name)))
	}
	pairs := ix.Pairs(DuplicateThreshold)
	if len(pairs) != 1 {
		t.Fatalf("pairs = %+v, want exactly the F7 pair", pairs)
	}
	if pairs[0].A != "soft-delete.md" || pairs[0].B != "soft-deletion.md" {
		t.Errorf("pair = %+v, want the soft-delete pair", pairs[0])
	}
}

// bodyOf strips frontmatter. Indexing whole files makes same-type notes look alike through
// their shared schema keys: on the real vault it inflated the candidate set from 1 pair to 7.
func bodyOf(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if rest, ok := strings.CutPrefix(s, "---"); ok {
		if i := strings.Index(rest, "\n---"); i >= 0 {
			return rest[i+4:]
		}
	}
	return s
}

// Two notes with no shingles must not come out as a perfect duplicate pair. An empty
// signature is all-maxima, so a naive index would score them 1.0 — the report's most
// embarrassing possible false positive, and the fixture vault has notes this short.
func TestEmptyDocumentsAreNotDuplicates(t *testing.T) {
	ix := NewIndex()
	ix.Add("empty-1.md", "concept", "")
	ix.Add("empty-2.md", "concept", "   \n\n")
	if pairs := ix.Pairs(DuplicateThreshold); len(pairs) != 0 {
		t.Errorf("pairs = %+v, want none", pairs)
	}
}

// Determinism is a T0 invariant: the same vault must produce the same report on every
// machine, so nothing here may seed from a random source or map iteration order.
func TestSignatureIsDeterministic(t *testing.T) {
	if Sign(noteA) != Sign(noteA) {
		t.Error("two signings of the same text differed")
	}
}

// A pair colliding in several bands must be scored once, not once per band.
func TestIdenticalDocumentsYieldOnePair(t *testing.T) {
	ix := NewIndex()
	ix.Add("a.md", "concept", noteA)
	ix.Add("copy.md", "concept", noteA)
	if pairs := ix.Pairs(DuplicateThreshold); len(pairs) != 1 {
		t.Fatalf("pairs = %+v, want one", pairs)
	}
}

// At ShingleWords=1 a document is a bag of words: order carries no signal, which is what the
// F7 measurement showed and what makes punctuation stripping matter.
func TestShingleIsABagOfWords(t *testing.T) {
	if got := Shingle("Order, order!  ORDER"); len(got) != 3 || got[0] != "order" {
		t.Errorf("shingles = %q, want three lowercased word tokens", got)
	}
	if got := Shingle("  \n\t "); got != nil {
		t.Errorf("shingles = %q, want nil for text with no words", got)
	}
}

func exactJaccard(a, b string) float64 {
	sa, sb := set(Shingle(a)), set(Shingle(b))
	inter := 0
	for s := range sa {
		if sb[s] {
			inter++
		}
	}
	union := len(sa) + len(sb) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func set(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}

func BenchmarkSign(b *testing.B) {
	text := strings.Repeat(noteA+" ", 4) // ~a real note's length
	b.ReportAllocs()
	for b.Loop() {
		Sign(text)
	}
}

// Worst case, deliberately: 500 notes that are all ~0.95 similar to each other, so banding
// nominates every one of the 124750 pairs and no pruning is possible. This does not show LSH
// beating O(n^2) — a corpus where every pair really is a duplicate cannot show that — it
// bounds the price when banding prunes nothing. 500 notes is five times the real vault.
func BenchmarkPairs500Notes(b *testing.B) {
	ix := NewIndex()
	for i := 0; i < 500; i++ {
		ix.Add(fmt.Sprintf("n%03d.md", i), "concept", fmt.Sprintf("%s variation %d", noteA, i))
	}
	b.ResetTimer()
	for b.Loop() {
		ix.Pairs(DuplicateThreshold)
	}
}
