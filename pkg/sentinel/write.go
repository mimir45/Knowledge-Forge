package sentinel

import (
	"os"
	"path/filepath"
	"strings"
)

// writeFile skips the write when the rendered content already matches what's on disk —
// an identical rewrite still bumps mtime, the same reason cmd/forge's own writers compare
// before they write. That comparison is what makes Upsert idempotent on disk, not merely
// in content.
func writeFile(path string, lines []string) error {
	content := render(lines)
	if old, err := os.ReadFile(path); err == nil && string(old) == content {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func render(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
