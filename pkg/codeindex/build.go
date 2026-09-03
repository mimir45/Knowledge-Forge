package codeindex

import (
	"runtime"
	"sync"
)

// Build parses every supported file in the list at one git revision.
func Build(repo, root, rev string, files []string) (Index, error) {
	ix := Index{Repo: repo, Commit: rev, Extractor: Extractor,
		Files: map[string]File{}, Deps: Deps(root)}
	if !Available() {
		return ix, ErrUnavailable
	}
	blobs, errc := readBlobs(root, rev, supported(files))
	parseAll(&ix, blobs)
	return ix, <-errc
}

func parseAll(ix *Index, blobs <-chan blob) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Go(func() {
			for b := range blobs {
				f, err := Parse(b.path, b.src)
				if err != nil || len(f.Symbols) == 0 {
					continue
				}
				mu.Lock()
				ix.Files[b.path] = f
				mu.Unlock()
			}
		})
	}
	wg.Wait()
}

func supported(files []string) []string {
	out := make([]string, 0, len(files))
	for _, p := range files {
		if Lang(p) != "" {
			out = append(out, p)
		}
	}
	return out
}
