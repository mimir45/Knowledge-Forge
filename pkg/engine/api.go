package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// API is the real-HTTP tier.
type API struct {
	RoundTripper http.RoundTripper
	Provider     string // anthropic | openai | openrouter | ollama
	Model        string
	BaseURL      string
	APIKey       string // empty means no auth header — httptest servers need none
}

// envelope is the uniform response contract this package reads regardless of provider.
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
