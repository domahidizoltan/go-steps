package steps

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupBy_Success(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}).
		With(Aggregate(
			GroupBy(func(in int) (bool, int, error) {
				return in%2 == 0, in, nil
			}))).
		AsMultiMap(expectsError(t, false))

	expected := map[any][]any{
		true:  {2, 4},
		false: {1, 3},
	}
	assert.Equal(t, expected, actual)
}

func TestGroupBy_Failure(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}).
		With(Aggregate(
			GroupBy(func(in int) (bool, int, error) {
				var err error
				if in == 4 {
					err = errors.New("groupby error")
				}
				return in%2 == 0, in, err
			}))).
		AsMultiMap(expectsError(t, true))

	assert.Nil(t, actual)
}

func TestGroupBy_Validate(t *testing.T) {
	for _, sc := range []struct {
		name         string
		prevStepOut  ArgTypes
		expectedOut  ArgTypes
		expectsError bool
	}{
		{
			name:        "matching_prev_step_out_type",
			prevStepOut: ArgTypes{reflect.TypeFor[int]()},
			expectedOut: ArgTypes{reflect.TypeFor[bool](), reflect.TypeFor[int]()},
		}, {
			name:         "different_prev_step_out_type",
			prevStepOut:  ArgTypes{reflect.TypeFor[string]()},
			expectsError: true,
		}, {
			name:        "skip_type_check_when_first_step",
			prevStepOut: ArgTypes{reflect.TypeFor[SkipFirstArgValidation]()},
			expectedOut: ArgTypes{reflect.TypeFor[bool](), reflect.TypeFor[int]()},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actualOut, actualErr := GroupBy(func(in int) (bool, int, error) {
				return false, 0, nil
			}).Validate(sc.prevStepOut)

			assert.Equal(t, sc.expectedOut, actualOut)
			if sc.expectsError {
				assert.ErrorIs(t, actualErr, ErrIncompatibleInArgType)
				assert.ErrorContains(t, actualErr, "[string!=int:1]")
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestFold_Success(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}).
		With(Aggregate(
			Fold(-4, func(in1, in2 int) (int, error) {
				return in1 + in2, nil
			}))).
		AsSlice(expectsError(t, false))

	assert.Equal(t, []any{6}, actual)
}

func TestFold_Failure(t *testing.T) {
	actual := Transform[int]([]int{1, 2}).
		With(Aggregate(
			Fold(-4, func(in1, in2 int) (int, error) {
				var err error
				if in2 == 2 {
					err = errors.New("fold error")
				}
				return in1 + in2, err
			}))).
		AsSlice(expectsError(t, true))

	assert.Nil(t, actual)
}

func TestFold_Validate(t *testing.T) {
	for _, sc := range []struct {
		name         string
		prevStepOut  ArgTypes
		expectedOut  ArgTypes
		expectsError bool
	}{
		{
			name:        "matching_prev_step_out_type",
			prevStepOut: ArgTypes{reflect.TypeFor[int]()},
			expectedOut: ArgTypes{reflect.TypeFor[int]()},
		}, {
			name:         "different_prev_step_out_type",
			prevStepOut:  ArgTypes{reflect.TypeFor[string]()},
			expectsError: true,
		}, {
			name:        "skip_type_check_when_first_step",
			prevStepOut: ArgTypes{reflect.TypeFor[SkipFirstArgValidation]()},
			expectedOut: ArgTypes{reflect.TypeFor[int]()},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actualOut, actualErr := Fold(-4, func(in1, in2 int) (int, error) {
				return 0, nil
			}).Validate(sc.prevStepOut)

			assert.Equal(t, sc.expectedOut, actualOut)
			if sc.expectsError {
				assert.ErrorIs(t, actualErr, ErrIncompatibleInArgType)
				assert.ErrorContains(t, actualErr, "[string!=int:1]")
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

// the following reducers are re-using Fold, so only the happy paths are tested here
func TestReduce_Success(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}).
		With(Aggregate(
			Reduce(func(in1, in2 int) (int, error) {
				return in1 + in2, nil
			}))).
		AsSlice(expectsError(t, false))

	assert.Equal(t, []any{10}, actual)
}

func TestSum_Success(t *testing.T) {
	actual := Transform[uint8]([]uint8{1, 2, 3, 4}).
		With(Aggregate(
			Sum[uint8](),
		)).
		AsSlice(expectsError(t, false))

	assert.Equal(t, []any{uint8(10)}, actual)
}

const float64EqualityThreshold = 1e-9

func TestMax_Success(t *testing.T) {
	testMax(t, []int{-1, -2, -3}, int(-1))
	testMax(t, []int8{-1, -2, -3}, int8(-1))
	testMax(t, []int16{-1, -2, -3}, int16(-1))
	testMax(t, []int32{-1, -2, -3}, int32(-1))
	testMax(t, []int64{-1, -2, -3}, int64(-1))

	testMax(t, []uint{1, 2, 3}, uint(3))
	testMax(t, []uint8{1, 2, 3}, uint8(3))
	testMax(t, []uint16{1, 2, 3}, uint16(3))
	testMax(t, []uint32{1, 2, 3}, uint32(3))
	testMax(t, []uint64{1, 2, 3}, uint64(3))

	testMax(t, []float32{-1, -2, -3}, float32(-1))
	testMax(t, []float64{-1, -2, -3}, float64(-1))
}

func testMax[T number](t *testing.T, input []T, expected T) {
	t.Helper()
	actual := Transform[T](input).
		With(Aggregate(
			Max[T](),
		)).
		AsSlice(expectsError(t, false))

	switch reflect.TypeFor[T]().Kind() {
	case reflect.Float32:
		a := actual[0].(float32)
		assert.Less(t, math.Abs(float64(expected)-float64(a)), float64EqualityThreshold)
	case reflect.Float64:
		a := actual[0].(float64)
		assert.Less(t, math.Abs(float64(expected)-float64(a)), float64EqualityThreshold)
	default:
		assert.Equal(t, []any{expected}, actual)
	}
}

func TestMin_Success(t *testing.T) {
	testMin(t, []int{-1, -2, -3}, int(-3))
	testMin(t, []int8{-1, -2, -3}, int8(-3))
	testMin(t, []int16{-1, -2, -3}, int16(-3))
	testMin(t, []int32{-1, -2, -3}, int32(-3))
	testMin(t, []int64{-1, -2, -3}, int64(-3))

	testMin(t, []uint{1, 2, 3}, uint(1))
	testMin(t, []uint8{1, 2, 3}, uint8(1))
	testMin(t, []uint16{1, 2, 3}, uint16(1))
	testMin(t, []uint32{1, 2, 3}, uint32(1))
	testMin(t, []uint64{1, 2, 3}, uint64(1))

	testMin(t, []float32{-1, -2, -3}, float32(-3))
	testMin(t, []float64{-1, -2, -3}, float64(-3))
}

func testMin[T number](t *testing.T, input []T, expected T) {
	t.Helper()
	actual := Transform[T](input).
		With(Aggregate(
			Min[T](),
		)).
		AsSlice(expectsError(t, false))

	switch reflect.TypeFor[T]().Kind() {
	case reflect.Float32:
		a := actual[0].(float32)
		assert.Less(t, math.Abs(float64(expected)-float64(a)), float64EqualityThreshold)
	case reflect.Float64:
		a := actual[0].(float64)
		assert.Less(t, math.Abs(float64(expected)-float64(a)), float64EqualityThreshold)
	default:
		assert.Equal(t, []any{expected}, actual)
	}
}

func TestAvg_Success(t *testing.T) {
	actual := Transform[float64]([]float64{-1.25, 1.33, -0.77}).
		With(Aggregate(
			Avg(),
		)).
		AsSlice(expectsError(t, false))

	assert.Less(t, math.Abs(float64(-0.23)-actual[0].(float64)), float64EqualityThreshold)
}
