package codeindex

import (
	"runtime"
	"sync"
)

// Build parses every supported file in the list at one git revision. The worker pool is
// sized to GOMAXPROCS per STACK §7: parsing N source files is the one embarrassingly
// parallel hotspot in the profile, and it is the reason Go was chosen over Python here.
//
// The blobs come from the object store, never the working tree, so the index is a pure
// function of tree state — the same property drift's verdicts depend on.
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
