package engine

// None is the T0 tier: zero I/O, always refuses. It exists as a real Engine value (rather
// than a nil special case) so callers that hold an Engine interface never need a type
// switch to find out a stage is static.
type None struct{}

func (None) Run(req Request) (Result, error) {
	return Result{}, &NoGenerationError{
		Stage:  req.Stage,
		Reason: "this stage is locked to engine \"none\" (T0 static core)",
	}
}
