package linkcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func fast() Options {
	return Options{PerHostDelay: time.Millisecond, Timeout: 2 * time.Second, TTL: time.Hour}
}

// server answers by path and records the methods it was asked for.
func server(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var mu sync.Mutex
	var methods []string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		methods = append(methods, r.Method+" "+r.URL.Path)
		mu.Unlock()
		switch r.URL.Path {
		case "/gone":
			w.WriteHeader(http.StatusNotFound)
		case "/no-head":
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}
	}))
	t.Cleanup(s.Close)
	return s, &methods
}

func TestVerdicts(t *testing.T) {
	s, _ := server(t)
	c := New(t.TempDir(), fast())
	got := c.Check(context.Background(), []string{s.URL + "/ok", s.URL + "/gone", "ftp://x/y"})
	for i, want := range []Verdict{Alive, Dead, Unreachable} {
		if got[i].Verdict != want {
			t.Errorf("%s = %s (%d), want %s", got[i].URL, got[i].Verdict, got[i].Code, want)
		}
	}
}

// A host that refuses HEAD must be retried with GET, or every source behind such a CDN reads
// as dead.
func TestHeadRefusalFallsBackToGet(t *testing.T) {
	s, methods := server(t)
	c := New(t.TempDir(), fast())
	if got := c.Check(context.Background(), []string{s.URL + "/no-head"}); got[0].Verdict != Alive {
		t.Fatalf("verdict = %s (%d), want alive after the GET retry", got[0].Verdict, got[0].Code)
	}
	if len(*methods) != 2 || (*methods)[1] != "GET /no-head" {
		t.Errorf("methods = %v, want a HEAD then a GET", *methods)
	}
}

// The report must not claim a URL is gone because the network was. This is the failure mode
// that would make deadlinks.md useless on a plane.
func TestTransportFailureIsUnreachableNotDead(t *testing.T) {
	c := New(t.TempDir(), Options{PerHostDelay: time.Millisecond, Timeout: 50 * time.Millisecond})
	got := c.Check(context.Background(), []string{"https://localhost:1/nothing"})
	if got[0].Verdict != Unreachable {
		t.Errorf("verdict = %s, want unreachable", got[0].Verdict)
	}
}

func TestCacheAvoidsASecondRequest(t *testing.T) {
	s, methods := server(t)
	dir := t.TempDir()
	c := New(dir, fast())
	c.Check(context.Background(), []string{s.URL + "/ok"})
	if err := c.cache.Save(); err != nil {
		t.Fatal(err)
	}
	before := len(*methods)
	if got := New(dir, fast()).Check(context.Background(), []string{s.URL + "/ok"}); got[0].Verdict != Alive {
		t.Fatalf("verdict = %s, want the cached alive", got[0].Verdict)
	}
	if len(*methods) != before {
		t.Errorf("methods = %v, want no new request", *methods)
	}
}

// One offline run must not poison the report for a week.
func TestUnreachableIsNotServedFromCache(t *testing.T) {
	c := LoadCache(t.TempDir())
	c.Put(Status{URL: "u", Verdict: Unreachable, Checked: now()})
	if _, ok := c.Get("u", time.Hour); ok {
		t.Error("an unreachable verdict was served from cache")
	}
	c.Put(Status{URL: "u", Verdict: Dead, Code: 404, Checked: now()})
	if _, ok := c.Get("u", time.Hour); !ok {
		t.Error("a real verdict was not cached")
	}
}

func TestExpiredEntriesAreRechecked(t *testing.T) {
	c := LoadCache(t.TempDir())
	c.Put(Status{URL: "u", Verdict: Alive, Checked: now().Add(-2 * time.Hour)})
	if _, ok := c.Get("u", time.Hour); ok {
		t.Error("an entry past its TTL was reused")
	}
}

// A cache that went bad must cost a slow report, never a failed one.
func TestCorruptCacheLoadsEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := writeFile(dir, "{not json"); err != nil {
		t.Fatal(err)
	}
	if got := LoadCache(dir); len(got.seen) != 0 {
		t.Errorf("seen = %v, want empty", got.seen)
	}
}

// Requests to one host are spaced; different hosts do not wait on each other.
func TestPerHostDelayIsRespected(t *testing.T) {
	s, _ := server(t)
	c := New(t.TempDir(), Options{PerHostDelay: 80 * time.Millisecond, Timeout: time.Second})
	start := time.Now()
	c.Check(context.Background(), []string{s.URL + "/a", s.URL + "/b", s.URL + "/c"})
	if d := time.Since(start); d < 160*time.Millisecond {
		t.Errorf("three requests to one host took %v, want at least two delays", d)
	}
}

func writeFile(dir, content string) error {
	return os.WriteFile(filepath.Join(dir, CacheFile), []byte(content), 0o644)
}

func BenchmarkCacheHit(b *testing.B) {
	c := LoadCache(b.TempDir())
	c.Put(Status{URL: "u", Verdict: Alive, Checked: now()})
	b.ReportAllocs()
	for b.Loop() {
		c.Get("u", time.Hour)
	}
}
