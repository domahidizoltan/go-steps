package step

import (
	"errors"
	"reflect"
)

const MaxArgs = 4

type (
	StepInput struct {
		Args    [MaxArgs]any
		ArgsLen uint8
	}

	StepOutput struct {
		Error   error
		Args    [MaxArgs]any
		ArgsLen uint8
		Skip    bool
		// Accumulate bool
	}

	StepWrapper struct {
		InTypes  [MaxArgs]reflect.Type
		OutTypes [MaxArgs]reflect.Type
		StepFn   StepFn
	}

	StepFn func(StepInput) StepOutput

	ReducerWrapper struct {
		InTypes   [MaxArgs]reflect.Type
		OutTypes  [MaxArgs]reflect.Type
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

var ErrInvalidInputType = errors.New("Invalid input type")

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
