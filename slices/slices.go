package slices

import (
	"fmt"
	"reflect"

	"github.com/domahidizoltan/go-steps/internal/pkg/step"
)

type (
	steps struct {
		err       error
		stps      []AnyStep
		fnTypes   []step.FnType
		validated bool
	}

	input[T any] []T

	transformator[T any] struct {
		in input[T]
		step.Transformator
	}
)

func Steps(s ...AnyStep) steps {
	return steps{
		stps: s,
	}
}

func (s *steps) Validate() error {
	if s.validated {
		return s.err
	}

	s.validated = true
	s.fnTypes, s.err = step.ValidateStepsNew(s.stps)
	return s.err
}

func Transform[T any](in []T) input[T] {
	return input[T](in)
}

func (i input[T]) WithSteps(steps ...AnyStep) transformator[T] {
	return i.With(Steps(steps...))
}

func (i input[T]) With(steps steps) transformator[T] {
	t := transformator[T]{
		Transformator: step.Transformator{
			Err: steps.Validate(),
		},
	}

	if t.Err != nil {
		return t
	}

	inType := reflect.TypeFor[T]()
	firstStepFirstInputType := steps.fnTypes[0].Type.In(0)
	if inType != firstStepFirstInputType {
		t.Err = step.ErrInvalidInputType
		return t
	}

	// TODO fill cache and truncate fnTypes
	// TODO in type must match first step first input type
	t.in = i
	t.ID = step.CreateCacheID()
	// t.StepsType = reflect.TypeFor[AnyStep]()
	t.Steps = step.ToAnySlice(steps.stps)

	step.FnCache[t.ID] = steps.fnTypes
	return t
}

func (t transformator[T]) AsRange() (func(yield func(i any) bool), error) {
	fmt.Println("slice AsRange")

	if t.Err != nil {
		delete(step.FnCache, t.ID)
		return nil, t.Err
	}

	fns := step.FnCache[t.ID]
	return func(yield func(i any) bool) {
		for _, i := range t.in {
			if step.Process(i, yield, fns) {
				break
			}
		}
	}, nil
}
