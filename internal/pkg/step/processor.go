package step

import (
	"fmt"
	"reflect"
)

func GetValidatedSteps[T any](stepWrappers []StepWrapper) ([]StepFn, error) {
	if len(stepWrappers) == 0 {
		return nil, nil
	}
	validSteps := make([]StepFn, 0, len(stepWrappers))

	inType := stepWrappers[0].InTypes[0]
	outTypes := [MaxArgs]reflect.Type{}
	for i := 0; i < MaxArgs; i++ {
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
				return nil, fmt.Errorf("%w: [pos %d.] %s : %s != %s", ErrInvalidInputType, pos, stepType.Name(), outTypes[i].Kind().String(), wrapper.InTypes[i].Kind().String())
			}
			outTypes = wrapper.OutTypes
		}

		validSteps = append(validSteps, StepFn(wrapper.StepFn))
	}

	return validSteps, nil
}

func getProcessResult[V any](v V, transformator *Transformator) (StepInput, bool) {
	in := StepInput{
		Args:    [4]any{v},
		ArgsLen: 1,
	}
	var skipped bool
	for _, fn := range transformator.Steps {
		out := fn(in)
		if out.Skip || out.Error != nil {
			skipped = true
			break
		}

		in = StepInput{
			Args:    out.Args,
			ArgsLen: out.ArgsLen,
		}
	}

	if !skipped && transformator.Aggregator != nil {
		out := transformator.Aggregator(in)
		transformator.LastAggregatedValue = &out

		if out.Error == nil {
			in = StepInput{
				Args:    out.Args,
				ArgsLen: out.ArgsLen,
			}
		}
		skipped = true
	}

	return in, skipped
}

func Process[V any](v V, yield func(any) bool, transformator *Transformator, isLastItem bool) bool {
	in, skipped := getProcessResult(v, transformator)

	if transformator.LastAggregatedValue != nil && isLastItem {
		return yield(transformator.LastAggregatedValue.Args[0])
	}

	if !skipped && !yield(in.Args[0]) {
		return true
	}

	return false
}

func ProcessIndexed[V any](k any, v V, yield func(any, any) bool, transformator *Transformator, isLastItem bool) bool {
	in, skipped := getProcessResult(v, transformator)

	idx := k
	val := in.Args[0]
	if in.ArgsLen > 1 {
		idx = in.Args[0]
		val = in.Args[1]
	}

	if transformator.LastAggregatedValue != nil && isLastItem {
		return yield(idx, transformator.LastAggregatedValue.Args[0])
	}

	if !skipped && !yield(idx, val) {
		return true
	}

	return false
}
