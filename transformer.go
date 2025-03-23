package steps

import (
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
		stateResets         []func()
	}

	stepsTransformer[T any, IT inputType[T]] struct {
		input IT
		transformer
	}
)

// TransformFn is an alternative for [Transform] where the input is a function.
// This could be used to implement new input sources like files or database connections.
func TransformFn[T any, IT inputType[T]](in func(TransformerOptions) IT, options ...func(*TransformerOptions)) input[T, IT] {
	input := in(buildOpts(options...))
	return Transform[T](input, options...)
}

// Transform creates a builder for the transforation chain.
// It takes the input data (slice or chan) and the transformer options.
func Transform[T any, IT inputType[T]](in IT, options ...func(*TransformerOptions)) input[T, IT] {
	return input[T, IT]{in, buildOpts(options...)}
}

// Steps creates the steps of a transformation chain.
func Steps(s ...StepWrapper) StepsBranch {
	return StepsBranch{
		StepWrappers: s,
	}
}

// WithSteps is a shortcut to With(Steps(...))
func (i input[T, IT]) WithSteps(steps ...StepWrapper) stepsTransformer[T, IT] {
	return i.With(Steps(steps...))
}

// With adds the steps to the existing transformation chain.
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
	for _, s := range steps.StepWrappers {
		if s.Reset != nil {
			t.stateResets = append(t.stateResets, s.Reset)
		}
	}

	if steps.AggregatorWrapper != nil {
		t.aggregator = steps.AggregatorWrapper.ReducerFn
		if steps.AggregatorWrapper.Reset != nil {
			t.stateResets = append(t.stateResets, steps.AggregatorWrapper.Reset)
		}
	}

	return t
}

// Aggregate adds a reducer to the transformer
func Aggregate(fn ReducerWrapper) StepsBranch {
	return StepsBranch{
		AggregatorWrapper: &fn,
	}
}

// Aggregate ads a reducer to the transformer with steps previously defined
func (s StepsBranch) Aggregate(fn ReducerWrapper) StepsBranch {
	return StepsBranch{
		StepWrappers:      s.StepWrappers,
		AggregatorWrapper: &fn,
	}
}

// Validate is runs the validation for the steps in the transformation chain.
// One of the main purpose of the validator is to reduce the possible runtime errors caused by reflection usage.
// The validator tries to check that the output of the steps are matching the inputs of the next steps.
// It could be triggered explicitly to validate the chain before running it,
// but it will also run automatically (if not ran before) when the chain is processing it's first item.
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

func (t transformer) resetStates() {
	for _, reset := range t.stateResets {
		reset()
	}
}
