package step

import (
	"errors"
	"reflect"
)

const maxArgs = 4

type (
	StepInput struct {
		Args    [maxArgs]any
		ArgsLen uint8
	}

	StepOutput struct {
		Error   error
		Args    [maxArgs]any
		ArgsLen uint8
		Skip    bool
		// Accumulate bool
	}

	StepWrapper struct {
		Name            string
		InTypes         [maxArgs]reflect.Type
		OutTypes        [maxArgs]reflect.Type
		StepFn          StepFn
		InTypeZeroValue any
		InnerValidator  func(reflect.Type) error
	}

	StepFn func(StepInput) StepOutput

	ReducerWrapper struct {
		InTypes   [maxArgs]reflect.Type
		OutTypes  [maxArgs]reflect.Type
		ReducerFn ReducerFn
	}
	ReducerFn func(StepInput) StepOutput

	stepType uint8

	StepsContainer struct {
		Error        error
		StepWrappers []StepWrapper
		Aggregator   ReducerFn
		Steps        []StepFn
		Validated    stepType
	}

	Transformator struct {
		Error               error
		Aggregator          ReducerFn
		LastAggregatedValue *StepOutput
		Steps               []StepFn
		Validated           stepType
	}
)

var (
	ErrTransformInputTypeIsDifferent   = errors.New("transform input type is different from first step in type")
	ErrEmptyFirstStepInType            = errors.New("first step in type is empty")
	ErrStepOutAndNextInTypeIsDifferent = errors.New("step out type is different from next step in type")
	ErrEmptyStepOutType                = errors.New("step out type is empty")
	ErrInnerStepValidationFailed       = errors.New("inner step validation failed")
)

func Steps(s ...StepWrapper) StepsContainer {
	return StepsContainer{
		StepWrappers: s,
	}
}

func (t StepsContainer) Aggregate(fn ReducerWrapper) StepsContainer { //?
	return StepsContainer{
		StepWrappers: t.StepWrappers,
		Aggregator:   fn.ReducerFn,
	}
}
