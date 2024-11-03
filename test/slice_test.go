package test

import (
	"strconv"
	"testing"

	"github.com/domahidizoltan/go-steps"
	s "github.com/domahidizoltan/go-steps/kind/slicesteps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformSliceAsRange(t *testing.T) {
	r, err := steps.TransformSlice([]string{"1", "2", "3", "4", "5"}).
		With(
			s.Map(func(i string) (int, error) {
				return strconv.Atoi(i)
			}),
			s.Filter(func(i int) (bool, error) {
				return i%2 == 0, nil
			}),
			s.MultiplyBy(3),
			s.Map(func(i int) (string, error) {
				return "_" + strconv.Itoa(i*2), nil
			}),
		).AsRange()
	require.NoError(t, err)

	expected := []string{"_6", "_18", "_30"}
	actual := []string{}
	for i := range r {
		actual = append(actual, i.(string))
	}
	assert.Len(t, actual, 3)
	assert.Equal(t, expected, actual)
}
