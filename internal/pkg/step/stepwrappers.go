package step

import (
	"fmt"
	"reflect"
)

func Map[IN0, OUT0 any](fn func(in IN0) (OUT0, error)) StepWrapper {
	return StepWrapper{
		Name: "Map",
		InTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[OUT0](),
		},
		StepFn: func(in StepInput) StepOutput {
			out, err := fn(in.Args[0].(IN0))
			return StepOutput{
				Args:    [maxArgs]any{out},
				ArgsLen: 1,
				Error:   err,
				Skip:    err != nil,
			}
		},
	}
}

func Filter[IN0 any](fn func(in IN0) (bool, error)) StepWrapper {
	return StepWrapper{
		Name: "Filter",
		InTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		StepFn: func(in StepInput) StepOutput {
			ok, err := fn(in.Args[0].(IN0))
			return StepOutput{
				Args:    [maxArgs]any{in.Args[0].(IN0)},
				ArgsLen: 1,
				Error:   err,
				Skip:    !ok,
			}
		},
	}
}

type branch struct {
	value any
	T     reflect.Value // reflect.Type
	key   uint8
}

func Equaler(kv, other branch) bool {
	fmt.Println(kv, other)
	if kv == other {
		return true
	}
	if kv.key == other.key && kv.value == other.value {
		return true
	}
	return false
}

var branchType = reflect.TypeFor[branch]()

func Split[IN0 any, OUT0 ~uint8](fn func(in IN0) (OUT0, error)) StepWrapper {
	return StepWrapper{
		Name: "Split",
		InTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [maxArgs]reflect.Type{branchType},
		StepFn: func(in StepInput) StepOutput {
			idx, err := fn(in.Args[0].(IN0))
			out := branch{key: uint8(idx), value: in.Args[0].(IN0), T: reflect.ValueOf(in.Args[0])}
			return StepOutput{
				Args:    [maxArgs]any{out},
				ArgsLen: 1,
				Error:   err,
			}
		},
	}
}

// TODO document that this is not parallel processing, because it keeps ordering
func WithBranches[IN0 any](steps ...StepsContainer) StepWrapper {
	var inTypeZeroValue IN0
	return StepWrapper{
		Name:     "WithBranches",
		InTypes:  [maxArgs]reflect.Type{branchType},
		OutTypes: [maxArgs]reflect.Type{branchType},
		StepFn: func(in StepInput) StepOutput {
			keyVal := in.Args[0].(branch)
			val := StepInput{
				Args:    [maxArgs]any{keyVal.value},
				ArgsLen: 1,
			}
			var out StepOutput
			for _, stepWrapper := range steps[int(keyVal.key)].StepWrappers {
				out = stepWrapper.StepFn(val)
				val = StepInput{
					Args:    out.Args,
					ArgsLen: out.ArgsLen,
				}
				if out.Skip || out.Error != nil {
					break
				}
			}
			return StepOutput{
				Args:    [maxArgs]any{branch{key: keyVal.key, value: out.Args[0], T: reflect.ValueOf(out.Args[0])}},
				ArgsLen: 1,
				Error:   out.Error,
				Skip:    out.Skip,
			}
		},
		InTypeZeroValue: inTypeZeroValue,
		InnerValidator: func(firstInType reflect.Type) error {
			for _, stepWrapper := range steps {
				if _, err := getValidatedSteps(firstInType, stepWrapper.StepWrappers); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func Merge() StepWrapper {
	return StepWrapper{
		Name: "Merge",
		InTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[branch](),
		},
		OutTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[any](), // should not be any
		},
		StepFn: func(in StepInput) StepOutput {
			return StepOutput{
				Args:    [maxArgs]any{in.Args[0].(branch).value},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
	}
}

// TODO Added only for testing purposes

func MultiplyBy[IN0 ~int](multiplier IN0) StepWrapper {
	return StepWrapper{
		Name: "MultiplyBy",
		InTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [maxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		StepFn: func(in StepInput) StepOutput {
			return StepOutput{
				Args:    [maxArgs]any{in.Args[0].(IN0) * multiplier},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
	}
}
