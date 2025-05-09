package steps

import (
	"fmt"
	"reflect"
	"sync"
)

func getValidatedSteps[T any](stepWrappers []StepWrapper) ([]StepFn, ArgTypes, error) {
	if len(stepWrappers) == 0 {
		return nil, ArgTypes{}, nil
	}
	validSteps := make([]StepFn, 0, len(stepWrappers))

	transformInputType := reflect.TypeFor[T]()
	outTypes := ArgTypes{transformInputType}

	for pos, wrapper := range stepWrappers {
		if len(wrapper.Name) == 0 || wrapper.StepFn == nil {
			return nil, ArgTypes{}, fmt.Errorf("%w [%s:%d]", ErrInvalidStep, wrapper.Name, pos+1)
		}

		ot, err := wrapper.Validate(outTypes)
		if err != nil {
			return nil, ArgTypes{}, fmt.Errorf("%w [%s:%d]: %w", ErrStepValidationFailed, stepWrappers[pos].Name, pos+1, err)
		}
		outTypes = ot

		validSteps = append(validSteps, StepFn(wrapper.StepFn))
	}

	return validSteps, outTypes, nil
}

func (so *StepOutput) Reset() {
	for i := 0; i < maxArgs; i++ {
		so.Args[i] = nil
	}
	so.ArgsLen = 0
	so.Error = nil
	so.Skip = false
}

var poolOut sync.Pool = sync.Pool{
	New: func() any {
		return &StepOutput{}
	},
}

func process[V any](val V, yield func(any) bool, transformer *transformer, isLastItem bool) (bool, bool, error) {
	if transformer == nil {
		return false, !yield(val), nil
	}

	out, skipped := getProcessResult(val, transformer)
	defer func() {
		out.Reset()
		poolOut.Put(out)
	}()
	if out.Error != nil {
		return false, false, out.Error
	}
	aggOut := transformer.lastAggregatedValue
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

func processIndexed[V any](key any, val V, yield func(any, any) bool, transformer *transformer, isLastItem bool) (bool, bool, error) {
	if transformer == nil {
		return false, !yield(key, val), nil
	}

	out, skipped := getProcessResult(val, transformer)
	defer func() {
		out.Reset()
		poolOut.Put(out)
	}()
	if out.Error != nil {
		return false, false, out.Error
	}

	idx := key
	v := out.Args[0]
	if out.ArgsLen > 1 {
		idx = out.Args[0]
		v = out.Args[1]
	}

	aggOut := transformer.lastAggregatedValue
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

func getProcessResult[V any](val V, transformer *transformer) (*StepOutput, bool) {
	if transformer == nil || (transformer.steps == nil && transformer.aggregator == nil) {
		out := poolOut.Get().(*StepOutput)
		out.Args = [4]any{val}
		out.ArgsLen = 1
		return out, false
	}

	in := StepInput{
		Args:               [4]any{val},
		ArgsLen:            1,
		TransformerOptions: transformer.options,
	}

	var skipped bool
	out := poolOut.Get().(*StepOutput)
iterSteps:
	for _, fn := range transformer.steps {
		select {
		case <-transformer.options.Ctx.Done():
			out.Error = transformer.options.Ctx.Err()
			return out, false
		default:
			*out = fn(in)
			if out.Skip || out.Error != nil {
				skipped = true
				break iterSteps
			}

			in = StepInput{
				Args:               out.Args,
				ArgsLen:            out.ArgsLen,
				TransformerOptions: transformer.options,
			}
		}
	}

	if !skipped && transformer.aggregator != nil {
		*out = transformer.aggregator(in)
		transformer.lastAggregatedValue = out
		skipped = true
	}

	return out, skipped
}
