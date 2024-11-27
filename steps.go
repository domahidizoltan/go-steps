package steps

import (
	step "github.com/domahidizoltan/go-steps/internal/pkg/step"
	"github.com/domahidizoltan/go-steps/types"
)

func Map[IN0, OUT0 any](fn func(in IN0) (OUT0, error)) types.StepWrapper {
	return step.Map(fn)
}

func Filter[IN0 any](fn func(in IN0) (bool, error)) types.StepWrapper {
	return step.Filter(fn)
}

func GroupBy[IN0 any, OUT0 comparable, OUT1 any](fn func(in IN0) (OUT0, OUT1, error)) types.ReducerWrapper {
	return step.GroupBy(fn)
}

func MultiplyBy[IN0 ~int](multiplier IN0) types.StepWrapper {
	return step.MultiplyBy(multiplier)
}
