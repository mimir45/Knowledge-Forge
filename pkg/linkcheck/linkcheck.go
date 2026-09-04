// Package linkcheck asks whether a note's sources still resolve.
package linkcheck

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Verdict is what a check concluded.
type Verdict string

const (
	Alive       Verdict = "alive"       // 2xx or 3xx
	Dead        Verdict = "dead"        // 4xx or 5xx: the server answered, and said no
	Unreachable Verdict = "unreachable" // DNS, TLS, timeout: we cannot tell, so we do not say
)

// Status is one URL's result.
type Status struct {
	URL     string    `json:"url"`
	Verdict Verdict   `json:"verdict"`
	Code    int       `json:"code,omitempty"`
	Detail  string    `json:"detail,omitempty"`
	Checked time.Time `json:"checked"`
}

// Options tune politeness and freshness.
type Options struct {
	PerHostDelay time.Duration // minimum gap between two requests to one host
	Timeout      time.Duration // per-request
	TTL          time.Duration // how long a cached verdict stands
}

// DefaultOptions are deliberately slow. deadlinks.md is a weekly report; nothing waits on it.
func DefaultOptions() Options {
	return Options{PerHostDelay: time.Second, Timeout: 10 * time.Second, TTL: 7 * 24 * time.Hour}
}

// Checker checks URLs, one host at a time but hosts in parallel, and remembers what it found.
type Checker struct {
	opts   Options
	cache  *Cache
	client *http.Client
}

// New returns a Checker whose cache lives in dir (typically <vault>/.forge).
func New(dir string, opts Options) *Checker {
	return &Checker{
		opts:  opts,
		cache: LoadCache(dir),
		// Redirects are followed — a moved source is not a dead one — but a redirect loop is
		// not worth a report's time.
		client: &http.Client{Timeout: opts.Timeout},
	}
}

// Check returns one Status per URL, in input order. Cached verdicts inside the TTL are
// returned without a request.
func (c *Checker) Check(ctx context.Context, urls []string) []Status {
	out := make([]Status, len(urls))
	byHost := map[string][]int{}
	for i, u := range urls {
		if s, ok := c.cache.Get(u, c.opts.TTL); ok {
			out[i] = s
			continue
		}
		h := hostOf(u)
		byHost[h] = append(byHost[h], i)
	}
	c.run(ctx, urls, byHost, out)
	return out
}

// run walks each host's URLs sequentially, spaced by PerHostDelay, with hosts in parallel.
func (c *Checker) run(ctx context.Context, urls []string, byHost map[string][]int, out []Status) {
	var wg sync.WaitGroup
	for _, idxs := range byHost {
		wg.Add(1)
		go func(idxs []int) {
			defer wg.Done()
			for n, i := range idxs {
				if n > 0 {
					sleep(ctx, c.opts.PerHostDelay)
				}
				out[i] = c.probe(ctx, urls[i])
				c.cache.Put(out[i])
			}
		}(idxs)
	}
	wg.Wait()
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
	case <-ctx.Done():
	}
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw // unparseable: its own bucket, so it cannot serialise a real host
	}
	return u.Host
}
