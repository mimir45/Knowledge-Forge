// Package engine is Phase 3b's execution layer over the four tiers pkg/config already
// names: none, host, api, advisor. pkg/config decides *what* a pipeline stage should use;
// this package is the only place that acts on that decision — the one part of this binary
// AUDIT §8.4 D-8's neighbour clause and main.go's doc-comment name as making a model call.
//
// "local" is not a fifth implementation. It is select.go's routing alias for the api
// backend pointed at engines.local.base_url — see select.go's doc-comment.
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

// Request is one pipeline stage's ask. Context and Constraints are separate because a
// stage's constraints (e.g. "cite every claim") are not part of what gets asked, they are
// part of how the answer is judged — host.go serializes both for the skill to apply.
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
