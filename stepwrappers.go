package steps

import (
	"fmt"
	"reflect"
	"strings"
)

// Map transforms a single input into a single output
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

func simpleFilterValidation[IN0 any](prevStepOut ArgTypes) (ArgTypes, error) {
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
}

// Filter skips inputs that do not pass the filter
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
		Validate: simpleFilterValidation[IN0],
	}
}

// Take is processing the first N inputs
func Take[IN0 any](count uint64) StepWrapper {
	var counter uint64
	return StepWrapper{
		Name: "Take",
		StepFn: func(in StepInput) StepOutput {
			skip := counter >= count
			counter++
			return StepOutput{
				Args:    Args{in.Args[0].(IN0)},
				ArgsLen: 1,
				Skip:    skip,
			}
		},
		Validate: simpleFilterValidation[IN0],
		Reset: func() {
			counter = 0
		},
	}
}

// TakeWhile processes inputs while the filter returns true
func TakeWhile[IN0 any](fn func(in IN0) (bool, error)) StepWrapper {
	var skip bool
	return StepWrapper{
		Name: "TakeWhile",
		StepFn: func(in StepInput) StepOutput {
			ok, err := fn(in.Args[0].(IN0))
			if !skip && !ok {
				skip = true
			}
			return StepOutput{
				Args:    Args{in.Args[0].(IN0)},
				ArgsLen: 1,
				Error:   err,
				Skip:    skip,
			}
		},
		Validate: simpleFilterValidation[IN0],
		Reset: func() {
			skip = false
		},
	}
}

// Skip is skipping the first N inputs
func Skip[IN0 any](count uint64) StepWrapper {
	var counter uint64
	return StepWrapper{
		Name: "Skip",
		StepFn: func(in StepInput) StepOutput {
			skip := counter < count
			counter++
			return StepOutput{
				Args:    Args{in.Args[0].(IN0)},
				ArgsLen: 1,
				Skip:    skip,
			}
		},
		Validate: simpleFilterValidation[IN0],
		Reset: func() {
			counter = 0
		},
	}
}

// SkipWhile skips processing inputs while the filter returns true
func SkipWhile[IN0 any](fn func(in IN0) (bool, error)) StepWrapper {
	skip := true
	return StepWrapper{
		Name: "SkipWhile",
		StepFn: func(in StepInput) StepOutput {
			ok, err := fn(in.Args[0].(IN0))
			if skip && !ok {
				skip = false
			}
			return StepOutput{
				Args:    Args{in.Args[0].(IN0)},
				ArgsLen: 1,
				Error:   err,
				Skip:    skip,
			}
		},
		Validate: simpleFilterValidation[IN0],
		Reset: func() {
			skip = true
		},
	}
}

// Do runs a function on each input item
func Do[IN0 any](fn func(in IN0) error) StepWrapper {
	return StepWrapper{
		Name: "Do",
		StepFn: func(in StepInput) StepOutput {
			err := fn(in.Args[0].(IN0))
			return StepOutput{
				Args:    in.Args,
				ArgsLen: in.ArgsLen,
				Error:   err,
			}
		},
		Validate: simpleFilterValidation[IN0],
	}
}

// Log logs debug informations between steps
func Log(prefix ...string) StepWrapper {
	return StepWrapper{
		Name: "Log",
		StepFn: func(in StepInput) StepOutput {
			opts := in.TransformerOptions
			out := strings.Builder{}
			var err error
			if len(prefix) != 0 {
				_, err = out.WriteString(prefix[0] + " ")
				if err != nil {
					goto writeErr
				}
			}
			if opts.Name != "" {
				_, err = out.WriteString("transformer:" + opts.Name + " ")
				if err != nil {
					goto writeErr
				}
			}
			for i := range in.ArgsLen {
				_, err = out.WriteString(fmt.Sprintf("\targ%d: %v ", i, in.Args[i]))
				if err != nil {
					goto writeErr
				}
			}
			_, err = fmt.Fprintln(in.TransformerOptions.LogWriter, out.String())

		writeErr:
			return StepOutput{
				Args:    in.Args,
				ArgsLen: in.ArgsLen,
				Error:   err,
			}
		},
		Validate: func(prevStepOut ArgTypes) (ArgTypes, error) {
			return prevStepOut, nil
		},
	}
}

type branch struct {
	value any
	T     reflect.Value
	key   uint8
}

var branchType = reflect.TypeFor[branch]()

// Split defines how to split inputs into transformation branches
// The function returns the number of the branch where the input will be sent
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

// WithBranches applies a set of steps to each branch.
// This is not parallel processing. The items keeps the order even if one branch could possibly process faster.
func WithBranches[IN0 any](stepsBranches ...StepsBranch) StepWrapper {
	return StepWrapper{
		Name: "WithBranches",
		StepFn: func(in StepInput) StepOutput {
			keyVal := in.Args[0].(branch)
			val := StepInput{
				Args:               Args{keyVal.value},
				ArgsLen:            1,
				TransformerOptions: in.TransformerOptions,
			}
			var out StepOutput
			for _, stepWrapper := range stepsBranches[int(keyVal.key)].StepWrappers {
				out = stepWrapper.StepFn(val)
				val = StepInput{
					Args:               out.Args,
					ArgsLen:            out.ArgsLen,
					TransformerOptions: in.TransformerOptions,
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

// Merge merges back the transformation branches
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
