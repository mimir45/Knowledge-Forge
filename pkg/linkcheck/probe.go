package linkcheck

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// UserAgent identifies the checker. A blank or default Go user agent is what several docs
// hosts answer 403 to, which would read as a dead link.
const UserAgent = "knowledge-forge-linkcheck/1 (+https://github.com/knowledge-forge)"

// probe asks the server about one URL: HEAD first, then GET if HEAD is not honoured.
func (c *Checker) probe(ctx context.Context, raw string) Status {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return Status{URL: raw, Verdict: Unreachable, Detail: "not an http url", Checked: now()}
	}
	st := c.request(ctx, http.MethodHead, raw)
	if headRefused(st) {
		st = c.request(ctx, http.MethodGet, raw)
	}
	return st
}

// headRefused reports whether a response means "this server does not do HEAD" rather
// than "this resource is gone".
func headRefused(s Status) bool {
	switch s.Code {
	case http.StatusMethodNotAllowed, http.StatusNotImplemented,
		http.StatusForbidden, http.StatusBadRequest:
		return true
	}
	return false
}

func (c *Checker) request(ctx context.Context, method, raw string) Status {
	req, err := http.NewRequestWithContext(ctx, method, raw, nil)
	if err != nil {
		return Status{URL: raw, Verdict: Unreachable, Detail: err.Error(), Checked: now()}
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		// No answer at all. We do not know whether the resource is gone, so we do not claim it.
		return Status{URL: raw, Verdict: Unreachable, Detail: trimErr(err), Checked: now()}
	}
	resp.Body.Close()
	return Status{URL: raw, Verdict: verdictOf(resp.StatusCode), Code: resp.StatusCode, Checked: now()}
}

func verdictOf(code int) Verdict {
	if code >= 200 && code < 400 {
		return Alive
	}
	return Dead
}

// trimErr drops the "Get \"https://...\": " prefix net/http prepends, which repeats the URL
// the Status already carries.
func trimErr(err error) string {
	s := err.Error()
	if _, rest, ok := strings.Cut(s, "\": "); ok {
		return rest
	}
	return s
}

// now is a variable so tests can freeze it.
var now = func() time.Time { return time.Now().UTC() }
