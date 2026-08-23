package engine

import (
	"net/http"
	"testing"
)

func TestAdvisorAcceptsAWellFormedCritique(t *testing.T) {
	rt := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"output":"{\"disputed\":[\"claim 1\"],\"missing\":[],` +
			`\"confidence\":\"medium\",\"patch\":\"add a source\"}"}`))
	})
	adv := Advisor{API: API{RoundTripper: rt, Provider: "anthropic", BaseURL: "http://x"}}
	res, err := adv.Run(Request{Stage: "verify", Prompt: "check this note"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Tier != TierAdvisor {
		t.Errorf("Tier = %q, want %q", res.Tier, TierAdvisor)
	}
}

func TestAdvisorRejectsAnEmptyConfidence(t *testing.T) {
	rt := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"output":"{}"}`)) // valid JSON, but judges nothing
	})
	adv := Advisor{API: API{RoundTripper: rt, Provider: "anthropic", BaseURL: "http://x"}}
	if _, err := adv.Run(Request{Stage: "verify", Prompt: "p"}); err == nil {
		t.Fatal("want an error when Output has no confidence verdict")
	}
}

func TestAdvisorRejectsNonJSONOutput(t *testing.T) {
	rt := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"output":"looks great, ship it"}`))
	})
	adv := Advisor{API: API{RoundTripper: rt, Provider: "anthropic", BaseURL: "http://x"}}
	if _, err := adv.Run(Request{Stage: "verify", Prompt: "p"}); err == nil {
		t.Fatal("want an error when Output is prose, not a Critique")
	}
}
