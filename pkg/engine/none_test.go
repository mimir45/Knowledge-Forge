package engine

import (
	"errors"
	"testing"
)

func TestNoneAlwaysRefuses(t *testing.T) {
	_, err := None{}.Run(Request{Stage: "recall"})
	var ng *NoGenerationError
	if !errors.As(err, &ng) {
		t.Fatalf("Run error = %v, want *NoGenerationError", err)
	}
	if ng.Stage != "recall" {
		t.Errorf("Stage = %q, want %q", ng.Stage, "recall")
	}
}

func TestNoneNeverReturnsAResult(t *testing.T) {
	res, _ := None{}.Run(Request{Stage: "write"})
	if res != (Result{}) {
		t.Errorf("Result = %+v, want the zero value", res)
	}
}
