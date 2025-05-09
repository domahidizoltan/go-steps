package steps

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap_Success(t *testing.T) {
	actual := Transform[string]([]string{"1", "2"}, WithErrorHandler(expectsError(t, false))).
		WithSteps(
			Map(func(in string) (int, error) {
				return strconv.Atoi(in)
			})).
		AsSlice()

	assert.Equal(t, []any{1, 2}, actual)
}

func TestMap_Failure(t *testing.T) {
	actual := Transform[string]([]string{"1", "x", "2"}, WithErrorHandler(expectsError(t, true))).
		WithSteps(
			Map(func(in string) (int, error) {
				return strconv.Atoi(in)
			})).
		AsSlice()

	assert.Equal(t, []any{1}, actual)
}

func TestMap_Validate(t *testing.T) {
	for _, sc := range []struct {
		name         string
		prevStepOut  ArgTypes
		expectedOut  ArgTypes
		expectsError bool
	}{
		{
			name:        "matching_prev_step_out_type",
			prevStepOut: ArgTypes{reflect.TypeFor[string]()},
			expectedOut: ArgTypes{reflect.TypeFor[int]()},
		}, {
			name:         "different_prev_step_out_type",
			prevStepOut:  ArgTypes{reflect.TypeFor[int]()},
			expectsError: true,
		}, {
			name:        "skip_type_check_when_first_step",
			prevStepOut: ArgTypes{reflect.TypeFor[SkipFirstArgValidation]()},
			expectedOut: ArgTypes{reflect.TypeFor[int]()},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actualOut, actualErr := Map(func(in string) (int, error) {
				return 0, nil
			}).Validate(sc.prevStepOut)

			assert.Equal(t, sc.expectedOut, actualOut)
			if sc.expectsError {
				assert.ErrorIs(t, actualErr, ErrIncompatibleInArgType)
				assert.ErrorContains(t, actualErr, "[int!=string:1]")
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func testSimpleFilterValidate(t *testing.T, stepWrapper StepWrapper) {
	t.Helper()
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
			actualOut, actualErr := stepWrapper.Validate(sc.prevStepOut)

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

func TestFilter_Success(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		WithSteps(
			Filter(func(in int) (bool, error) {
				return in%2 == 0, nil
			})).
		AsSlice()

	assert.Equal(t, []any{2, 4}, actual)
}

func TestFilter_Failure(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3}, WithErrorHandler(expectsError(t, true))).
		WithSteps(
			Filter(func(in int) (bool, error) {
				if in == 2 {
					return false, errors.New("filter error")
				}
				return true, nil
			})).
		AsSlice()

	assert.Equal(t, []any{1}, actual)
}

func TestFilter_Validate(t *testing.T) {
	testSimpleFilterValidate(t, Filter(func(in int) (bool, error) {
		return true, nil
	}))
}

func TestTake_Success(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		WithSteps(Take[int](2)).
		AsSlice()

	assert.Equal(t, []any{1, 2}, actual)
}

func TestTake_Validate(t *testing.T) {
	testSimpleFilterValidate(t, Take[int](0))
}

func TestTakeWhile_Success(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		WithSteps(
			TakeWhile(func(in int) (bool, error) {
				return in <= 2, nil
			})).
		AsSlice()

	assert.Equal(t, []any{1, 2}, actual)
}

func TestTakeWhile_Failure(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, true))).
		WithSteps(
			TakeWhile(func(in int) (bool, error) {
				if in == 2 {
					return false, errors.New("TakeWhile error")
				}
				return true, nil
			})).
		AsSlice()

	assert.Equal(t, []any{1}, actual)
}

func TestTakeWhile_Validate(t *testing.T) {
	testSimpleFilterValidate(t, Take[int](0))
}

func TestSkip_Success(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		WithSteps(Skip[int](2)).
		AsSlice()

	assert.Equal(t, []any{3, 4}, actual)
}

func TestSkip_Validate(t *testing.T) {
	testSimpleFilterValidate(t, Skip[int](0))
}

