package steps

import (
	"github.com/domahidizoltan/go-steps/internal/pkg/step"
)

func Map[IN0, OUT0 any](fn func(in IN0) (OUT0, error)) step.StepWrapper {
	return step.Map(fn)
}

func Filter[IN0 any](fn func(in IN0) (bool, error)) step.StepWrapper {
	return step.Filter(fn)
}

func Split[IN0 any, OUT0 ~uint8](fn func(in IN0) (OUT0, error)) step.StepWrapper {
	return step.Split(fn)
}

func WithBranches[IN0 any](steps ...step.StepsContainer) step.StepWrapper {
	return step.WithBranches[IN0](steps...)
}

func Merge() step.StepWrapper {
	return step.Merge()
}

func GroupBy[IN0 any, OUT0 comparable, OUT1 any](fn func(in IN0) (OUT0, OUT1, error)) step.ReducerWrapper {
	return step.GroupBy(fn)
}

func MultiplyBy[IN0 ~int](multiplier IN0) step.StepWrapper {
	return step.MultiplyBy(multiplier)
}
