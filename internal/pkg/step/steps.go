package step

import (
	"reflect"

	"github.com/domahidizoltan/go-steps/types"
)

func Map[U, V any](fn func(in U) (V, error)) types.StepWrapper {
	return types.StepWrapper{
		InTypes:  [types.MaxArgs]reflect.Type{reflect.TypeFor[U]()},
		OutTypes: [types.MaxArgs]reflect.Type{reflect.TypeFor[V]()},
		StepFn: func(in types.StepInput) types.StepOutput {
			out, err := fn(in.Args[0].(U))
			return types.StepOutput{
				Args:    [types.MaxArgs]any{out},
				ArgsLen: 1,
				Error:   err,
				Skip:    err != nil,
			}
		},
	}
}

func Filter[U any](fn func(in U) (bool, error)) types.StepWrapper {
	return types.StepWrapper{
		InTypes:  [types.MaxArgs]reflect.Type{reflect.TypeFor[U]()},
		OutTypes: [types.MaxArgs]reflect.Type{reflect.TypeFor[U]()},
		StepFn: func(in types.StepInput) types.StepOutput {
			ok, err := fn(in.Args[0].(U))
			return types.StepOutput{
				Args:    [types.MaxArgs]any{in.Args[0].(U)},
				ArgsLen: 1,
				Error:   err,
				Skip:    ok,
			}
		},
	}
}

// TODO Added only for testing purposes

func MultiplyBy[U ~int](multiplier U) types.StepWrapper {
	return types.StepWrapper{
		InTypes:  [types.MaxArgs]reflect.Type{reflect.TypeFor[U]()},
		OutTypes: [types.MaxArgs]reflect.Type{reflect.TypeFor[U]()},
		StepFn: func(in types.StepInput) types.StepOutput {
			return types.StepOutput{
				Args:    [types.MaxArgs]any{in.Args[0].(U) * multiplier},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
	}
}