func TestSkipWhile_Success(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		WithSteps(
			SkipWhile(func(in int) (bool, error) {
				return in <= 2, nil
			})).
		AsSlice()

	assert.Equal(t, []any{3, 4}, actual)
}

func TestSkipWhile_Failure(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, true))).
		WithSteps(
			SkipWhile(func(in int) (bool, error) {
				if in == 2 {
					return false, errors.New("SkipWhile error")
				}
				return true, nil
			})).
		AsSlice()

	assert.Empty(t, actual)
}

func TestSkipWhile_Validate(t *testing.T) {
	testSimpleFilterValidate(t, Skip[int](0))
}

type testLogWriter struct {
	output      []byte
	returnError error
}

func (t *testLogWriter) ReturnError(err error) {
	t.returnError = err
}

func (t *testLogWriter) Write(b []byte) (int, error) {
	if t.returnError != nil {
		return 0, t.returnError
	}

	t.output = append(t.output, b...)
	return len(b), nil
}

func TestDo_Success(t *testing.T) {
	doCounter := 0
	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		WithSteps(
			Do(func(in int) error {
				doCounter++
				return nil
			}),
			Filter(func(in int) (bool, error) {
				return in%2 == 0, nil
			}),
		).
		AsSlice()

	assert.Equal(t, doCounter, 4)
	assert.Equal(t, []any{2, 4}, actual)
}

func TestDo_Failure(t *testing.T) {
	actual := Transform[int]([]int{1, 2, 3}, WithErrorHandler(expectsError(t, true))).
		WithSteps(
			Do(func(in int) error {
				if in == 3 {
					return errors.New("do error")
				}
				return nil
			}),
			Filter(func(in int) (bool, error) {
				return in%2 == 0, nil
			})).
		AsSlice()

	assert.Equal(t, []any{2}, actual)
}

func TestDo_Validate(t *testing.T) {
	testSimpleFilterValidate(t, Do(func(in int) error {
		return nil
	}))
}

func TestLog_Success(t *testing.T) {
	testLogWriter := testLogWriter{}
	actual := Transform[int]([]int{1, 2, 3, 4}, WithName("testLog"), WithLogWriter(&testLogWriter), WithErrorHandler(expectsError(t, false))).
		WithSteps(
			Log("before"),
			Filter(func(in int) (bool, error) {
				return in%2 == 0, nil
			}),
			Log("after")).
		AsSlice()

	expectedLogOutput := `before transformer:testLog 	arg0: 1 
before transformer:testLog 	arg0: 2 
after transformer:testLog 	arg0: 2 
before transformer:testLog 	arg0: 3 
before transformer:testLog 	arg0: 4 
after transformer:testLog 	arg0: 4 
`
	assert.Equal(t, expectedLogOutput, string(testLogWriter.output))
	assert.Equal(t, []any{2, 4}, actual)
}

func TestLog_Failure(t *testing.T) {
	testLogWriter := testLogWriter{}
	testLogWriter.returnError = errors.New("log error")
	actual := Transform[int]([]int{1, 2, 3}, WithLogWriter(&testLogWriter), WithErrorHandler(expectsError(t, true))).
		WithSteps(
			Log(),
			Filter(func(in int) (bool, error) {
				return in%2 == 0, nil
			})).
		AsSlice()

	assert.Empty(t, actual)
}

func TestLog_Validate_ReturnsSameArgTypes(t *testing.T) {
	prevStepOut := ArgTypes{reflect.TypeFor[SkipFirstArgValidation](), reflect.TypeFor[float32]()}
	actualOut, actualErr := Log().Validate(prevStepOut)

	assert.Equal(t, prevStepOut, actualOut)
	assert.NoError(t, actualErr)
}

type evenOdd uint8

const (
	even evenOdd = iota
	odd
)

