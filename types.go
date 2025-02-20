package steps

import (
	"context"
	"errors"
	"io"
	"reflect"
)

const maxArgs = 4

type (
	StepFn    func(StepInput) StepOutput
	ReducerFn StepFn
	Args      [maxArgs]any
	ArgTypes  [maxArgs]reflect.Type

	StepInput struct {
		Args               Args
		ArgsLen            uint8
		TransformerOptions TransformerOptions
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
		Name         string
		LogWriter    io.Writer
		ErrorHandler func(error)
		PanicHandler func(error)
		Ctx          context.Context
		ChanSize     uint
	}
)

var (
	ErrEmptyTransformInputType = errors.New("first step in type is empty")
	ErrStepValidationFailed    = errors.New("step validation failed")
	ErrIncompatibleInArgType   = errors.New("incompatible input argument type")
	ErrInvalidAggregator       = errors.New("invalid aggregator")
	ErrInvalidStep             = errors.New("invalid step")
)

func WithName(name string) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.Name = name
	}
}

func WithLogWriter(writer io.Writer) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.LogWriter = writer
	}
}

func WithErrorHandler(handler func(error)) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.ErrorHandler = handler
	}
}

func WithPanicHandler(handler func(error)) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.PanicHandler = handler
	}
}

func WithContext(ctx context.Context) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.Ctx = ctx
	}
}

func WithChanSize(size uint) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.ChanSize = size
	}
}
