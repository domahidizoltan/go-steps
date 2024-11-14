package chans

import (
	"reflect"

	"github.com/domahidizoltan/go-steps/internal/pkg/step"
	"github.com/domahidizoltan/go-steps/types"
)

type (
	tempSteps struct {
		error        error
		inType       reflect.Type
		stepWrappers []types.StepWrapper
		steps        []types.StepFn
	}

	input[T any] <-chan T

	transformator[T any] struct {
		in input[T]
		step.Transformator
	}
)

// TODO can I proxy some of these functions as well?
func Steps(s ...types.StepWrapper) tempSteps {
	// fmt.Println("addSteps")
	return tempSteps{
		stepWrappers: s,
	}
}

func (s *tempSteps) Validate() error {
	if s.error != nil {
		return s.error
	}

	s.steps, s.error = step.GetValidatedSteps[tempSteps](s.stepWrappers)
	return s.error
}

func Transform[T any](in <-chan T) input[T] {
	return input[T](in)
}

// TODO allow adding external steps

func (i input[T]) WithSteps(steps ...types.StepWrapper) transformator[T] {
	return i.With(Steps(steps...))
}

func (i input[T]) With(steps tempSteps) transformator[T] {
	t := transformator[T]{
		Transformator: step.Transformator{
			Error: steps.Validate(),
		},
	}

	if t.Error != nil {
		return t
	}

	// TODO in type must match first step first input type
	t.in = i
	t.Steps = append(t.Steps, steps.steps...)
	return t
}

func (t transformator[T]) AsRange() (func(yield func(i any) bool), error) {
	// fmt.Println("chan AsRange")

	if t.Error != nil {
		return nil, t.Error
	}

	return func(yield func(i any) bool) {
		for i := range t.in {
			if step.Process(i, yield, t.Steps) {
				break
			}
		}
	}, nil
}
