package steps

import (
	"fmt"

	"github.com/domahidizoltan/go-steps/internal/pkg/step"
)

type sliceInput[T any] []T

type sliceTransformator[T any] struct {
	in sliceInput[T]
	step.Transformator
}

func TransformSlice[T any](in []T) sliceInput[T] {
	return sliceInput[T](in)
}

func (i sliceInput[T]) With(steps ...any) sliceTransformator[T] {
	t := sliceTransformator[T]{
		in: i,
		Transformator: step.Transformator{
			ID:    step.CreateCacheID(),
			Steps: steps,
		},
	}

	step.ValidateSteps[T](&t.Transformator)
	return t
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