func TestSplit_Success(t *testing.T) {
	actual := map[uint8][]any{}
	iter := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, false))).
		WithSteps(
			Split(func(in int) (evenOdd, error) {
				if in%2 == 0 {
					return even, nil
				}
				return odd, nil
			})).
		AsRange()

	for item := range iter {
		i := item.(branch)
		actual[i.key] = append(actual[i.key], i.value)
	}
	expected := map[uint8][]any{
		uint8(even): {2, 4},
		uint8(odd):  {1, 3},
	}
	assert.EqualValues(t, expected, actual)
}

func TestSplit_Failure(t *testing.T) {
	actual := map[uint8][]any{}
	iter := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, true))).
		WithSteps(
			Split(func(in int) (evenOdd, error) {
				if in == 4 {
					return even, errors.New("split error")
				}
				if in%2 == 0 {
					return even, nil
				}
				return odd, nil
			})).
		AsRange()

	for item := range iter {
		i := item.(branch)
		actual[i.key] = append(actual[i.key], i.value)
	}
	expected := map[uint8][]any{
		uint8(even): {2},
		uint8(odd):  {1, 3},
	}
	assert.EqualValues(t, expected, actual)
}

func TestSplit_Validate(t *testing.T) {
	for _, sc := range []struct {
		name         string
		prevStepOut  ArgTypes
		expectedOut  ArgTypes
		expectsError bool
	}{
		{
			name:        "matching_prev_step_out_type_returns_branch_type",
			prevStepOut: ArgTypes{reflect.TypeFor[int]()},
			expectedOut: ArgTypes{reflect.TypeFor[branch]()},
		}, {
			name:         "different_prev_step_out_type",
			prevStepOut:  ArgTypes{reflect.TypeFor[string]()},
			expectsError: true,
		}, {
			name:        "skip_type_check_when_first_step",
			prevStepOut: ArgTypes{reflect.TypeFor[SkipFirstArgValidation]()},
			expectedOut: ArgTypes{reflect.TypeFor[branch]()},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actualOut, actualErr := Split(func(in int) (uint8, error) {
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

func TestMerge_Success(t *testing.T) {
	actual := Transform[branch]([]branch{
		{key: uint8(odd), T: reflect.ValueOf(1), value: 1},
		{key: uint8(even), T: reflect.ValueOf(2), value: 2},
		{key: uint8(odd), T: reflect.ValueOf(3), value: 3},
		{key: uint8(even), T: reflect.ValueOf(4), value: 4},
	}, WithErrorHandler(expectsError(t, false))).
		WithSteps(Merge()).
		AsSlice()

	assert.EqualValues(t, []any{1, 2, 3, 4}, actual)
}

func TestMerge_Validate(t *testing.T) {
	for _, sc := range []struct {
		name         string
		prevStepOut  ArgTypes
		expectedOut  ArgTypes
		expectsError bool
	}{
		{
			name:        "matching_prev_step_out_type_returns_any_type",
			prevStepOut: ArgTypes{reflect.TypeFor[branch]()},
			expectedOut: ArgTypes{reflect.TypeFor[any]()},
		}, {
			name:         "only_branch_type_as_prev_step_out",
			prevStepOut:  ArgTypes{reflect.TypeFor[string]()},
			expectsError: true,
		}, {
			name:        "skip_type_check_when_first_step",
			prevStepOut: ArgTypes{reflect.TypeFor[SkipFirstArgValidation]()},
			expectedOut: ArgTypes{reflect.TypeFor[any]()},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actualOut, actualErr := Merge().Validate(sc.prevStepOut)

			assert.Equal(t, sc.expectedOut, actualOut)
			if sc.expectsError {
				assert.ErrorIs(t, actualErr, ErrIncompatibleInArgType)
				assert.ErrorContains(t, actualErr, "[string!=steps.branch:1]")
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func TestWithBranches_Success(t *testing.T) {
	split := Split(func(in int) (uint8, error) {
		return uint8(in % 2), nil
	})
	addOne := Map(func(in int) (int, error) {
		return in + 1, nil
	})
	multiplyByTwo := Map(func(in int) (int, error) {
		return in * 2, nil
	})

	testLogWriter := testLogWriter{}
	actual := Transform[int]([]int{1, 2, 3, 4}, WithLogWriter(&testLogWriter), WithErrorHandler(expectsError(t, false))).
		WithSteps(split,
			WithBranches[int](
				Steps(addOne),
				Steps(Log("test"), multiplyByTwo),
			),
			Merge()).
		AsSlice()

	assert.EqualValues(t, []any{2, 3, 6, 5}, actual)
	assert.True(t, bytes.HasPrefix(testLogWriter.output, []byte("test")))
}

func TestWithBranches_Failure(t *testing.T) {
	split := Split(func(in int) (uint8, error) {
		return uint8(in % 2), nil
	})
	addOne := Map(func(in int) (int, error) {
		if in == 4 {
			return 0, errors.New("add error")
		}
		return in + 1, nil
	})
	multiplyByTwo := Map(func(in int) (int, error) {
		return in * 2, nil
	})

	actual := Transform[int]([]int{1, 2, 3, 4}, WithErrorHandler(expectsError(t, true))).
		WithSteps(split,
			WithBranches[int](
				Steps(addOne),
				Steps(multiplyByTwo),
			),
			Merge()).
		AsSlice()

	assert.EqualValues(t, []any{2, 3, 6}, actual)
}

func TestWithBranches_Validate(t *testing.T) {
	for _, sc := range []struct {
		name                  string
		prevStepOut           ArgTypes
		expectedOut           ArgTypes
		expectsErrorContain   string
		overrideAddOne        *StepWrapper
		overrideMultiplyByTwo *StepWrapper
	}{
		{
			name:        "matching_prev_step_out_type",
			prevStepOut: ArgTypes{reflect.TypeFor[branch]()},
			expectedOut: ArgTypes{reflect.TypeFor[branch]()},
		}, {
			name:                "only_branch_type_as_prev_step_out",
			prevStepOut:         ArgTypes{reflect.TypeFor[string]()},
			expectsErrorContain: "[string!=steps.branch:",
		}, {
			name:        "skip_type_check_when_first_step",
			prevStepOut: ArgTypes{reflect.TypeFor[SkipFirstArgValidation]()},
			expectedOut: ArgTypes{reflect.TypeFor[branch]()},
		}, {
			name:        "different_branch_in_type",
			prevStepOut: ArgTypes{reflect.TypeFor[branch]()},
			overrideMultiplyByTwo: func() *StepWrapper {
				m := Map(func(in string) (string, error) {
					return "", nil
				})
				return &m
			}(),
			expectsErrorContain: "step validation failed [Map:3]: incompatible input argument type [int!=string:1]",
		}, {
			name:        "validates_first_branch",
			prevStepOut: ArgTypes{reflect.TypeFor[branch]()},
			overrideAddOne: func() *StepWrapper {
				m := Map(func(in string) (int, error) {
					return 0, errors.New("add error")
				})
				return &m
			}(),
			expectsErrorContain: "step validation failed [Map:2]: incompatible input argument type [int!=string:1]",
		}, {
			name:        "validates_second_branch",
			prevStepOut: ArgTypes{reflect.TypeFor[branch]()},
			overrideMultiplyByTwo: func() *StepWrapper {
				m := Map(func(in string) (int, error) {
					return 0, errors.New("multiply error")
				})
				return &m
			}(),
			expectsErrorContain: "step validation failed [Map:3]: incompatible input argument type [int!=string:1]",
		},
	} {
		addZero := Map(func(in int) (int, error) {
			return 0, nil
		})
		addOne := Map(func(in int) (int, error) {
			return in + 1, nil
		})
		multiplyByTwo := Map(func(in int) (int, error) {
			return in * 2, nil
		})

		t.Run(sc.name, func(t *testing.T) {
			if sc.overrideAddOne != nil {
				addOne = *sc.overrideAddOne
			}
			if sc.overrideMultiplyByTwo != nil {
				multiplyByTwo = *sc.overrideMultiplyByTwo
			}

			actualOut, actualErr := WithBranches[int](
				Steps(addZero, addOne),
				Steps(addZero, addZero, multiplyByTwo),
			).Validate(sc.prevStepOut)

			assert.Equal(t, sc.expectedOut, actualOut)
			if len(sc.expectsErrorContain) > 0 {
				assert.ErrorIs(t, actualErr, ErrIncompatibleInArgType)
				assert.ErrorContains(t, actualErr, sc.expectsErrorContain)
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

func expectsError(t *testing.T, expectsError bool) func(error) {
	t.Helper()
	return func(err error) {
		if expectsError {
			assert.Error(t, err)
			return
		}
		assert.NoError(t, err)
	}
}

func ExampleMap() {
	res := Transform[int]([]int{104, 101, 108, 108, 111}).
		WithSteps(
			Map(func(in int) (string, error) {
				return string(rune(in)), nil
			}),
		).AsSlice()

	fmt.Println(res)
	// Output: [h e l l o]
}

func ExampleFilter() {
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		WithSteps(
			Filter(func(in int) (bool, error) {
				return in%2 == 1, nil
			}),
		).AsSlice()

	fmt.Println(res)
	// Output: [1 3 5]
}

func ExampleTake() {
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		WithSteps(
			Take[int](3),
		).AsSlice()

	fmt.Println(res)
	// Output: [1 2 3]
}

func ExampleTakeWhile() {
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		WithSteps(
			TakeWhile(func(in int) (bool, error) {
				return in <= 3, nil
			}),
		).AsSlice()

	fmt.Println(res)
	// Output: [1 2 3]
}

func ExampleSkip() {
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		WithSteps(
			Skip[int](3),
		).AsSlice()

	fmt.Println(res)
	// Output: [4 5]
}

func ExampleSkipWhile() {
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		WithSteps(
			SkipWhile(func(in int) (bool, error) {
				return in <= 3, nil
			}),
		).AsSlice()

	fmt.Println(res)
	// Output: [4 5]
}

func ExampleDo() {
	total := 0
	res := Transform[int]([]int{1, 2, 3, 4, 5}).
		WithSteps(
			Do(func(in int) error {
				total += in
				return nil
			}),
		).AsSlice()

	fmt.Printf("res: %v\ntotal: %d", res, total)
	// Output:
	// res: [1 2 3 4 5]
	// total: 15
}

func ExampleLog() {
	buf := bytes.NewBufferString("")

	res := Transform[int]([]int{1, 2, 3}, WithLogWriter(buf)).
		WithSteps(
			Log("before filter"),
			Filter(func(i int) (bool, error) {
				return i%2 == 1, nil
			}),
			Log("after filter"),
		).AsSlice()

	logs := strings.ReplaceAll(buf.String(), " \n", "\n")
	fmt.Printf("res: %v\nlogs: %s", res, logs)
	// Output:
	// res: [1 3]
	// logs: before filter 	arg0: 1
	// after filter 	arg0: 1
	// before filter 	arg0: 2
	// before filter 	arg0: 3
	// after filter 	arg0: 3
}

func ExampleSplit() {
	fmt.Println("see WithBranches")
	// Output: see WithBranches
}

func ExampleWithBranches() {
	res := Transform[int]([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
		WithSteps(
			Split(func(in int) (uint8, error) {
				return uint8(in % 2), nil
			}),
			WithBranches[int](
				Steps(
					Map(func(in int) (string, error) {
						return strconv.Itoa(-in), nil
					}),
				),
				Steps(
					Map(func(in int) (string, error) {
						return strconv.Itoa(in * 10), nil
					}),
				),
			),
			Merge(),
		).
		AsSlice()

	fmt.Printf("%#v", res)
	// Output: []interface {}{"10", "-2", "30", "-4", "50", "-6", "70", "-8", "90", "-10"}
}

func ExampleMerge() {
	fmt.Println("see WithBranches")
	// Output: see WithBranches
}
