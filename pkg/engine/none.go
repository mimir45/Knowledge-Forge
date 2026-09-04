package engine

// None is the T0 tier: zero I/O, always refuses.
type None struct{}

func (None) Run(req Request) (Result, error) {
	return Result{}, &NoGenerationError{
		Stage:  req.Stage,
		Reason: "this stage is locked to engine \"none\" (T0 static core)",
	}
}
