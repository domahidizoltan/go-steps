package step

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validateScenario struct {
	name        string
	error       error
	errorDetail string
	wrappers    []StepWrapper
	expected    []StepFn
}

type testStruct struct{}

var (
	filterFn = Filter(func(in int) (bool, error) {
		return true, nil
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
	actual, err := GetValidatedSteps[any](nil)
	require.NoError(t, err)
	assert.Nil(t, actual)
}

func TestGetValidatedSteps_HasNoError(t *testing.T) {
	for _, sc := range []validateScenario{
		{
			name:     "single_wrapper",
			wrappers: []StepWrapper{mapFn},
			expected: []StepFn{mapFn.StepFn},
		},
		{
			name:     "multiple_wrappers",
			wrappers: []StepWrapper{mapFn, filterFn},
			expected: []StepFn{mapFn.StepFn, filterFn.StepFn},
		},
		{
			name:     "branch_wrappers",
			wrappers: []StepWrapper{mapFn, splitFn, branchFn, mergeFn},
			expected: []StepFn{mapFn.StepFn, splitFn.StepFn, branchFn.StepFn, mergeFn.StepFn},
		},
		{
			name:     "collections_and_structs",
			wrappers: []StepWrapper{sliceToMapFn, mapToStructFn},
			expected: []StepFn{sliceToMapFn.StepFn, mapToStructFn.StepFn},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actual, err := GetValidatedSteps[int](sc.wrappers)
			require.NoError(t, err)
			assert.Len(t, sc.expected, len(actual))
			for i, act := range actual {
				assert.Equal(t, reflect.TypeOf(sc.expected[i]), reflect.TypeOf(act))
			}
		})
	}
}

func TestGetValidatedSteps_ReturnsError(t *testing.T) {
	for _, sc := range []validateScenario{
		{
			name:        "transform_input_type_is_different",
			wrappers:    []StepWrapper{mapStrFn},
			error:       ErrStepValidationFailed,
			errorDetail: "[Map:1]: incompatible input argument type [int!=string:1]",
		},
		{
			name:        "step_input_type_is_different",
			wrappers:    []StepWrapper{mapFn, mapStrFn},
			error:       ErrStepValidationFailed,
			errorDetail: "[Map:2]: incompatible input argument type [int!=string:1]",
		},
		{
			name: "pointers_used_when_not_needed",
			wrappers: []StepWrapper{mapFn, Map(func(in *int) (*int, error) {
				return nil, nil
			})},
			error:       ErrStepValidationFailed,
			errorDetail: "[Map:2]: incompatible input argument type [int!=*int:1]",
		},
		{
			name: "validation_called_inside_branch",
			wrappers: []StepWrapper{mapFn, splitFn, WithBranches[int](
				Steps(mapFn), Steps(Map(func(in *int) (*int, error) { return nil, nil })))},
			error:       ErrStepValidationFailed,
			errorDetail: "[WithBranches:3]: step validation failed [Map:1]: incompatible input argument type [int!=*int:1]",
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actual, err := GetValidatedSteps[int](sc.wrappers)
			require.Empty(t, actual)
			assert.ErrorIs(t, err, sc.error, fmt.Sprintf("unexpected error: %s", err.Error()))
			assert.Contains(t, err.Error(), sc.errorDetail, fmt.Sprintf("unexpected error detail: %s", err.Error()))
		})
	}
}

// run aggregator
// skip
// yield simple
// yield indexed
