package steps

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mapStringToInt = Map(func(in string) (int, error) {
		return strconv.Atoi(in)
	})
	filterEven = Filter(func(in int) (bool, error) {
		return in%2 == 0, nil
	})
	filterStr = Filter(func(in string) (bool, error) {
		return len(in) > 0, nil
	})
	groupBy = GroupBy(func(in int) (int, []int, error) {
		return in % 2, []int{in}, nil
	})
)

func TestStepsCreation(t *testing.T) {
	type scenario struct {
		name                      string
		steps                     StepsBranch
		expectedStepWrappers      []StepWrapper
		expectedAggregatorWrapper *ReducerWrapper
	}

	for _, sc := range []scenario{
		{
			name:  "empty_steps",
			steps: Steps(),
		}, {
			name:                 "steps_added",
			steps:                Steps(mapStringToInt, filterEven),
			expectedStepWrappers: []StepWrapper{mapStringToInt, filterEven},
		}, {
			name:                      "aggregator_added_without_steps",
			steps:                     Steps().Aggregate(groupBy),
			expectedAggregatorWrapper: &groupBy,
		}, {
			name:                      "steps_and_aggregator_added",
			steps:                     Steps(mapStringToInt, filterEven).Aggregate(groupBy),
			expectedStepWrappers:      []StepWrapper{mapStringToInt, filterEven},
			expectedAggregatorWrapper: &groupBy,
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			actualSteps := sc.steps
			assert.NoError(t, actualSteps.Error)
			matchStepWrappers(t, sc.expectedStepWrappers, actualSteps.StepWrappers)
			matchAggregatorWrapper(t, sc.expectedAggregatorWrapper, actualSteps.AggregatorWrapper)
		})
	}
}

func TestStepsValidationAndCreation(t *testing.T) {
	type scenario struct {
		name                      string
		steps                     StepsBranch
		expectedSteps             []StepFn
		expectedStepWrappers      []StepWrapper
		expectedAggregatorWrapper *ReducerWrapper
		hasError                  bool
	}

	for _, sc := range []scenario{
		{
			name: "validate_empty_steps",
		}, {
			name:                 "steps_validated_without_error",
			steps:                Steps(mapStringToInt, filterEven),
			expectedSteps:        []StepFn{mapStringToInt.StepFn, filterEven.StepFn},
			expectedStepWrappers: []StepWrapper{mapStringToInt, filterEven},
		}, {
			name:                 "steps_validated_with_error",
			steps:                Steps(filterEven, mapStringToInt),
			expectedStepWrappers: []StepWrapper{filterEven, mapStringToInt},
			hasError:             true,
		}, {
			name:                      "aggregator_added_without_steps",
			steps:                     Steps().Aggregate(groupBy),
			expectedAggregatorWrapper: &groupBy,
		}, {
			name:                      "steps_and_aggregator_validated_without_error",
			steps:                     Steps(mapStringToInt, filterEven).Aggregate(groupBy),
			expectedSteps:             []StepFn{mapStringToInt.StepFn, filterEven.StepFn},
			expectedStepWrappers:      []StepWrapper{mapStringToInt, filterEven},
			expectedAggregatorWrapper: &groupBy,
		}, {
			name:                      "steps_and_aggregator_validated_with_error",
			steps:                     Steps(filterStr).Aggregate(groupBy),
			expectedSteps:             []StepFn{filterStr.StepFn},
			expectedStepWrappers:      []StepWrapper{filterStr},
			expectedAggregatorWrapper: &groupBy,
			hasError:                  true,
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			err := sc.steps.Validate()
			actualSteps := sc.steps

			if sc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			matchSteps(t, sc.expectedSteps, actualSteps.Steps)
			matchStepWrappers(t, sc.expectedStepWrappers, actualSteps.StepWrappers)
			matchAggregatorWrapper(t, sc.expectedAggregatorWrapper, actualSteps.AggregatorWrapper)
		})
	}
}

