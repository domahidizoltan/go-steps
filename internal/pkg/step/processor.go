package step

import (
	"fmt"
	"reflect"
)

// TODO refactor validation: each step should validate itself by getting the output type of the previous step
func GetValidatedSteps[T any](stepWrappers []StepWrapper) ([]StepFn, error) {
	transformInputType := reflect.TypeFor[T]()
	return getValidatedSteps(transformInputType, stepWrappers)
}

func getValidatedSteps(transformInputType reflect.Type, stepWrappers []StepWrapper) ([]StepFn, error) {
	if len(stepWrappers) == 0 {
		return nil, nil
	}
	validSteps := make([]StepFn, 0, len(stepWrappers))

	firstInType := stepWrappers[0].InTypes[0]
	if firstInType == nil {
		return nil, fmt.Errorf("%w: %s step in type is missing", ErrEmptyFirstStepInType, stepWrappers[0].Name)
	}
	if transformInputType != firstInType {
		return nil, fmt.Errorf("%w: %s transform input not equals with %s %s step in type", ErrTransformInputTypeIsDifferent, transformInputType.String(), firstInType.String(), stepWrappers[0].Name)
	}

	outTypes := [maxArgs]reflect.Type{}
	for i := range maxArgs {
		outTypes[i] = reflect.Zero(firstInType).Type()
	}

	for pos, wrapper := range stepWrappers {
		if wrapper.OutTypes[0] == nil {
			return nil, fmt.Errorf("%w: %s step at position %d has no out type", ErrEmptyStepOutType, wrapper.Name, pos+1)
		}

		for i := range maxArgs {
			// handle split and merge
			if outTypes[i] != wrapper.OutTypes[i] {
				return nil, fmt.Errorf("%w: %s step at position %d has %s out type but %s step at position %d expects %s in type",
					ErrStepOutAndNextInTypeIsDifferent, stepWrappers[pos-1].Name, pos, outTypes[i].String(), wrapper.Name, pos+1, wrapper.InTypes[i].String())
			}

			outTypes = wrapper.OutTypes
		}

		if wrapper.InnerValidator != nil {
			if branchType != wrapper.InTypes[0] {
				// TODO
				continue
			}
			branchInType := reflect.TypeOf(wrapper.InTypeZeroValue)
			if err := wrapper.InnerValidator(branchInType); err != nil {
				return nil, fmt.Errorf("%w: %s step at position %d failed inner validation: %s", ErrInnerStepValidationFailed, wrapper.Name, pos+1, err.Error())
			}
		}

		validSteps = append(validSteps, StepFn(wrapper.StepFn))
	}

	return validSteps, nil
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
