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

// exportLineCap bounds one JSONL line on the export path. D3 pairs carry two whole note
// bodies and D2/D4 carry drafts and critiques, so the ceiling is generous; what matters
// is that exceeding it is reported rather than mistaken for end-of-file, which is what
// bufio.Scanner does by default and what all three existing readers in this tree do.
const exportLineCap = 16 << 20

// readStrict decodes every line of a tier's JSONL file into T, refusing anything it
// cannot parse. It is deliberately the opposite of the readers already in the tree —
// jsonl.go's existingKeys, cmd/forge's check_asklog.go and session_capture.go all skip a
// bad line, and jsonl.go says why: "a truncated tail must not wedge every future commit".
//
// Export inverts that on purpose. A line that does not parse is a line no anonymizer
// looked at, and quietly dropping it produces an export that is short by an unknown
// amount and, worse, looks complete. The cost is real: one crash-torn line makes export
// fail until someone edits the file. That is why the error names the file and the line
// number — it turns a permanent failure into a thirty-second fix.
//
// A missing file is not an error. A tier that never captured is an empty export, which
// the datasheet then reports as zero records.
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

// scanErr separates the too-long case from every other read failure. Left as a bare
// wrapped error it reads as corruption; named, it points at the one thing a caller can
// act on — a record larger than the export limit.
func scanErr(err error, name string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, bufio.ErrTooLong) {
		return fmt.Errorf("%s: a record exceeds the %d-byte export line limit", name, exportLineCap)
	}
	return fmt.Errorf("%s: %w", name, err)
}
