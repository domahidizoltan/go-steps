package steps

import (
	"fmt"
	"math"
	"reflect"
)

// GroupBy is an aggregator grouping inputs by comparable values.
// The grouped values are in a slice.
func GroupBy[IN0 any, OUT0 comparable, OUT1 any](fn func(in IN0) (OUT0, OUT1, error)) ReducerWrapper {
	acc := map[OUT0][]OUT1{}
	return ReducerWrapper{
		Name: "GroupBy",
		ReducerFn: func(in StepInput) StepOutput {
			groupKey, value, err := fn(in.Args[0].(IN0))
			acc[groupKey] = append(acc[groupKey], value)
			return StepOutput{
				Args:    Args{acc},
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
			return ArgTypes{reflect.TypeFor[OUT0](), reflect.TypeFor[OUT1]()}, nil
		},
		Reset: func() {
			acc = map[OUT0][]OUT1{}
		},
	}
}

// Fold reduces a series of inputs into a single value using a custom initial value.
func Fold[IN0 any](initValue IN0, reduceFn func(in1, in2 IN0) (IN0, error)) ReducerWrapper {
	prevValue := initValue
	return ReducerWrapper{
		Name: "Fold",
		ReducerFn: func(in StepInput) StepOutput {
			currentVal := in.Args[0].(IN0)
			nextValue, err := reduceFn(prevValue, currentVal)
			prevValue = nextValue
			return StepOutput{
				Args:    Args{nextValue},
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
			return ArgTypes{reflect.TypeFor[IN0]()}, nil
		},
		Reset: func() {
			prevValue = initValue
		},
	}
}

// Reduce reduces a series of inputs into a single value using a zero value as initial value.
func Reduce[IN0 any](fn func(in1, in2 IN0) (IN0, error)) ReducerWrapper {
	var initValue IN0
	fold := Fold(initValue, fn)
	fold.Name = "Reduce"
	return fold
}

type number interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

// Sum returns the total of all number inputs
func Sum[IN0 number]() ReducerWrapper {
	sumFn := Reduce(func(in1, in2 IN0) (IN0, error) {
		return in1 + in2, nil
	})
	sumFn.Name = "Sum"
	return sumFn
}

// Max returns the largest number input
func Max[IN0 number]() ReducerWrapper {
	var minValue any
	switch reflect.TypeFor[IN0]().Kind() {
	case reflect.Int:
		minValue = int(math.MinInt)
	case reflect.Int8:
		minValue = int8(math.MinInt8)
	case reflect.Int16:
		minValue = int16(math.MinInt16)
	case reflect.Int32:
		minValue = int32(math.MinInt32)
	case reflect.Int64:
		minValue = int64(math.MinInt64)
	case reflect.Uint:
		minValue = uint(0)
	case reflect.Uint8:
		minValue = uint8(0)
	case reflect.Uint16:
		minValue = uint16(0)
	case reflect.Uint32:
		minValue = uint32(0)
	case reflect.Uint64:
		minValue = uint64(0)
	case reflect.Float32:
		minValue = -float32(math.MaxFloat32)
	case reflect.Float64:
		minValue = -float64(math.MaxFloat64)
	}
	maxFn := Fold(minValue.(IN0), func(in1, in2 IN0) (IN0, error) {
		if in1 > in2 {
			return in1, nil
		}
		return in2, nil
	})
	maxFn.Name = "Max"
	return maxFn
}

// Min returns the smallest number input
func Min[IN0 number]() ReducerWrapper {
	var maxValue any
	switch reflect.TypeFor[IN0]().Kind() {
	case reflect.Int:
		maxValue = int(math.MaxInt)
	case reflect.Int8:
		maxValue = int8(math.MaxInt8)
	case reflect.Int16:
		maxValue = int16(math.MaxInt16)
	case reflect.Int32:
		maxValue = int32(math.MaxInt32)
	case reflect.Int64:
		maxValue = int64(math.MaxInt64)
	case reflect.Uint:
		maxValue = uint(math.MaxUint)
	case reflect.Uint8:
		maxValue = uint8(math.MaxUint8)
	case reflect.Uint16:
		maxValue = uint16(math.MaxUint16)
	case reflect.Uint32:
		maxValue = uint32(math.MaxUint32)
	case reflect.Uint64:
		maxValue = uint64(math.MaxUint64)
	case reflect.Float32:
		maxValue = float32(math.MaxFloat32)
	case reflect.Float64:
		maxValue = float64(math.MaxFloat64)
	}

	minFn := Fold(maxValue.(IN0), func(in1, in2 IN0) (IN0, error) {
		if in1 < in2 {
			return in1, nil
		}
		return in2, nil
	})
	minFn.Name = "Min"
	return minFn
}

// Avg returns the average of all float64 inputs
func Avg() ReducerWrapper {
	var counter, avg, sum float64

	avgFn := Reduce(func(in1, in2 float64) (float64, error) {
		in1 = in1 * counter
		sum = in1 + in2
		counter++
		avg = float64(sum) / counter

		return avg, nil
	})
	avgFn.Name = "Avg"
	return avgFn
}
