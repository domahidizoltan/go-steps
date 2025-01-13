package step

import (
	"fmt"
	"reflect"
)

func GroupBy[IN0 any, OUT0 comparable, OUT1 any](fn func(in IN0) (OUT0, OUT1, error)) ReducerWrapper {
	acc := map[OUT0][]OUT1{}
	return ReducerWrapper{
		Name: "GroupBy",
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
		Validate: func(prevStepOut ArgTypes) (ArgTypes, error) {
			inArgTypes := ArgTypes{reflect.TypeFor[IN0]()}
			for i := range maxArgs {
				if prevStepOut[i] != inArgTypes[i] {
					return ArgTypes{}, fmt.Errorf("%w [%s!=%s:%d]", ErrIncompatibleInArgType, prevStepOut[i].String(), inArgTypes[i].String(), i+1)
				}
			}
			return ArgTypes{reflect.TypeFor[OUT0](), reflect.TypeFor[OUT1]()}, nil
		},
	}
}
