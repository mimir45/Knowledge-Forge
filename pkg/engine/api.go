package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// API is the real-HTTP tier. RoundTripper is always injected — pkg/engine never dials a
// live network in a test, and select.go/api.go together are the only places in this
// binary allowed to make a model call at all (main.go's doc-comment names the exception).
//
// API also serves the "local" alias (select.go): a Local config with Enabled and a
// non-empty BaseURL resolves to an API value with Provider "ollama", the shape a locally
// hosted model server speaks.
type API struct {
	RoundTripper http.RoundTripper
	Provider     string // anthropic | openai | openrouter | ollama
	Model        string
	BaseURL      string
	APIKey       string // empty means no auth header — httptest servers need none
}

// envelope is the uniform response contract this package reads regardless of provider.
// No pricing table lives in this repo (cfg.Engines carries no $/token rates for any
// provider), so cost is read verbatim from the server's response rather than invented
// client-side — the caller (a real provider, or a test's httptest.Server) states its own
// price. Request *payload* shape still varies per provider; only the response is uniform.
type envelope struct {
	Output  string  `json:"output"`
	Tokens  int     `json:"tokens"`
	CostUSD float64 `json:"cost_usd"`
}

func (a API) Run(req Request) (Result, error) {
	body, err := a.payload(req)
	if err != nil {
		return Result{}, err
	}
	httpReq, err := http.NewRequest(http.MethodPost, a.endpoint(), bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	a.authorize(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")
	return a.do(httpReq)
}

func (a API) do(httpReq *http.Request) (Result, error) {
	resp, err := a.client().Do(httpReq)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("engine: %s returned %s: %s", a.Provider, resp.Status, raw)
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return Result{}, fmt.Errorf("engine: %s response was not the expected envelope: %w", a.Provider, err)
	}
	return Result{Output: env.Output, Tokens: env.Tokens, CostUSD: env.CostUSD, Tier: TierAPI}, nil
}

func (a API) client() *http.Client {
	return &http.Client{Transport: a.RoundTripper}
}
