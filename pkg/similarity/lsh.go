package similarity

import (
	"encoding/binary"
	"hash/fnv"
	"sort"
)

// Bands and Rows partition the signature for locality-sensitive hashing.
const (
	Bands = 64
	Rows  = Hashes / Bands
)

// Doc is one document in the index.
type Doc struct {
	ID    string
	Group string
	sig   Signature
}

// Pair is two documents found similar, with the estimated Jaccard between them.
type Pair struct {
	A, B  string
	Score float64
}

// Index holds signatures and their band buckets.
type Index struct {
	docs    []Doc
	buckets []map[uint64][]int // one map per band
}

// NewIndex returns an empty index.
func NewIndex() *Index {
	ix := &Index{buckets: make([]map[uint64][]int, Bands)}
	for i := range ix.buckets {
		ix.buckets[i] = map[uint64][]int{}
	}
	return ix
}

// Add indexes one document. group scopes comparison: only documents sharing a group are
// ever paired.
func (ix *Index) Add(id, group, text string) {
	if len(Shingle(text)) == 0 {
		return
	}
	n := len(ix.docs)
	ix.docs = append(ix.docs, Doc{ID: id, Group: group, sig: Sign(text)})
	for b := range ix.buckets {
		key := bandKey(ix.docs[n].sig, b)
		ix.buckets[b][key] = append(ix.buckets[b][key], n)
	}
}

// bandKey hashes one band of a signature into a bucket key.
func bandKey(sig Signature, band int) uint64 {
	h := fnv.New64a()
	var buf [8]byte
	for r := 0; r < Rows; r++ {
		binary.LittleEndian.PutUint64(buf[:], sig[band*Rows+r])
		h.Write(buf[:])
	}
	return h.Sum64()
}

// Candidates reports how many distinct pairs banding nominated for exact comparison.
func (ix *Index) Candidates() int { return len(ix.candidateList()) }

// Pairs returns every document pair estimated at or above threshold, most similar first.
// Ties break on ID so the report is stable across runs.
func (ix *Index) Pairs(threshold float64) []Pair {
	out := make([]Pair, 0, 16)
	for _, c := range ix.candidateList() {
		a, b := ix.docs[c[0]], ix.docs[c[1]]
		if a.Group != b.Group {
			continue
		}
		if s := Estimate(a.sig, b.sig); s >= threshold {
			out = append(out, Pair{A: a.ID, B: b.ID, Score: s})
		}
	}
	sortPairs(out)
	return out
}

func sortPairs(p []Pair) {
	sort.Slice(p, func(i, j int) bool {
		if p[i].Score != p[j].Score {
			return p[i].Score > p[j].Score
		}
		if p[i].A != p[j].A {
			return p[i].A < p[j].A
		}
		return p[i].B < p[j].B
	})
}

// candidateList returns each colliding index pair once. A pair that collides in several
// bands must be scored once, not once per band.
func (ix *Index) candidateList() [][2]int {
	seen := map[[2]int]bool{}
	var out [][2]int
	for _, bucket := range ix.buckets {
		for _, ids := range bucket {
			collect(ids, seen, &out)
		}
	}
	return out
}

func collect(ids []int, seen map[[2]int]bool, out *[][2]int) {
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			key := [2]int{min(ids[i], ids[j]), max(ids[i], ids[j])}
			if !seen[key] {
				seen[key] = true
				*out = append(*out, key)
			}
		}
	}
}
