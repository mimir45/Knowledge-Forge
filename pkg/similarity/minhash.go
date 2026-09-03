// Package similarity finds near-duplicate notes by MinHash over word shingles, with LSH
// banding so the vault is not compared pair-by-pair.
//
// It is hand-rolled and deliberately holds **no embeddings**: the whole T0
// core makes zero model calls, and a duplicate report that needed an embedding service
// would be the first thing to break that. Jaccard over shingles is also the honest measure
// here — two notes are duplicates when they say the same words, not when they are about
// adjacent topics, and cosine over embeddings cheerfully conflates the two.
//
// Everything is deterministic: the permutations are derived from a fixed seed, so the same
// vault produces the same report on every machine and a diff of two runs means the vault
// changed.
package similarity

import (
	"hash/fnv"
	"strings"
	"unicode"
)

// Hashes is the signature length. 128 puts the standard error of the Jaccard estimate near
// 1/sqrt(128) ≈ 9%, which is fine for a report that ranks candidates a human then reads.
const Hashes = 128

// ShingleWords is the shingle width in words. One — a bag of words — is measured, not
// conventional; five is the usual choice for prose and it is wrong for this corpus. The
// vault's duplicates are notes written months apart about one behaviour, systematically
// reworded, so word order carries no signal and every extra word of context costs overlap.
// Against the fixture's deliberate near-duplicate pair (F7, soft-delete.md/soft-deletion.md)
// and its nearest same-type non-duplicate, the margin shrinks monotonically with width:
// w=1 0.575/0.196, w=2 0.322/0.067, w=3 0.214/0.032, w=5 0.096/0.006. Wider shingles do
// separate copy-paste, which this vault does not contain.
const ShingleWords = 1

// DuplicateThreshold is the score at or above which a pair is worth a human's attention.
//
// It replaces the original spec's ">0.85 similar", which is a copy-paste detector: at 0.85 the
// real 91-note vault yields zero rows and the fixture's deliberate near-duplicate is never
// nominated at any shingle width. 0.40 sits four standard errors below F7's 0.575 (the
// sketch's standard error is ~1/sqrt(128) ≈ 0.09, and 0.044 at this score), so sampling
// noise cannot drop the pair the fixture exists to catch, and it yields three rows on the
// real vault against a same-type ceiling of 0.504.
//
// It is only meaningful on same-type pairs — see Index.Add.
const DuplicateThreshold = 0.40

// Signature is a MinHash sketch of one document.
type Signature [Hashes]uint64

// Shingle splits text into overlapping word n-grams, lowercased and stripped of
// punctuation so that formatting churn does not change the sketch.
func Shingle(text string) []string {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	if len(words) < ShingleWords {
		if len(words) == 0 {
			return nil
		}
		return []string{strings.Join(words, " ")} // a note too short to shingle is still itself
	}
	out := make([]string, 0, len(words)-ShingleWords+1)
	for i := 0; i+ShingleWords <= len(words); i++ {
		out = append(out, strings.Join(words[i:i+ShingleWords], " "))
	}
	return out
}

// Sign computes the MinHash signature of a document.
func Sign(text string) Signature {
	var sig Signature
	for i := range sig {
		sig[i] = ^uint64(0)
	}
	for _, sh := range Shingle(text) {
		h := hash64(sh)
		for i := range sig {
			if p := permute(h, i); p < sig[i] {
				sig[i] = p
			}
		}
	}
	return sig
}

// permute applies the i-th hash of the family h -> a*h + b over the 64-bit ring. The
// multipliers are odd, which makes each permutation a bijection, and derived from i rather
// than from a random source so runs are reproducible.
func permute(h uint64, i int) uint64 {
	a := uint64(i)*0x9E3779B97F4A7C15 | 1
	b := uint64(i)*0xC2B2AE3D27D4EB4F + 0x165667B19E3779F9
	return a*h + b
}

func hash64(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// Estimate returns the estimated Jaccard similarity of two documents: the share of
// signature positions on which their minima agree.
func Estimate(a, b Signature) float64 {
	same := 0
	for i := range a {
		if a[i] == b[i] {
			same++
		}
	}
	return float64(same) / float64(Hashes)
}
