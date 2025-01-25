package customwrapper

import (
	"reflect"

	s "github.com/domahidizoltan/go-steps"
)

func MultiplyBy[IN0 ~int](multiplier IN0) s.StepWrapper {
	return s.StepWrapper{
		Name: "MultiplyBy",
		StepFn: func(in s.StepInput) s.StepOutput {
			return s.StepOutput{
				Args:    s.Args{in.Args[0].(IN0) * multiplier},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
		Validate: func(prevStepOut s.ArgTypes) (s.ArgTypes, error) {
			return s.ArgTypes{reflect.TypeFor[IN0]()}, nil
		},
	}
}
