// Package similarity finds near-duplicate notes by MinHash over word shingles, with LSH
// banding so the vault is not compared pair-by-pair.
package similarity

import (
	"hash/fnv"
	"strings"
	"unicode"
)

// Hashes is the signature length. 128 puts the standard error of the Jaccard estimate near
// 1/sqrt(128) ≈ 9%, which is fine for a report that ranks candidates a human then reads.
const Hashes = 128

// ShingleWords is the shingle width in words.
const ShingleWords = 1

// DuplicateThreshold is the score at or above which a pair is worth a human's
// attention.
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

// permute applies the i-th hash of the family h -> a*h + b over the 64-bit ring.
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
