package steps

import (
	"fmt"
	"reflect"
)

func Map[IN0, OUT0 any](fn func(in IN0) (OUT0, error)) StepWrapper {
	return StepWrapper{
		Name: "Map",
		StepFn: func(in StepInput) StepOutput {
			out, err := fn(in.Args[0].(IN0))
			return StepOutput{
				Args:    Args{out},
				ArgsLen: 1,
				Error:   err,
			}
		},
		Validate: func(prevStepOut ArgTypes) (ArgTypes, error) {
			inArgTypes := ArgTypes{reflect.TypeFor[IN0]()}
			for i := range maxArgs {
				if i == 0 && prevStepOut[0] == reflect.TypeFor[SkipFirstArgValidation]() {
					continue
				}
				if prevStepOut[i] != inArgTypes[i] {
					return ArgTypes{}, fmt.Errorf("%w [%s!=%s:%d]", ErrIncompatibleInArgType, prevStepOut[i].String(), inArgTypes[i].String(), i+1)
				}
			}
			return ArgTypes{reflect.TypeFor[OUT0]()}, nil
		},
	}
}

func Filter[IN0 any](fn func(in IN0) (bool, error)) StepWrapper {
	return StepWrapper{
		Name: "Filter",
		StepFn: func(in StepInput) StepOutput {
			ok, err := fn(in.Args[0].(IN0))
			return StepOutput{
				Args:    Args{in.Args[0].(IN0)},
				ArgsLen: 1,
				Error:   err,
				Skip:    !ok,
			}
		},
		Validate: func(prevStepOut ArgTypes) (ArgTypes, error) {
			inArgTypes := ArgTypes{reflect.TypeFor[IN0]()}
			for i := range maxArgs {
				if i == 0 && prevStepOut[0] == reflect.TypeFor[SkipFirstArgValidation]() {
					continue
				}
				if prevStepOut[i] != inArgTypes[i] {
					return ArgTypes{}, fmt.Errorf("%w [%s!=%s:%d]", ErrIncompatibleInArgType, prevStepOut[i].String(), inArgTypes[i].String(), i+1)
				}
			}
			return ArgTypes{reflect.TypeFor[IN0]()}, nil
		},
	}
}

type branch struct {
	value any
	T     reflect.Value
	key   uint8
}

var branchType = reflect.TypeFor[branch]()

func Split[IN0 any, OUT0 ~uint8](fn func(in IN0) (OUT0, error)) StepWrapper {
	return StepWrapper{
		Name: "Split",
		StepFn: func(in StepInput) StepOutput {
			idx, err := fn(in.Args[0].(IN0))
			out := branch{key: uint8(idx), value: in.Args[0].(IN0), T: reflect.ValueOf(in.Args[0])}
			return StepOutput{
				Args:    Args{out},
				ArgsLen: 1,
				Error:   err,
			}
		},
		Validate: func(prevStepOut ArgTypes) (ArgTypes, error) {
			inArgTypes := ArgTypes{reflect.TypeFor[IN0]()}
			for i := range maxArgs {
				if i == 0 && prevStepOut[0] == reflect.TypeFor[SkipFirstArgValidation]() {
					continue
				}
				if prevStepOut[i] != inArgTypes[i] {
					return ArgTypes{}, fmt.Errorf("%w [%s!=%s:%d]", ErrIncompatibleInArgType, prevStepOut[i].String(), inArgTypes[i].String(), i+1)
				}
			}
			return ArgTypes{branchType}, nil
		},
	}
}

// TODO document that this is not parallel processing, because it keeps ordering
func WithBranches[IN0 any](stepsBranches ...StepsBranch) StepWrapper {
	return StepWrapper{
		Name: "WithBranches",
		StepFn: func(in StepInput) StepOutput {
			keyVal := in.Args[0].(branch)
			val := StepInput{
				Args:    Args{keyVal.value},
				ArgsLen: 1,
			}
			var out StepOutput
			for _, stepWrapper := range stepsBranches[int(keyVal.key)].StepWrappers {
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
				Args:    Args{branch{key: keyVal.key, value: out.Args[0], T: reflect.ValueOf(out.Args[0])}},
				ArgsLen: 1,
				Error:   out.Error,
				Skip:    out.Skip,
			}
		},
		Validate: func(prevStepOut ArgTypes) (ArgTypes, error) {
			inArgTypes := ArgTypes{branchType}
			for i := range maxArgs {
				if i == 0 && prevStepOut[0] == reflect.TypeFor[SkipFirstArgValidation]() {
					continue
				}
				if prevStepOut[i] != inArgTypes[i] {
					return ArgTypes{}, fmt.Errorf("%w [%s!=%s:%d]", ErrIncompatibleInArgType, prevStepOut[i].String(), inArgTypes[i].String(), i)
				}
			}
			for _, container := range stepsBranches {
				if _, _, err := getValidatedSteps[IN0](container.StepWrappers); err != nil {
					return ArgTypes{}, err
				}
			}
			return ArgTypes{branchType}, nil
		},
	}
}

func Merge() StepWrapper {
	return StepWrapper{
		Name: "Merge",
		StepFn: func(in StepInput) StepOutput {
			return StepOutput{
				Args:    Args{in.Args[0].(branch).value},
				ArgsLen: 1,
				Error:   nil,
			}
		},
		Validate: func(prevStepOut ArgTypes) (ArgTypes, error) {
			inArgTypes := ArgTypes{branchType}
			for i := range maxArgs {
				if i == 0 && prevStepOut[0] == reflect.TypeFor[SkipFirstArgValidation]() {
					continue
				}
				if prevStepOut[i] != inArgTypes[i] {
					return ArgTypes{}, fmt.Errorf("%w [%s!=%s:%d]", ErrIncompatibleInArgType, prevStepOut[i].String(), inArgTypes[i].String(), i+1)
				}
			}
			return ArgTypes{reflect.TypeFor[any]()}, nil
		},
	}
}
