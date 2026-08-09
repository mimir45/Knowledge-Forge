package engine

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// roundTripperFunc adapts a func to http.RoundTripper, so each test can stand in its own
// httptest.Server without ever dialing a real network.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testServer(t *testing.T, handler http.HandlerFunc) http.RoundTripper {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		r.URL.Scheme, r.URL.Host = "http", srv.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(r)
	})
}

func TestAPIRunReadsTheEnvelope(t *testing.T) {
	rt := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("no key configured but an Authorization header was sent")
		}
		w.Write([]byte(`{"output":"hi","tokens":42,"cost_usd":0.01}`))
	})
	a := API{RoundTripper: rt, Provider: "openai", Model: "gpt", BaseURL: "http://x"}
	res, err := a.Run(Request{Stage: "research", Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Output != "hi" || res.Tokens != 42 || res.CostUSD != 0.01 || res.Tier != TierAPI {
		t.Errorf("Result = %+v", res)
	}
}

func TestAPISendsProviderAuthHeader(t *testing.T) {
	rt := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != "sk-1" {
			t.Errorf("x-api-key = %q, want sk-1", got)
		}
		w.Write([]byte(`{"output":"ok"}`))
	})
	a := API{RoundTripper: rt, Provider: "anthropic", BaseURL: "http://x", APIKey: "sk-1"}
	if _, err := a.Run(Request{Stage: "verify", Prompt: "p"}); err != nil {
		t.Fatal(err)
	}
}

func TestAPINonOKStatusIsAnError(t *testing.T) {
	rt := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limited"))
	})
	a := API{RoundTripper: rt, Provider: "openai", BaseURL: "http://x"}
	if _, err := a.Run(Request{Stage: "research", Prompt: "p"}); err == nil {
		t.Fatal("want an error on a non-200 response")
	}
}

func TestAPIMalformedEnvelopeIsAnError(t *testing.T) {
	rt := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	})
	a := API{RoundTripper: rt, Provider: "ollama", BaseURL: "http://x"}
	if _, err := a.Run(Request{Stage: "research", Prompt: "p"}); err == nil {
		t.Fatal("want an error on a response that is not the envelope shape")
	}
}
