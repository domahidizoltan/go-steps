package step

import (
	"fmt"
	"reflect"
)

func GetValidatedSteps[T any](stepWrappers []StepWrapper) ([]StepFn, ArgTypes, error) {
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

func Process[V any](val V, yield func(any) bool, transformator *Transformator, isLastItem bool) (bool, error) {
	out, skipped := getProcessResult(val, transformator)
	if out.Error != nil {
		return false, out.Error
	}

	aggOut := transformator.LastAggregatedValue
	if aggOut != nil {
		if aggOut.Error != nil {
			return false, aggOut.Error
		}
		if isLastItem {
			return yield(aggOut.Args[0]), nil
		}
	}

	terminated := skipped || !yield(out.Args[0])
	return terminated, nil
}

func ProcessIndexed[V any](key any, val V, yield func(any, any) bool, transformator *Transformator, isLastItem bool) (bool, error) {
	out, skipped := getProcessResult(val, transformator)
	if out.Error != nil {
		return false, out.Error
	}

	idx := key
	v := out.Args[0]
	if out.ArgsLen > 1 {
		idx = out.Args[0]
		v = out.Args[1]
	}

	aggOut := transformator.LastAggregatedValue
	if aggOut != nil {
		if aggOut.Error != nil {
			return false, aggOut.Error
		}
		if isLastItem {
			return yield(idx, aggOut.Args[0]), nil
		}
	}

	terminated := skipped || !yield(idx, v)
	return terminated, nil
}

func getProcessResult[V any](val V, transformator *Transformator) (StepOutput, bool) {
	in := StepInput{
		Args:    [4]any{val},
		ArgsLen: 1,
	}
	var skipped bool
	var out StepOutput
	for _, fn := range transformator.Steps {
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

	if !skipped && transformator.Aggregator != nil {
		out = transformator.Aggregator(in)
		transformator.LastAggregatedValue = &out
		skipped = true
	}

	return out, skipped
}
