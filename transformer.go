package steps

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

type (
	inputType[T any] interface {
		chan T | []T
	}

	input[T any, IT inputType[T]] struct {
		data    IT
		options TransformerOptions
	}

	transformer struct {
		options             TransformerOptions
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

func Transform[T any, IT inputType[T]](in IT, options ...func(*TransformerOptions)) input[T, IT] {
	opts := TransformerOptions{
		LogWriter: os.Stdout,
	}
	for _, withOption := range options {
		withOption(&opts)
	}
	return input[T, IT]{in, opts}
}

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
			options: i.options,
			error:   steps.Validate(),
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

func Aggregate(fn ReducerWrapper) StepsBranch {
	return StepsBranch{
		AggregatorWrapper: &fn,
	}
}

func (s StepsBranch) Aggregate(fn ReducerWrapper) StepsBranch {
	return StepsBranch{
		StepWrappers:      s.StepWrappers,
		AggregatorWrapper: &fn,
	}
}

func (s *StepsBranch) Validate() error {
	if s.Error != nil {
		s.StepWrappers = nil
		return s.Error
	}

	var lastOutTypes ArgTypes
	s.Steps, lastOutTypes, s.Error = getValidatedSteps[SkipFirstArgValidation](s.StepWrappers)

	aggWr := s.AggregatorWrapper
	if aggWr != nil {
		if s.Error != nil {
			s.StepWrappers = nil
			return s.Error
		}

		if len(aggWr.Name) == 0 || aggWr.ReducerFn == nil {
			s.Error = ErrInvalidAggregator
			s.AggregatorWrapper = nil
			return s.Error
		}

		if s.Steps == nil {
			lastOutTypes = ArgTypes{reflect.TypeOf(SkipFirstArgValidation{})}
		}
		_, s.Error = aggWr.Validate(lastOutTypes)
	}

	return s.Error
}

func (s stepsTransformer[T, IT]) emptyErrorHandler() func(error) {
	return func(err error) {
		fmt.Fprintln(s.options.LogWriter, "error occured:", err)
	}
}
