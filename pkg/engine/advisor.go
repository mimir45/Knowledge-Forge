package engine

import (
	"encoding/json"
	"fmt"
)

// Advisor is T3: critique-only. It never rewrites a note — DESIGN's contract is disputed
// claims, what's missing, a confidence verdict, and a minimal patch, and this type is what
// keeps that contract structural rather than a hope expressed in a prompt string.
type Advisor struct {
	API API
}

// Critique is the shape Advisor.Run requires the model's Output to decode as. A response
// that is not this shape is treated as a protocol violation, not a low-effort critique.
type Critique struct {
	Disputed   []string `json:"disputed"`
	Missing    []string `json:"missing"`
	Confidence string   `json:"confidence"`
	Patch      string   `json:"patch"`
}

// preamble is prepended so the model sees the contract even though this package never
// sends provider-specific "system" fields — Advisor's payload still goes through
// API.payload, so the instruction has to travel inside Prompt.
const preamble = "You are a critique-only reviewer. Return ONLY JSON matching " +
	"{disputed:[string],missing:[string],confidence:string,patch:string}. Never rewrite; " +
	"list disputed claims, what's missing, a confidence verdict, and a minimal patch.\n\n"

func (a Advisor) Run(req Request) (Result, error) {
	req.Prompt = preamble + req.Prompt
	res, err := a.API.Run(req)
	if err != nil {
		return Result{}, err
	}
	if err := validateCritique(res.Output); err != nil {
		return Result{}, err
	}
	res.Tier = TierAdvisor
	return res, nil
}

// validateCritique requires Confidence to be non-empty, not merely that Output parses as
// JSON — {} parses cleanly and would otherwise pass, and an empty confidence verdict means
// nothing was actually judged.
func validateCritique(output string) error {
	var c Critique
	if err := json.Unmarshal([]byte(output), &c); err != nil {
		return fmt.Errorf("engine: advisor output was not a Critique: %w", err)
	}
	if c.Confidence == "" {
		return fmt.Errorf("engine: advisor output has no confidence verdict")
	}
	return nil
}
