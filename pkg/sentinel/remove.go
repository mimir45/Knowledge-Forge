package sentinel

import "os"

// Remove strips a block by id, plus the one blank separator line Upsert's appendBlock
// inserted ahead of it when the block sits at end of file.
func Remove(path, id string, style Style) error {
	orig, err := readOrEmpty(path)
	if err != nil {
		return nil
	}
	lines := splitLines(orig)
	begin, end, found := findBlock(lines, style, id)
	if !found {
		return nil
	}
	return writeFile(path, removeBlock(lines, begin, end))
}

func removeBlock(lines []string, begin, end int) []string {
	if end == len(lines)-1 && begin > 0 && lines[begin-1] == "" {
		begin--
	}
	out := append([]string{}, lines[:begin]...)
	return append(out, lines[end+1:]...)
}

func readOrEmpty(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return b, err
}
