package step

import (
	"fmt"
	"reflect"

	"github.com/domahidizoltan/go-steps/types"
)

func GetValidatedSteps[T any](stepWrappers []types.StepWrapper) ([]types.StepFn, error) {
	if len(stepWrappers) == 0 {
		return nil, nil
	}
	validSteps := make([]types.StepFn, 0, len(stepWrappers))

	inType := stepWrappers[0].InTypes[0]
	outTypes := [types.MaxArgs]reflect.Type{}
	for i := 0; i < types.MaxArgs; i++ {
		outTypes[i] = reflect.Zero(inType).Type()
	}

	for pos, wrapper := range stepWrappers {
		stepType := reflect.TypeOf(wrapper.StepFn)
		// fmt.Printf("stepType %s \ntransfType %s", stepType.PkgPath(), transformatorTypePkg)
		// if stepType.PkgPath() != transformatorTypePkg {
		// 	return fmt.Errorf("%w: [pos %d.] %s", ErrInvalidStepType, pos, stepType.Name())
		// }

		for i := 0; i < len(wrapper.OutTypes); i++ {
			if outTypes[i] != wrapper.InTypes[i] {
				return nil, fmt.Errorf("%w: [pos %d.] %s", ErrInvalidInputType, pos, stepType.Name())
			}
			outTypes = wrapper.OutTypes
		}

		validSteps = append(validSteps, types.StepFn(wrapper.StepFn))
	}

	return validSteps, nil
}

func Process[T any](i T, yield func(in any) bool, fns []types.StepFn) bool {
	in := types.StepInput{
		Args:    [4]any{i},
		ArgsLen: 1,
	}
	var skipped bool
	for _, fn := range fns {
		out := fn(in)

		if out.Skip || out.Error != nil {
			skipped = true
			break
		}

		in = types.StepInput{
			Args:    out.Args,
			ArgsLen: out.ArgsLen,
		}
	}

	if !skipped && !yield(in.Args[0]) {
		return true
	}

	return false
}
