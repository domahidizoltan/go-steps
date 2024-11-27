package step

import (
	"reflect"

	"github.com/domahidizoltan/go-steps/types"
)

func Map[IN0, OUT0 any](fn func(in IN0) (OUT0, error)) types.StepWrapper {
	return types.StepWrapper{
		InTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[OUT0](),
		},
		StepFn: func(in types.StepInput) types.StepOutput {
			out, err := fn(in.Args[0].(IN0))
			return types.StepOutput{
				Args:    [types.MaxArgs]any{out},
				ArgsLen: 1,
				Error:   err,
				Skip:    err != nil,
			}
		},
	}
}

func Filter[IN0 any](fn func(in IN0) (bool, error)) types.StepWrapper {
	return types.StepWrapper{
		InTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		StepFn: func(in types.StepInput) types.StepOutput {
			ok, err := fn(in.Args[0].(IN0))
			return types.StepOutput{
				Args:    [types.MaxArgs]any{in.Args[0].(IN0)},
				ArgsLen: 1,
				Error:   err,
				Skip:    !ok,
			}
		},
	}
}

func GroupBy[IN0 any, OUT0 comparable, OUT1 any](fn func(in IN0) (OUT0, OUT1, error)) types.ReducerWrapper {
	type mapType map[OUT0][]OUT1
	return types.ReducerWrapper{
		InTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[mapType](),
		},
		ReducerFn: func(in types.StepInput) types.StepOutput {
			groupKey, value, err := fn(in.Args[0].(IN0))
			return types.StepOutput{
				Args:    [types.MaxArgs]any{groupKey, value},
				ArgsLen: 2,
				Error:   err,
			}
		},
	}
}

// TODO Added only for testing purposes

func MultiplyBy[IN0 ~int](multiplier IN0) types.StepWrapper {
	return types.StepWrapper{
		InTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		StepFn: func(in types.StepInput) types.StepOutput {
			return types.StepOutput{
				Args:    [types.MaxArgs]any{in.Args[0].(IN0) * multiplier},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
	}
}
