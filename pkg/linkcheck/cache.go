package linkcheck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// CacheFile is the cache's name inside the directory given to New.
const CacheFile = "linkcheck.json"

// Cache remembers verdicts between runs. It is a derived cache like every other one in this
// project: deleting it costs a slow report, never a wrong one.
type Cache struct {
	mu   sync.Mutex
	dir  string
	seen map[string]Status
}

// LoadCache reads the cache from dir. A missing or corrupt file is an empty cache, not an
// error — a report must never fail to run because a cache went bad.
func LoadCache(dir string) *Cache {
	c := &Cache{dir: dir, seen: map[string]Status{}}
	b, err := os.ReadFile(filepath.Join(dir, CacheFile))
	if err != nil {
		return c
	}
	var entries []Status
	if json.Unmarshal(b, &entries) != nil {
		return c
	}
	for _, s := range entries {
		c.seen[s.URL] = s
	}
	return c
}

// Get returns a cached verdict if it is younger than ttl.
//
// Unreachable is never cached back: it is not a fact about the URL, it is a fact about the
// network at one moment, and holding it for a week would mean one offline run poisons the
// report until the TTL expires.
func (c *Cache) Get(url string, ttl time.Duration) (Status, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.seen[url]
	if !ok || s.Verdict == Unreachable || now().Sub(s.Checked) > ttl {
		return Status{}, false
	}
	return s, true
}

// Put records a verdict.
func (c *Cache) Put(s Status) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen[s.URL] = s
}

// Save writes the cache back, sorted by URL so a commit of .forge/ shows real changes.
func (c *Cache) Save() error {
	c.mu.Lock()
	entries := make([]Status, 0, len(c.seen))
	for _, s := range c.seen {
		entries = append(entries, s)
	}
	c.mu.Unlock()
	sort.Slice(entries, func(i, j int) bool { return entries[i].URL < entries[j].URL })
	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dir, CacheFile), b, 0o644)
}
