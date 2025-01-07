package step

import (
	"reflect"
)

func GroupBy[IN0 any, OUT0 comparable, OUT1 any](fn func(in IN0) (OUT0, OUT1, error)) ReducerWrapper {
	type mapType map[OUT0][]OUT1
	acc := mapType{}
	return ReducerWrapper{
		InTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[mapType](),
		},
		ReducerFn: func(in StepInput) StepOutput {
			groupKey, value, err := fn(in.Args[0].(IN0))
			acc[groupKey] = append(acc[groupKey], value)
			return StepOutput{
				Args:    [maxArgs]any{acc},
				ArgsLen: 1,
				Error:   err,
				Skip:    true,
			}
		},
	}
}