func TestSliceTransformator(t *testing.T) {
	input := []string{"1", "2", "3", "4", "5"}
	type scenario struct {
		name               string
		transformator      stepsTransformator[string, []string]
		expectedData       []string
		expectedSteps      []StepFn
		expectedAggregator ReducerFn
		hasError           bool
	}

	for _, sc := range []scenario{
		{
			name: "transform_using_steps",
			transformator: Transform[string](input).
				With(Steps(mapStringToInt, filterEven)),
			expectedData:  input,
			expectedSteps: []StepFn{mapStringToInt.StepFn, filterEven.StepFn},
		}, {
			name: "transform_using_withsteps",
			transformator: Transform[string](input).
				WithSteps(mapStringToInt, filterEven),
			expectedData:  input,
			expectedSteps: []StepFn{mapStringToInt.StepFn, filterEven.StepFn},
		}, {
			name: "transform_using_steps_and_aggregate",
			transformator: Transform[string](input).
				With(Steps(mapStringToInt, filterEven).
					Aggregate(groupBy)),
			expectedData:       input,
			expectedSteps:      []StepFn{mapStringToInt.StepFn, filterEven.StepFn},
			expectedAggregator: groupBy.ReducerFn,
		}, {
			name: "transform_with_error",
			transformator: Transform[string](input).
				With(Steps(filterStr).Aggregate(groupBy)),
			hasError: true,
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			assert.Equal(t, sc.expectedData, sc.transformator.in)
			matchSteps(t, sc.expectedSteps, sc.transformator.Steps)
			matchAggregator(t, sc.expectedAggregator, sc.transformator.Aggregator)

			if sc.hasError {
				assert.Error(t, sc.transformator.Error)
			} else {
				assert.NoError(t, sc.transformator.Error)
			}
		})
	}
}

func TestChanTransformator(t *testing.T) {
	input := make(chan string, 5)
	type scenario struct {
		name               string
		transformator      stepsTransformator[string, chan string]
		expectedData       chan string
		expectedSteps      []StepFn
		expectedAggregator ReducerFn
		hasError           bool
	}

	for _, sc := range []scenario{
		{
			name: "transform_using_steps",
			transformator: Transform[string](input).
				With(Steps(mapStringToInt, filterEven)),
			expectedData:  input,
			expectedSteps: []StepFn{mapStringToInt.StepFn, filterEven.StepFn},
		}, {
			name: "transform_using_withsteps",
			transformator: Transform[string](input).
				WithSteps(mapStringToInt, filterEven),
			expectedData:  input,
			expectedSteps: []StepFn{mapStringToInt.StepFn, filterEven.StepFn},
		}, {
			name: "transform_using_steps_and_aggregate",
			transformator: Transform[string](input).
				With(Steps(mapStringToInt, filterEven).
					Aggregate(groupBy)),
			expectedData:       input,
			expectedSteps:      []StepFn{mapStringToInt.StepFn, filterEven.StepFn},
			expectedAggregator: groupBy.ReducerFn,
		}, {
			name: "transform_with_error",
			transformator: Transform[string](input).
				With(Steps(filterStr).Aggregate(groupBy)),
			hasError: true,
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			assert.Equal(t, sc.expectedData, sc.transformator.in)
			matchSteps(t, sc.expectedSteps, sc.transformator.Steps)
			matchAggregator(t, sc.expectedAggregator, sc.transformator.Aggregator)

			if sc.hasError {
				assert.Error(t, sc.transformator.Error)
			} else {
				assert.NoError(t, sc.transformator.Error)
			}
		})
	}
}

func matchSteps(t *testing.T, expected, actual []StepFn) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("steps length are different: %d != %d", len(expected), len(actual))
	}

	for i, exp := range expected {
		if !funcPtrAreEqual(actual[i], exp) {
			t.Errorf("steps are different at pos %d", i)
		}
	}
}

func matchStepWrappers(t *testing.T, expected, actual []StepWrapper) {
	t.Helper()
	for i, exp := range expected {
		if !funcPtrAreEqual(actual[i].StepFn, exp.StepFn) {
			t.Errorf("stepWrappers are different at pos %d: %s != %s", i, actual[i].Name, exp.Name)
		}
	}
}

func matchAggregator(t *testing.T, expected, actual ReducerFn) {
	t.Helper()
	if !funcPtrAreEqual(expected, actual) {
		t.Errorf("aggregator is different: %v != %v", actual, expected)
	}
}

func matchAggregatorWrapper(t *testing.T, expected, actual *ReducerWrapper) {
	t.Helper()
	if expected == nil {
		if actual != nil {
			t.Errorf("aggregatorWrapper is different: %s != nil", actual.Name)
		}
		return
	}

	if !funcPtrAreEqual(actual.ReducerFn, expected.ReducerFn) {
		t.Errorf("aggregatorWrapper is different: %s != %s", actual.Name, expected.Name)
	}
}

func funcPtrAreEqual(expected, actual any) bool {
	expectedFn := reflect.ValueOf(expected)
	actualFn := reflect.ValueOf(actual)

	return expectedFn.Pointer() == actualFn.Pointer()
}
