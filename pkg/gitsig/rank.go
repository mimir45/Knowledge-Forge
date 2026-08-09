package gitsig

import "sort"

// FileCount is a file and how many commits touched it.
type FileCount struct {
	File  string
	Count int
}

// Couple is two files and how many commits touched both.
type Couple struct {
	A, B  string
	Count int
}

// TopChurn returns the most-committed files, most first. Ties break on path so a report
// generated twice from one history is byte-identical.
func TopChurn(s *Stats, n int) []FileCount {
	out := make([]FileCount, 0, len(s.Churn))
	for f, c := range s.Churn {
		out = append(out, FileCount{File: f, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].File < out[j].File
	})
	return truncate(out, n)
}

// TopCoupled returns the file pairs that most often change together, requiring at least min
// co-occurrences. Two files that changed together once are a coincidence, not a coupling.
func TopCoupled(s *Stats, min, n int) []Couple {
	out := make([]Couple, 0, 16)
	for p, c := range s.Coupled {
		if c >= min {
			out = append(out, Couple{A: p[0], B: p[1], Count: c})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		if out[i].A != out[j].A {
			return out[i].A < out[j].A
		}
		return out[i].B < out[j].B
	})
	return truncate(out, n)
}

func truncate[T any](s []T, n int) []T {
	if n > 0 && len(s) > n {
		return s[:n]
	}
	return s
}
