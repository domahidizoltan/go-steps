package chans

import (
	"github.com/domahidizoltan/go-steps/internal/pkg/step"
	"github.com/domahidizoltan/go-steps/types"
)

func Map[U, V any](fn func(in U) (V, error)) types.StepWrapper {
	return step.Map(fn)
}

func Filter[U any](fn func(in U) (bool, error)) types.StepWrapper {
	return step.Filter(fn)
}

func MultiplyBy[U ~int](multiplier U) types.StepWrapper {
	return step.MultiplyBy(multiplier)
}
