package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Append writes one JSON line to <vaultDir>/.forge/log.jsonl, creating the directory on
// first use. Pure I/O — callers gate this on cfg.Telemetry.Enabled themselves, the same
// posture engine_run.go's captureD2 takes toward pkg/dataset's own config check.
func Append(vaultDir string, ev Event) error {
	dir := filepath.Join(vaultDir, ".forge")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "log.jsonl"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	return err
}
