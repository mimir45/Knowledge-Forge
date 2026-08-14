package sentinel

import "strings"

// splitLines treats a missing or empty file as zero lines, not one empty line — that is
// what keeps writeFile's round trip idempotent on a fresh file.
func splitLines(b []byte) []string {
	s := string(b)
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// findBlock locates a block by id, not by position: a marker survives edits elsewhere in
// the file, and a symbol whose source line has shifted since the last run is still found.
func findBlock(lines []string, style Style, id string) (begin, end int, ok bool) {
	b, e := style.begin(id), style.end(id)
	begin, end = -1, -1
	for i, l := range lines {
		switch {
		case l == b:
			begin = i
		case l == e && begin >= 0:
			end = i
		}
		if begin >= 0 && end >= 0 {
			break
		}
	}
	return begin, end, begin >= 0 && end >= 0
}

// renderBlock comments out the body for a line-comment Style (Close == "") — a marker
// dropped into a Java or Python file must stay a comment on every line, not just its
// begin/end pair, or Upsert would insert a line of uncommented prose into source code. A
// block-comment Style like Markdown wraps the body as-is; the open/close pair alone
// already comments the whole span.
func renderBlock(style Style, id, body string) []string {
	block := []string{style.begin(id)}
	for _, l := range splitLines([]byte(body)) {
		block = append(block, style.comment(l))
	}
	return append(block, style.end(id))
}

func (s Style) comment(line string) string {
	if s.Close != "" || line == "" {
		return line
	}
	return s.Open + " " + line
}

func replaceBlock(lines []string, begin, end int, block []string) []string {
	out := append([]string{}, lines[:begin]...)
	out = append(out, block...)
	return append(out, lines[end+1:]...)
}
