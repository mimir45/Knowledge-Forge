package engine

import (
	"encoding/json"
	"net/http"
)

// payload builds the provider-specific request body. This is the one part of api.go that
// genuinely differs per provider — the response envelope does not (see api.go's comment).
func (a API) payload(req Request) ([]byte, error) {
	switch a.Provider {
	case "anthropic":
		return json.Marshal(map[string]any{
			"model": a.Model, "max_tokens": 4096,
			"messages": []map[string]string{{"role": "user", "content": req.Prompt}},
		})
	case "openai", "openrouter":
		return json.Marshal(map[string]any{
			"model":    a.Model,
			"messages": []map[string]string{{"role": "user", "content": req.Prompt}},
		})
	case "ollama":
		return json.Marshal(map[string]any{
			"model": a.Model, "prompt": req.Prompt, "stream": false,
		})
	default:
		return json.Marshal(map[string]any{"model": a.Model, "prompt": req.Prompt})
	}
}

func (a API) endpoint() string {
	switch a.Provider {
	case "anthropic":
		return a.BaseURL + "/v1/messages"
	case "openai", "openrouter":
		return a.BaseURL + "/v1/chat/completions"
	case "ollama":
		return a.BaseURL + "/api/generate"
	default:
		return a.BaseURL
	}
}

// authorize sets no header at all when APIKey is empty, so an httptest server needs zero
// auth setup — a deliberate part of "no real key ever read for auth in tests".
func (a API) authorize(r *http.Request) {
	if a.APIKey == "" {
		return
	}
	switch a.Provider {
	case "anthropic":
		r.Header.Set("x-api-key", a.APIKey)
		r.Header.Set("anthropic-version", "2023-06-01")
	default: // openai, openrouter, ollama all speak Bearer when a key is configured
		r.Header.Set("Authorization", "Bearer "+a.APIKey)
	}
}
