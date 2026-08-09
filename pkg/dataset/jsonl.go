package dataset

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

// Append writes pairs the file does not already hold and returns how many it added.
// The dataset lives under .forge/, which is derived and gitignored in the vault, so it
// stays local — ADDENDUM §D's datasets are never transmitted anywhere.
func Append(path string, pairs []Pair) (int, error) {
	seen, err := existingKeys(path)
	if err != nil {
		return 0, err
	}
	fresh := make([]Pair, 0, len(pairs))
	for _, p := range pairs {
		if !seen[p.Key()] {
			seen[p.Key()] = true
			fresh = append(fresh, p)
		}
	}
	if len(fresh) == 0 {
		return 0, nil
	}
	return len(fresh), write(path, fresh)
}

func write(path string, pairs []Pair) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, p := range pairs {
		if err := encode(f, p); err != nil {
			return err
		}
	}
	return f.Sync()
}

func encode(f *os.File, p Pair) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

// existingKeys makes the capture idempotent. The hook can legitimately fire twice on one
// commit — `commit --amend`, a rebase, a manual re-run — and the file is append-only.
// A line that no longer parses is skipped rather than fatal: a truncated tail must not
// wedge every future commit.
func existingKeys(path string) (map[string]bool, error) {
	keys := map[string]bool{}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return keys, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		var p Pair
		if json.Unmarshal(sc.Bytes(), &p) == nil {
			keys[p.Key()] = true
		}
	}
	return keys, sc.Err()
}
