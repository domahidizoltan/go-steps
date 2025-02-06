package steps

import (
	"errors"
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
