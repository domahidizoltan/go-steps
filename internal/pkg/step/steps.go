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

type kv struct {
	key   uint8
	value any
	t     reflect.Type
}

func Split[IN0 any, OUT0 ~uint8](fn func(in IN0) (OUT0, error)) types.StepWrapper {
	return types.StepWrapper{
		InTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[kv](),
		},
		StepFn: func(in types.StepInput) types.StepOutput {
			idx, err := fn(in.Args[0].(IN0))
			out := kv{uint8(idx), in.Args[0].(IN0), reflect.TypeOf(in.Args[0])}
			return types.StepOutput{
				Args:    [types.MaxArgs]any{out},
				ArgsLen: 1,
				Error:   err,
			}
		},
	}
}

func WithBranches[IN0 any](steps ...TempSteps) types.StepWrapper {
	return types.StepWrapper{
		InTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[kv](),
		},
		OutTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[kv](),
		},
		StepFn: func(in types.StepInput) types.StepOutput {
			keyVal := in.Args[0].(kv)
			val := types.StepInput{
				Args:    [types.MaxArgs]any{keyVal.value},
				ArgsLen: 1,
			}
			var out types.StepOutput
			for _, stepWrapper := range steps[int(keyVal.key)].StepWrappers {
				out = stepWrapper.StepFn(val)
				val = types.StepInput{
					Args:    out.Args,
					ArgsLen: out.ArgsLen,
				}
				if out.Skip || out.Error != nil {
					break
				}
			}
			return types.StepOutput{
				Args:    [types.MaxArgs]any{kv{keyVal.key, out.Args[0], reflect.TypeOf(out.Args[0])}},
				ArgsLen: 1,
				Error:   out.Error,
				Skip:    out.Skip,
			}
		},
	}
}

func Zip() types.StepWrapper {
	return types.StepWrapper{
		InTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[kv](),
		},
		OutTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[any](), // should not be any
		},
		StepFn: func(in types.StepInput) types.StepOutput {
			return types.StepOutput{
				Args:    [types.MaxArgs]any{in.Args[0].(kv).value},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
	}
}

func GroupBy[IN0 any, OUT0 comparable, OUT1 any](fn func(in IN0) (OUT0, OUT1, error)) types.ReducerWrapper {
	type mapType map[OUT0][]OUT1
	acc := mapType{}
	return types.ReducerWrapper{
		InTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [types.MaxArgs]reflect.Type{
			reflect.TypeFor[mapType](),
		},
		ReducerFn: func(in types.StepInput) types.StepOutput {
			groupKey, value, err := fn(in.Args[0].(IN0))
			acc[groupKey] = append(acc[groupKey], value)
			return types.StepOutput{
				Args:    [types.MaxArgs]any{acc},
				ArgsLen: 1,
				Error:   err,
				Skip:    true,
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
