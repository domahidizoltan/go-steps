// Package steps is a collection of functions that can be used to
// create and validate transformation steps, transformator options,
// aggregators, input and output adapters and more.
package steps

import (
	"context"
	"errors"
	"io"
	"reflect"
)

const maxArgs = 4

type (
	// StepFn defines a function used as a single step in a transformation chain
	StepFn func(StepInput) StepOutput

	// ReducerFn is an alias for StepFn used as a last step transformation (aggregator)
	ReducerFn StepFn

	// Args are the array of input or output arguments.
	// Most of the time only the first argument is used but others could be used for multi-input or output steps.
	Args [maxArgs]any

	// ArgTypes are holding the [reflect.Type] of arguments.
	// They are mostly used for the validation process.
	ArgTypes [maxArgs]reflect.Type

	// StepInput holds the input arguments for a single step
	StepInput struct {
		Args               Args               // input arguments
		ArgsLen            uint8              // length of input arguments (Args can hold zero values, so we need to know it's length)
		TransformerOptions TransformerOptions // transformer options to pass around in the chain
	}

	// StepOutput holds the output arguments for a single step
	StepOutput struct {
		Error   error // error result of the step
		Args    Args  // output arguments
		ArgsLen uint8 // length of output arguments (Args can hold zero values, so we need to know it's length)
		Skip    bool  // used to notify the processor that further transformation steps should be skipped
	}

	// StepWrapper is a container for a single transformation step
	StepWrapper struct {
		Name     string                                            // name of the step
		StepFn   StepFn                                            // the transformation step function
		Validate func(prevStepArgTypes ArgTypes) (ArgTypes, error) // validation of the current step in the chain
		Reset    func()                                            // reset the step state before processing
	}

	// ReducerWrapper is a container for an aggregation step
	ReducerWrapper struct {
		Name      string                                            // name of the step
		ReducerFn ReducerFn                                         // the aggregation step function
		Validate  func(prevStepArgTypes ArgTypes) (ArgTypes, error) // validation of the aggregation step in the chain
		Reset     func()                                            // reset the aggregation state before processing
	}

	// StepsBranch represents a sub-path of a branching transformation chain
	StepsBranch struct {
		Error             error           // error result of the sub-path
		StepWrappers      []StepWrapper   // the steps in the sub-path
		Steps             []StepFn        // the already validated steps
		AggregatorWrapper *ReducerWrapper // the aggregation step in the sub-path
		Aggregator        ReducerFn       // the already validated aggregator function
	}

	// SkipFirstArgValidation is used to tell the validator this step is the first in the chain,
	// so it can't be validated against the previous step.
	SkipFirstArgValidation struct{}

	// TransformerOptions holds the options for the transformer
	TransformerOptions struct {
		Name         string
		LogWriter    io.Writer
		ErrorHandler func(error)
		PanicHandler func(error)
		Ctx          context.Context
		ChanSize     uint
	}
)

var (
	ErrStepValidationFailed  = errors.New("step validation failed")           // step validation returned an error
	ErrIncompatibleInArgType = errors.New("incompatible input argument type") // the outputs of the previous step don't match the inputs of the current step
	ErrInvalidAggregator     = errors.New("invalid aggregator")               // aggregator has no reducer or name defined
	ErrInvalidStep           = errors.New("invalid step")                     // step has no step or name defined
)
