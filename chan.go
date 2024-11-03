package steps

import (
	"fmt"
	"reflect"

	"github.com/domahidizoltan/go-steps/internal/pkg/step"
	c "github.com/domahidizoltan/go-steps/kind/chansteps"
)

type chanInput[T any] <-chan T

type chanTransformator[T any] struct {
	in chanInput[T]
	step.Transformator
}

func TransformChan[T any](in <-chan T) chanInput[T] {
	return chanInput[T](in)
}

func (i chanInput[T]) With(steps ...c.AnyStep) chanTransformator[T] {
	anySteps := make([]any, 0, len(steps))
	for _, s := range steps {
		anySteps = append(anySteps, s)
	}

	t := chanTransformator[T]{
		in: i,

		Transformator: step.Transformator{
			ID:        step.CreateCacheID(),
			StepsType: reflect.TypeOf(steps[0]),
			Steps:     anySteps,
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
