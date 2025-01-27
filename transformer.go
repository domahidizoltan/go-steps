package steps

import (
	"reflect"
)

type (
	inputType[T any] interface {
		chan T | []T
	}

	input[T any, IT inputType[T]] struct {
		data IT
	}

	transformer struct {
		error               error
		aggregator          ReducerFn
		lastAggregatedValue *StepOutput
		steps               []StepFn
	}

	stepsTransformer[T any, IT inputType[T]] struct {
		input IT
		transformer
	}
)

func Transform[T any, IT inputType[T]](in IT) input[T, IT] {
	return input[T, IT]{in}
}

// TODO .WithOptions to add debug options, transformer name, etc
func Steps(s ...StepWrapper) StepsBranch {
	return StepsBranch{
		StepWrappers: s,
	}
}

func (i input[T, IT]) WithSteps(steps ...StepWrapper) stepsTransformer[T, IT] {
	return i.With(Steps(steps...))
}

func (i input[T, IT]) With(steps StepsBranch) stepsTransformer[T, IT] {
	t := stepsTransformer[T, IT]{
		transformer: transformer{
			error: steps.Validate(),
		},
	}
	if t.error != nil {
		return t
	}

	t.input = i.data
	t.steps = steps.Steps
	if steps.AggregatorWrapper != nil {
		t.aggregator = steps.AggregatorWrapper.ReducerFn
	}
	return t
}

func (s StepsBranch) Aggregate(fn ReducerWrapper) StepsBranch {
	return StepsBranch{
		StepWrappers:      s.StepWrappers,
		AggregatorWrapper: &fn,
	}
}

func (s *StepsBranch) Validate() error {
	if s.Error != nil {
		return s.Error
	}

	var lastOutTypes ArgTypes
	s.Steps, lastOutTypes, s.Error = getValidatedSteps[SkipFirstArgValidation](s.StepWrappers)

	if s.AggregatorWrapper != nil {
		if s.Error != nil {
			return s.Error
		}

		if s.Steps == nil {
			lastOutTypes = ArgTypes{reflect.TypeOf(SkipFirstArgValidation{})}
		}
		_, s.Error = s.AggregatorWrapper.Validate(lastOutTypes)
	}

	return s.Error
}
