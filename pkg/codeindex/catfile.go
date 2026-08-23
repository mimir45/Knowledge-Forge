package codeindex

import (
	"bufio"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type blob struct {
	path string
	src  []byte
}

// readBlobs streams the contents of many files at one git revision through a single
// `git cat-file --batch` process. Reading from the object store rather than the working
// tree is what makes an index a pure function of tree state: an unstaged edit must not
// move a verdict, and a historical revision has no working tree to read at all. One
// process rather than one per file is the difference between milliseconds and seconds.
func readBlobs(root, rev string, files []string) (<-chan blob, <-chan error) {
	out := make(chan blob)
	errc := make(chan error, 1)
	go func() {
		defer close(out)
		errc <- streamBlobs(root, rev, files, out)
	}()
	return out, errc
}

func streamBlobs(root, rev string, files []string, out chan<- blob) error {
	cmd := exec.Command("git", "cat-file", "--batch")
	cmd.Dir = root
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go feedRequests(in, rev, files)
	err = drainBlobs(bufio.NewReaderSize(pipe, 1<<16), files, out)
	// The exit status only adds information when every reply arrived: a truncated stream
	// already surfaces as drainBlobs' read error, and Wait would bury it under a broken
	// pipe. What this catches is the narrow case of a full transcript then a non-zero exit.
	if werr := cmd.Wait(); err == nil {
		err = werr
	}
	return err
}

func feedRequests(in io.WriteCloser, rev string, files []string) {
	defer in.Close()
	w := bufio.NewWriter(in)
	for _, p := range files {
		w.WriteString(rev + ":" + p + "\n") //nolint:errcheck // this goroutine has nowhere to report; a failed write ends the reply stream and surfaces as a short read in drainBlobs
	}
	w.Flush()
}

// drainBlobs walks the replies in request order — git guarantees one per line of input,
// so the requested path is the key, not anything in the reply.
func drainBlobs(r *bufio.Reader, files []string, out chan<- blob) error {
	for _, p := range files {
		hdr, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		n, ok := blobSize(hdr)
		if !ok {
			continue // "<name> missing": the path does not exist at this revision
		}
		src := make([]byte, n+1) // git appends a newline after the object
		if _, err := io.ReadFull(r, src); err != nil {
			return err
		}
		out <- blob{path: p, src: src[:n]}
	}
	return nil
}

func blobSize(hdr string) (int, bool) {
	f := strings.Fields(hdr)
	if len(f) != 3 || f[1] != "blob" {
		return 0, false
	}
	n, err := strconv.Atoi(f[2])
	return n, err == nil
}
