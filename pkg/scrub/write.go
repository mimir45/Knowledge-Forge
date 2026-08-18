package scrub

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeAll commits every scrubbed note to disk. Called only after scrubAll has
// succeeded for the whole run, so a partial write never happens.
func writeAll(dstDir string, files []file) error {
	for _, f := range files {
		if err := writeOne(dstDir, f); err != nil {
			return err
		}
	}
	return nil
}

func writeOne(dstDir string, f file) error {
	dst := filepath.Join(dstDir, filepath.FromSlash(f.rel))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("scrub: mkdir %s: %w", filepath.Dir(dst), err)
	}
	if err := os.WriteFile(dst, f.data, 0o644); err != nil {
		return fmt.Errorf("scrub: write %s: %w", dst, err)
	}
	return nil
}
