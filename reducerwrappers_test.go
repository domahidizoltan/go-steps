package steps

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupBy_Success(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		With(Aggregate(
			GroupBy(func(in int) (bool, int, error) {
				return in%2 == 0, in, nil
			}))).
		AsMultiMap()

	expected := map[any][]any{
		true:  {2, 4},
		false: {1, 3},
	}
	assert.Equal(t, expected, actual)
}

func TestGroupBy_Failure(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, true))).
		With(Aggregate(
			GroupBy(func(in int) (bool, int, error) {
				var err error
				if in == 4 {
					err = errors.New("groupby error")
				}
				return in%2 == 0, in, err
			}))).
		AsMultiMap()

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
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		With(Aggregate(
			Fold(-4, func(in1, in2 int) (int, error) {
				return in1 + in2, nil
			}))).
		AsSlice()

	assert.Equal(t, []any{6}, actual)
}

func TestFold_Failure(t *testing.T) {
	actual := Transform[int]([]int{1, 2}, WithErrorHandler(expectsError(t, true))).
		With(Aggregate(
			Fold(-4, func(in1, in2 int) (int, error) {
				var err error
				if in2 == 2 {
					err = errors.New("fold error")
				}
				return in1 + in2, err
			}))).
		AsSlice()

	assert.Empty(t, actual)
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
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		With(Aggregate(
			Reduce(func(in1, in2 int) (int, error) {
				return in1 + in2, nil
			}))).
		AsSlice()

	assert.Equal(t, []any{10}, actual)
}

func TestSum_Success(t *testing.T) {
	actual := Transform[uint8]([]uint8{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		With(Aggregate(
			Sum[uint8](),
		)).
		AsSlice()

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
	actual := Transform[T](input, WithErrorHandler(expectsError(t, false))).
		With(Aggregate(
			Max[T](),
		)).
		AsSlice()

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
	actual := Transform[T](input, WithErrorHandler(expectsError(t, false))).
		With(Aggregate(
			Min[T](),
		)).
		AsSlice()

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
	actual := Transform[float64]([]float64{-1.25, 1.33, -0.77}, WithErrorHandler(expectsError(t, false))).
		With(Aggregate(
			Avg(),
		)).
		AsSlice()

	assert.Less(t, math.Abs(float64(-0.23)-actual[0].(float64)), float64EqualityThreshold)
}

func ExampleGroupBy() {
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		With(Aggregate(
			GroupBy(func(in int) (uint8, int, error) {
				return uint8(in % 2), in, nil
			}),
		)).
		AsMultiMap()

	fmt.Println(res)
	// Output: map[0:[2 4] 1:[1 3 5]]
}

func ExampleFold() {
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		With(Aggregate(
			Fold(10, func(in1, in2 int) (int, error) {
				return in1 + in2, nil
			}),
		)).
		AsSlice()

	fmt.Println(res)
	// Output: [25]
}

func ExampleReduce() {
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		With(Aggregate(
			Reduce(func(in1, in2 int) (int, error) {
				return in1 + in2, nil
			}),
		)).
		AsSlice()

	fmt.Println(res)
	// Output: [15]
}

func ExampleSum() {
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		With(Aggregate(
			Sum[int](),
		)).
		AsSlice()

	fmt.Println(res)
	// Output: [15]
}

func ExampleMax() {
	res := Transform[float64]([]float64{-1.1, -3.3, -2.2, -5.5, -4.4}).
		With(Aggregate(
			Max[float64](),
		)).
		AsSlice()

	fmt.Println(res)
	// Output: [-1.1]
}

func ExampleMin() {
	res := Transform[float64]([]float64{-1.1, -3.3, -2.2, -5.5, -4.4}).
		With(Aggregate(
			Min[float64](),
		)).
		AsSlice()

	fmt.Println(res)
	// Output: [-5.5]
}

func ExampleAvg() {
	res := Transform[float64]([]float64{-1.1, -2.1, -3.1, -4.1, -5.1}).
		With(Aggregate(
			Avg(),
		)).
		AsSlice()

	fmt.Println(res)
	// Output: [-3.1]
}
