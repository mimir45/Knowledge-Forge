package sentinel

// Upsert creates the file (and its parent directories) if absent, replaces an existing
// id block's body in place, or appends a new block at end of file.
func Upsert(path, id string, style Style, body string) error {
	return upsert(path, id, style, body, appendBlock)
}

// UpsertBefore is Upsert but anchors a *newly created* block immediately before the
// given 1-based source line.
func UpsertBefore(path, id string, style Style, body string, anchorLine int) error {
	insert := func(lines, block []string) []string { return insertBefore(lines, block, anchorLine) }
	return upsert(path, id, style, body, insert)
}

func upsert(path, id string, style Style, body string, insert func(lines, block []string) []string) error {
	orig, err := readOrEmpty(path)
	if err != nil {
		return err
	}
	lines := splitLines(orig)
	block := renderBlock(style, id, body)
	if begin, end, found := findBlock(lines, style, id); found {
		return writeFile(path, replaceBlock(lines, begin, end, block))
	}
	return writeFile(path, insert(lines, block))
}

// appendBlock is Upsert's insert strategy: a fresh block goes at end of file, separated
// from any existing content by one blank line.
func appendBlock(lines, block []string) []string {
	if len(lines) == 0 {
		return block
	}
	out := append([]string{}, lines...)
	if out[len(out)-1] != "" {
		out = append(out, "")
	}
	return append(out, block...)
}

// insertBefore is UpsertBefore's insert strategy. anchorLine is 1-based and clamped, so an
// out-of-range hint degrades to inserting at the nearest valid end rather than erroring.
func insertBefore(lines, block []string, anchorLine int) []string {
	i := clamp(anchorLine-1, 0, len(lines))
	out := append([]string{}, lines[:i]...)
	out = append(out, block...)
	return append(out, lines[i:]...)
}

func clamp(i, lo, hi int) int {
	if i < lo {
		return lo
	}
	if i > hi {
		return hi
	}
	return i
}
