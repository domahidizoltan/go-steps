package steps

import (
	"errors"
	"reflect"
)

const maxArgs = 4

type (
	StepFn    func(StepInput) StepOutput
	ReducerFn StepFn
	Args      [maxArgs]any
	ArgTypes  [maxArgs]reflect.Type

	StepInput struct {
		Args    Args
		ArgsLen uint8
	}

	StepOutput struct {
		Error   error
		Args    Args
		ArgsLen uint8
		Skip    bool
	}

	StepWrapper struct {
		Name     string
		StepFn   StepFn
		Validate func(prevStepArgTypes ArgTypes) (ArgTypes, error)
	}

	ReducerWrapper struct {
		Name      string
		ReducerFn ReducerFn
		Validate  func(prevStepArgTypes ArgTypes) (ArgTypes, error)
	}

	StepsBranch struct {
		Error             error
		StepWrappers      []StepWrapper
		AggregatorWrapper *ReducerWrapper
		Aggregator        ReducerFn
		Steps             []StepFn
	}

	SkipFirstArgValidation struct{}

	TransformerOptions struct {
		Name string
	}
)

var (
	ErrEmptyTransformInputType = errors.New("first step in type is empty")
	ErrStepValidationFailed    = errors.New("step validation failed")
	ErrIncompatibleInArgType   = errors.New("incompatible input argument type")
	ErrInvalidAggregator       = errors.New("invalid aggregator")
	ErrInvalidStep             = errors.New("invalid step")
)
