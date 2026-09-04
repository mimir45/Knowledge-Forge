package gitsig

import "sort"

// CouplingCap is the largest commit that contributes to co-change coupling.
const CouplingCap = 25

// Stats are the per-file signals derived from a commit range.
type Stats struct {
	Churn   map[string]int            // commits touching the file
	Authors map[string]map[string]int // file -> author -> commits
	Coupled map[[2]string]int         // file pair -> commits touching both
	Commits int
}

// Analyze folds a commit list into per-file signals.
func Analyze(commits []Commit) *Stats {
	s := &Stats{
		Churn:   map[string]int{},
		Authors: map[string]map[string]int{},
		Coupled: map[[2]string]int{},
		Commits: len(commits),
	}
	for _, c := range commits {
		s.countFiles(c)
		s.countPairs(c.Files)
	}
	return s
}

func (s *Stats) countFiles(c Commit) {
	for _, f := range c.Files {
		s.Churn[f]++
		if s.Authors[f] == nil {
			s.Authors[f] = map[string]int{}
		}
		s.Authors[f][c.Author]++
	}
}

func (s *Stats) countPairs(files []string) {
	if len(files) > CouplingCap {
		return
	}
	sorted := append([]string(nil), files...)
	sort.Strings(sorted)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			s.Coupled[[2]string{sorted[i], sorted[j]}]++
		}
	}
}

// Owner returns the author of the most commits touching a file and their share of them.
func (s *Stats) Owner(file string) (string, float64) {
	by := s.Authors[file]
	if len(by) == 0 {
		return "", 0
	}
	top, total := "", 0
	for a, n := range by {
		total += n
		if n > by[top] || (n == by[top] && a < top) {
			top = a
		}
	}
	return top, float64(by[top]) / float64(total)
}
