package engine

import (
	"encoding/json"
	"testing"
)

func TestHostDoesNoIOAndSerializesTheRequest(t *testing.T) {
	req := Request{
		Stage: "research", Prompt: "explain X",
		Context:     map[string]string{"question": "how does X work"},
		Constraints: map[string]string{"cite": "every claim"},
	}
	res, err := Host{}.Run(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Tier != TierHost {
		t.Errorf("Tier = %q, want %q", res.Tier, TierHost)
	}
	if res.Output != "" {
		t.Errorf("Output = %q, want empty — host produces an Instruction, not an Output", res.Output)
	}
	var got instruction
	if err := json.Unmarshal([]byte(res.Instruction), &got); err != nil {
		t.Fatalf("Instruction did not round-trip as JSON: %v", err)
	}
	if got.Stage != req.Stage || got.Prompt != req.Prompt {
		t.Errorf("instruction = %+v, want stage/prompt from the request", got)
	}
}
