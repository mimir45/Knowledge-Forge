// Package engine is the execution layer over the four tiers pkg/config already names:
// none, host, api, advisor. pkg/config decides *what* a pipeline stage should use.
package engine

import "fmt"

// Tier is one of the four engine implementations this package ships.
type Tier string

const (
	TierNone    Tier = "none"
	TierHost    Tier = "host"
	TierAPI     Tier = "api"
	TierAdvisor Tier = "advisor"
)

// Request is one pipeline stage's ask.
type Request struct {
	Stage       string
	Prompt      string
	Context     map[string]string
	Constraints map[string]string
}

// Result is what a tier produced. Instruction is non-empty only for TierHost: the binary
// cannot itself call the model, so it hands the skill what to run instead of an Output.
type Result struct {
	Output      string
	Instruction string
	Tokens      int
	CostUSD     float64
	Tier        Tier
}

// Engine is the one interface every tier implements.
type Engine interface {
	Run(req Request) (Result, error)
}

// NoGenerationError is TierNone's only possible return: a typed refusal, never a panic or
// a silent empty Result, so a caller can distinguish "T0 by design" from "the call failed".
type NoGenerationError struct {
	Stage, Reason string
}

func (e *NoGenerationError) Error() string {
	return fmt.Sprintf("engine: stage %q makes zero model calls: %s", e.Stage, e.Reason)
}
