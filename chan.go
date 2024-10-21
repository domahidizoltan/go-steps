package steps

import (
	"fmt"

	"github.com/domahidizoltan/go-steps/internal/pkg/step"
)

type chanInput[T any] <-chan T

type chanTransformator[T any] struct {
	in chanInput[T]
	step.Transformator
}

func TransformChan[T any](in <-chan T) chanInput[T] {
	return chanInput[T](in)
}

func (i chanInput[T]) With(steps ...any) chanTransformator[T] {
	t := chanTransformator[T]{
		in: i,
		Transformator: step.Transformator{
			ID:    step.CreateCacheID(),
			Steps: steps,
		},
	}

	step.ValidateSteps[T](&t.Transformator)
	return t
}

func (t chanTransformator[T]) AsRange() (func(yield func(i any) bool), error) {
	fmt.Println("chan AsRange")

	if t.Err != nil {
		delete(step.FnCache, t.ID)
		return nil, t.Err
	}

	fns := step.FnCache[t.ID]
	return func(yield func(i any) bool) {
		for i := range t.in {
			if step.Process(i, yield, fns) {
				break
			}
		}
	}, nil
}
