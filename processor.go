package steps

import (
	"fmt"
	"reflect"
)

func getValidatedSteps[T any](stepWrappers []StepWrapper) ([]StepFn, ArgTypes, error) {
	if len(stepWrappers) == 0 {
		return nil, ArgTypes{}, nil
	}
	validSteps := make([]StepFn, 0, len(stepWrappers))

	transformInputType := reflect.TypeFor[T]()
	outTypes := ArgTypes{transformInputType}

	for pos, wrapper := range stepWrappers {
		ot, err := wrapper.Validate(outTypes)
		if err != nil {
			return nil, ArgTypes{}, fmt.Errorf("%w [%s:%d]: %w", ErrStepValidationFailed, stepWrappers[pos].Name, pos+1, err)
		}
		outTypes = ot

		validSteps = append(validSteps, StepFn(wrapper.StepFn))
	}

	return validSteps, outTypes, nil
}

func process[V any](val V, yield func(any) bool, transformator *transformator, isLastItem bool) (bool, bool, error) {
	if transformator == nil {
		return false, !yield(val), nil
	}

	out, skipped := getProcessResult(val, transformator)
	if out.Error != nil {
		return false, false, out.Error
	}

	aggOut := transformator.lastAggregatedValue
	if aggOut != nil {
		if aggOut.Error != nil {
			return false, false, aggOut.Error
		}
		if isLastItem {
			return false, yield(aggOut.Args[0]), nil
		}
	}

	if skipped {
		return true, false, nil
	}
	return false, !yield(out.Args[0]), nil
}

func processIndexed[V any](key any, val V, yield func(any, any) bool, transformator *transformator, isLastItem bool) (bool, bool, error) {
	if transformator == nil {
		return false, !yield(key, val), nil
	}

	out, skipped := getProcessResult(val, transformator)
	if out.Error != nil {
		return false, false, out.Error
	}

	idx := key
	v := out.Args[0]
	if out.ArgsLen > 1 {
		idx = out.Args[0]
		v = out.Args[1]
	}

	aggOut := transformator.lastAggregatedValue
	if aggOut != nil {
		if aggOut.Error != nil {
			return false, false, aggOut.Error
		}
		if isLastItem {
			return false, yield(idx, aggOut.Args[0]), nil
		}
	}

	if skipped {
		return true, false, nil
	}
	return false, !yield(idx, v), nil
}

func getProcessResult[V any](val V, transformator *transformator) (StepOutput, bool) {
	if transformator == nil || len(transformator.steps) == 0 {
		return StepOutput{
			Args:    [4]any{val},
			ArgsLen: 1,
		}, false
	}

	in := StepInput{
		Args:    [4]any{val},
		ArgsLen: 1,
	}
	var skipped bool
	var out StepOutput
	for _, fn := range transformator.steps {
		out = fn(in)
		if out.Skip || out.Error != nil {
			skipped = true
			break
		}

		in = StepInput{
			Args:    out.Args,
			ArgsLen: out.ArgsLen,
		}
	}

	if !skipped && transformator.aggregator != nil {
		out = transformator.aggregator(in)
		transformator.lastAggregatedValue = &out
		skipped = true
	}

	return out, skipped
}
