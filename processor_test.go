package steps

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct{}

var (
	filterFn = Filter(func(in int) (bool, error) {
		return in%2 == 0, nil
	})
	mapFn = Map(func(in int) (int, error) {
		return in + 1, nil
	})
	mapStrFn = Map(func(in string) (string, error) {
		return in + "1", nil
	})
	splitFn = Split(func(in int) (uint8, error) {
		return uint8(in % 2), nil
	})
	branchFn = WithBranches[int](
		Steps(mapFn),
		Steps(mapFn),
	)
	mergeFn = Merge()

	sliceToMapFn = Map(func(in int) (map[int]string, error) {
		return map[int]string{}, nil
	})
	mapToStructFn = Map(func(in map[int]string) ([]testStruct, error) {
		return []testStruct{}, nil
	})
)

func TestGetValidatedSteps_HasNoError_WhenNoStepWrappersAreGiven(t *testing.T) {
	actual, _, err := getValidatedSteps[any](nil)
	require.NoError(t, err)
	assert.Nil(t, actual)
}

func TestGetValidatedSteps_HasNoError(t *testing.T) {
	type scenario struct {
		name     string
		wrappers []StepWrapper
		expected []StepFn
	}

	for _, sc := range []scenario{
		{
			name:     "empty_wrappers",
			wrappers: []StepWrapper{},
			expected: []StepFn{},
		}, {
			name:     "single_wrapper",
			wrappers: []StepWrapper{mapFn},
			expected: []StepFn{mapFn.StepFn},
		}, {
			name:     "multiple_wrappers",
			wrappers: []StepWrapper{mapFn, filterFn},
			expected: []StepFn{mapFn.StepFn, filterFn.StepFn},
		}, {
			name:     "branch_wrappers",
			wrappers: []StepWrapper{mapFn, splitFn, branchFn, mergeFn},
			expected: []StepFn{mapFn.StepFn, splitFn.StepFn, branchFn.StepFn, mergeFn.StepFn},
		}, {
			name:     "collections_and_structs",
			wrappers: []StepWrapper{sliceToMapFn, mapToStructFn},
			expected: []StepFn{sliceToMapFn.StepFn, mapToStructFn.StepFn},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actual, _, err := getValidatedSteps[int](sc.wrappers)
			require.NoError(t, err)
			assert.Len(t, sc.expected, len(actual))
			for i, act := range actual {
				assert.Equal(t, reflect.TypeOf(sc.expected[i]), reflect.TypeOf(act))
			}
		})
	}
}

