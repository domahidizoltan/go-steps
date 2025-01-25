package steps

import (
	"reflect"
)

type (
	inputType[T any] interface {
		chan T | []T
	}

	input[T any, I inputType[T]] struct {
		data I
	}

	transformator struct {
		Error               error
		Aggregator          ReducerFn
		LastAggregatedValue *StepOutput
		Steps               []StepFn
	}

	stepsTransformator[T any, I inputType[T]] struct {
		in I
		transformator
	}
)

func Transform[T any, I inputType[T]](in I) input[T, I] {
	return input[T, I]{in}
}

// TODO .WithOptions to add debug options, transformator name, etc
func Steps(s ...StepWrapper) StepsBranch {
	return StepsBranch{
		StepWrappers: s,
	}
}

func (i input[T, I]) WithSteps(steps ...StepWrapper) stepsTransformator[T, I] {
	return i.With(Steps(steps...))
}

func (i input[T, I]) With(steps StepsBranch) stepsTransformator[T, I] {
	t := stepsTransformator[T, I]{
		transformator: transformator{
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
