package steps

import (
	"reflect"

	"github.com/domahidizoltan/go-steps/internal/pkg/step"
)

type (
	inputType[T any] interface {
		chan T | []T
	}

	input[T any, I inputType[T]] struct {
		data I
	}

	stepsBranch step.StepsBranch

	transformator[T any, I inputType[T]] struct {
		in I
		step.Transformator
	}
)

func Transform[T any, I inputType[T]](in I) input[T, I] {
	return input[T, I]{in}
}

// TODO .WithOptions to add debug options, transformator name, etc
func Steps(s ...step.StepWrapper) stepsBranch {
	return stepsBranch(step.Steps(s...))
}

func (i input[T, I]) WithSteps(steps ...step.StepWrapper) transformator[T, I] {
	return i.With(Steps(steps...))
}

func (i input[T, I]) With(steps stepsBranch) transformator[T, I] {
	t := transformator[T, I]{
		Transformator: step.Transformator{
			Error: steps.Validate(),
		},
	}
	if t.Error != nil {
		return t
	}

	t.in = i.data
	t.Steps = steps.Steps
	if steps.AggregatorWrapper != nil {
		t.Aggregator = steps.AggregatorWrapper.ReducerFn
	}
	return t
}

func (s stepsBranch) Aggregate(fn step.ReducerWrapper) stepsBranch {
	return stepsBranch{
		StepWrappers:      s.StepWrappers,
		AggregatorWrapper: &fn,
	}
}

func (s *stepsBranch) Validate() error {
	if s.Error != nil {
		return s.Error
	}

	var lastOutTypes step.ArgTypes
	s.Steps, lastOutTypes, s.Error = step.GetValidatedSteps[step.SkipFirstArgValidation](s.StepWrappers)

	if s.AggregatorWrapper != nil {
		if s.Error != nil {
			return s.Error
		}

		if s.Steps == nil {
			lastOutTypes = step.ArgTypes{reflect.TypeOf(step.SkipFirstArgValidation{})}
		}
		_, s.Error = s.AggregatorWrapper.Validate(lastOutTypes)
	}

	return s.Error
}
