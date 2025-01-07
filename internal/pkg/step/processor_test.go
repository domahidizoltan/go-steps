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

var (
	filterFn = Filter(func(in int) (bool, error) {
		return true, nil
	})
	mapFn = Map(func(in int) (int, error) {
		return in + 1, nil
	})
	splitFn = Split(func(in int) (uint8, error) {
		return uint8(in % 2), nil
	})
	branchFn = WithBranches[int](
		Steps(mapFn),
		Steps(mapFn),
	)
	mergeFn = Merge()
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
			name: "input_type_is_different",
			wrappers: []StepWrapper{Map(func(in string) (string, error) {
				return "", nil
			})},
			error:       ErrTransformInputTypeIsDifferent,
			errorDetail: "int transform input not equals with string Map step in type",
		},
		{
			name:        "first_intype_is_empty",
			wrappers:    []StepWrapper{{Name: "Test", InTypes: [maxArgs]reflect.Type{}}},
			error:       ErrEmptyFirstStepInType,
			errorDetail: "Test step in type is missing",
		},
		{
			name:        "outtypes_and_next_intypes_not_matching",
			wrappers:    []StepWrapper{mapFn, {Name: "Test", InTypes: [maxArgs]reflect.Type{reflect.TypeFor[string]()}, OutTypes: [maxArgs]reflect.Type{reflect.TypeFor[string]()}}},
			error:       ErrStepOutAndNextInTypeIsDifferent,
			errorDetail: "Map step at position 1 has int out type but Test step at position 2 expects string in type",
		},
		{
			name:        "outtype_is_empty",
			wrappers:    []StepWrapper{mapFn, {Name: "Test", InTypes: [maxArgs]reflect.Type{reflect.TypeFor[int]()}}},
			error:       ErrEmptyStepOutType,
			errorDetail: "Test step at position 2 has no out type",
		},
		{
			name: "pointers_used_when_not_needed",
			wrappers: []StepWrapper{mapFn, Map(func(in *int) (*int, error) {
				return nil, nil
			})},
			error:       ErrStepOutAndNextInTypeIsDifferent,
			errorDetail: "Map step at position 1 has int out type but Map step at position 2 expects *int in type",
		},
		{
			name:        "validation_handles_split_and_merge",
			wrappers:    []StepWrapper{mapFn, splitFn, mergeFn},
			error:       ErrInnerStepValidationFailed,
			errorDetail: "orm input not equals with *int Map step in type",
		},
		{
			name: "validation_called_inside_branch",
			wrappers: []StepWrapper{mapFn, splitFn, WithBranches[int](
				Steps(mapFn), Steps(Map(func(in *int) (*int, error) { return nil, nil })))},
			error:       ErrInnerStepValidationFailed,
			errorDetail: "WithBranches step at position 3 failed inner validation: transform input type is different from first step in type: int transform input not equals with *int Map step in type",
		},

		// split and merge
	}[5:6] {
		t.Run(sc.name, func(t *testing.T) {
			actual, err := GetValidatedSteps[int](sc.wrappers)
			require.Empty(t, actual)
			assert.ErrorIs(t, err, sc.error, fmt.Sprintf("unexpected error: %s", err.Error()))
			assert.Contains(t, err.Error(), sc.errorDetail, fmt.Sprintf("unexpected error detail: %s", err.Error()))
		})
	}
}

// handles primitives and structs
// handles slices and maps
//
//
// run aggregator
// skip
// yield simple
// yield indexed
