package dataset

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// exportLineCap bounds one JSONL line on the export path.
const exportLineCap = 16 << 20

// readStrict decodes every line of a tier's JSONL file into T.
func readStrict[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return scanStrict[T](f, filepath.Base(path))
}

func scanStrict[T any](r io.Reader, name string) ([]T, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), exportLineCap)
	var out []T
	for n := 1; sc.Scan(); n++ {
		var rec T
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", name, n, err)
		}
		out = append(out, rec)
	}
	return out, scanErr(sc.Err(), name)
}

// scanErr separates the too-long case from every other read failure.
func scanErr(err error, name string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, bufio.ErrTooLong) {
		return fmt.Errorf("%s: a record exceeds the %d-byte export line limit", name, exportLineCap)
	}
	return fmt.Errorf("%s: %w", name, err)
}