func TestGetValidatedSteps_ReturnsError(t *testing.T) {
	type scenario struct {
		name        string
		error       error
		errorDetail string
		wrappers    []StepWrapper
	}

	for _, sc := range []scenario{
		{
			name:        "transform_input_type_is_different",
			wrappers:    []StepWrapper{mapStrFn},
			error:       ErrStepValidationFailed,
			errorDetail: "[Map:1]: incompatible input argument type [int!=string:1]",
		}, {
			name:        "step_input_type_is_different",
			wrappers:    []StepWrapper{mapFn, mapStrFn},
			error:       ErrStepValidationFailed,
			errorDetail: "[Map:2]: incompatible input argument type [int!=string:1]",
		}, {
			name: "pointers_used_when_not_needed",
			wrappers: []StepWrapper{mapFn, Map(func(in *int) (*int, error) {
				return nil, nil
			})},
			error:       ErrStepValidationFailed,
			errorDetail: "[Map:2]: incompatible input argument type [int!=*int:1]",
		}, {
			name: "validation_called_inside_branch",
			wrappers: []StepWrapper{mapFn, splitFn, WithBranches[int](
				Steps(mapFn), Steps(Map(func(in *int) (*int, error) { return nil, nil })))},
			error:       ErrStepValidationFailed,
			errorDetail: "[WithBranches:3]: step validation failed [Map:1]: incompatible input argument type [int!=*int:1]",
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actual, _, err := getValidatedSteps[int](sc.wrappers)
			require.Empty(t, actual)
			assert.ErrorIsf(t, err, sc.error, "unexpected error: %s", err.Error())
			assert.Containsf(t, err.Error(), sc.errorDetail, "unexpected error detail: %s", err.Error())
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestProcess(t *testing.T) {
	type scenario[T any] struct {
		name    string
		input   int
		output  T
		skipped bool
		err     error
	}

	for _, sc := range []scenario[*int]{
		{
			name:   "value_processed",
			input:  2,
			output: ptr(3),
		}, {
			name:    "value_skipped",
			input:   1,
			skipped: true,
		}, {
			name:  "step_returns_error",
			input: 2,
			err:   errors.New("map error"),
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			var processedValue *int
			yield := func(in any) bool {
				processedValue = ptr(in.(int))
				return true
			}
			mapFn := Map(func(in int) (int, error) {
				return in + 1, sc.err
			})
			trn := &transformer{
				steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
			}

			skipped, _, err := process(sc.input, yield, trn, false)

			assert.Equal(t, sc.skipped, skipped)
			assert.Equal(t, sc.err, err)
			assert.Equal(t, sc.output, processedValue)
		})
	}
}

func TestProcessAndAggregate(t *testing.T) {
	type scenario[T any] struct {
		name       string
		input      int
		output     T
		terminated bool
		err        error
	}

	for _, sc := range []scenario[map[bool][]int]{
		{
			name:       "value_processed",
			input:      2,
			output:     map[bool][]int{true: {2}},
			terminated: true,
		}, {
			name:  "reducer_returns_error",
			input: 2,
			err:   errors.New("reducer error"),
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			var processedValue map[bool][]int
			yield := func(in any) bool {
				processedValue = in.(map[bool][]int)
				return true
			}
			groupEven := GroupBy(func(in int) (bool, int, error) {
				return true, 2, sc.err
			})

			trn := &transformer{
				steps:      []StepFn{filterFn.StepFn, mapFn.StepFn},
				aggregator: groupEven.ReducerFn,
			}

			_, terminated, err := process(sc.input, yield, trn, true)

			assert.Equal(t, sc.terminated, terminated)
			assert.Equal(t, sc.err, err)
			assert.Equal(t, sc.output, processedValue)
		})
	}
}

func TestProcessIndexed(t *testing.T) {
	type scenario[T any] struct {
		name      string
		input     int
		output    T
		outputKey *string
		skipped   bool
		err       error
	}

	for _, sc := range []scenario[*int]{
		{
			name:      "value_processed",
			input:     2,
			outputKey: ptr("value_processed"),
			output:    ptr(3),
		}, {
			name:    "value_skipped",
			input:   1,
			skipped: true,
		}, {
			name:  "step_returns_error",
			input: 2,
			err:   errors.New("map error"),
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			var processedKey *string
			var processedValue *int
			yield := func(idx, in any) bool {
				processedKey = ptr(idx.(string))
				processedValue = ptr(in.(int))
				return true
			}
			mapFn := Map(func(in int) (int, error) {
				return in + 1, sc.err
			})
			trn := &transformer{
				steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
			}

			skipped, _, err := processIndexed(sc.name, sc.input, yield, trn, false)

			assert.Equal(t, sc.skipped, skipped)
			assert.Equal(t, sc.err, err)
			assert.Equal(t, sc.output, processedValue)
			assert.Equal(t, sc.outputKey, processedKey)
		})
	}
}

func TestProcessIndexedAndAggregate(t *testing.T) {
	type scenario[T any] struct {
		name       string
		input      int
		output     T
		outputKey  *string
		terminated bool
		err        error
	}

	for _, sc := range []scenario[map[bool][]int]{
		{
			name:       "value_processed",
			input:      2,
			output:     map[bool][]int{true: {2}},
			outputKey:  ptr("value_processed"),
			terminated: true,
		}, {
			name:  "reducer_returns_error",
			input: 2,
			err:   errors.New("reducer error"),
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			var processedKey *string
			var processedValue map[bool][]int
			yield := func(idx, in any) bool {
				processedKey = ptr(idx.(string))
				processedValue = in.(map[bool][]int)
				return true
			}
			groupEven := GroupBy(func(in int) (bool, int, error) {
				return true, 2, sc.err
			})

			trn := &transformer{
				steps:      []StepFn{filterFn.StepFn, mapFn.StepFn},
				aggregator: groupEven.ReducerFn,
			}

			_, terminated, err := processIndexed(sc.name, sc.input, yield, trn, true)

			assert.Equal(t, sc.terminated, terminated)
			assert.Equal(t, sc.err, err)
			assert.Equal(t, sc.output, processedValue)
			assert.Equal(t, sc.outputKey, processedKey)
		})
	}
}

func TestProcessYieldsRawInput_WhenStepsMissing(t *testing.T) {
	var processedValue *int
	yield := func(in any) bool {
		processedValue = ptr(in.(int))
		return true
	}

	skipped, _, err := process(42, yield, nil, false)

	assert.False(t, skipped)
	assert.NoError(t, err)
	assert.Equal(t, 42, *processedValue)
}
