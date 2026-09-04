package engine

import "encoding/json"

// Host is the seam for the tier the Go binary cannot itself call: the model runs inside
// the Claude Code session, not this process.
type Host struct{}

// instruction is what Result.Instruction serializes: the resolved request.
type instruction struct {
	Stage       string            `json:"stage"`
	Prompt      string            `json:"prompt"`
	Context     map[string]string `json:"context,omitempty"`
	Constraints map[string]string `json:"constraints,omitempty"`
}

func (Host) Run(req Request) (Result, error) {
	b, err := json.Marshal(instruction{ //nolint:gosimple // deliberate: see instruction's doc comment
		Stage: req.Stage, Prompt: req.Prompt,
		Context: req.Context, Constraints: req.Constraints,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Instruction: string(b), Tier: TierHost}, nil
}
