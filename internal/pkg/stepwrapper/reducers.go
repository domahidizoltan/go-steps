package stepwrapper

import (
	"reflect"

	"github.com/domahidizoltan/go-steps/internal/pkg/step"
)

func GroupBy[IN0 any, OUT0 comparable, OUT1 any](fn func(in IN0) (OUT0, OUT1, error)) step.ReducerWrapper {
	type mapType map[OUT0][]OUT1
	acc := mapType{}
	return step.ReducerWrapper{
		InTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[mapType](),
		},
		ReducerFn: func(in step.StepInput) step.StepOutput {
			groupKey, value, err := fn(in.Args[0].(IN0))
			acc[groupKey] = append(acc[groupKey], value)
			return step.StepOutput{
				Args:    [step.MaxArgs]any{acc},
				ArgsLen: 1,
				Error:   err,
				Skip:    true,
			}
		},
	}
}
