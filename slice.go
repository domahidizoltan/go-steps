package steps

import (
	"fmt"
	"reflect"

	"github.com/domahidizoltan/go-steps/internal/pkg/step"
	s "github.com/domahidizoltan/go-steps/kind/slicesteps"
)

type sliceInput[T any] []T

type sliceTransformator[T any] struct {
	in sliceInput[T]
	step.Transformator
}

func TransformSlice[T any](in []T) sliceInput[T] {
	return sliceInput[T](in)
}

func (i sliceInput[T]) With(steps ...s.AnyStep) sliceTransformator[T] {
	t := createSliceTransformator[T](steps...)
	step.ValidateSteps[T](&t.Transformator)
	t.in = i
	return t
}

func ValidateSliceSteps(steps ...s.AnyStep) error {
	// TODO take T with reflection for first parameter type of first step
	t := createSliceTransformator[string](steps...)
	step.ValidateSteps[string](&t.Transformator)
	return t.Err
}

func createSliceTransformator[T any](steps ...s.AnyStep) sliceTransformator[T] {
	anySteps := make([]any, 0, len(steps))
	for _, s := range steps {
		anySteps = append(anySteps, s)
	}

	return sliceTransformator[T]{
		Transformator: step.Transformator{
			ID:        step.CreateCacheID(),
			StepsType: reflect.TypeFor[s.AnyStep](),
			Steps:     anySteps,
		},
	}
}

func (t sliceTransformator[T]) AsRange() (func(yield func(i any) bool), error) {
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
