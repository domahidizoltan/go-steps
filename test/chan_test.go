package test

import (
	"strconv"
	"testing"

	"github.com/domahidizoltan/go-steps"
	c "github.com/domahidizoltan/go-steps/kind/chansteps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformChanAsRange(t *testing.T) {
	closeCh := make(chan struct{}, 1)
	inCh := make(chan string, 5)

	r, err := steps.TransformChan(inCh).
		With(
			c.Map(func(i string) (int, error) {
				return strconv.Atoi(i)
			}),
			c.Filter(func(i int) (bool, error) {
				return i%2 == 0, nil
			}),
			c.MultiplyBy(3),
			c.Map(func(i int) (string, error) {
				return "_" + strconv.Itoa(i*2), nil
			}),
		).AsRange()

	go func(in chan string) {
		in <- "1"
		in <- "2"
		in <- "3"
		in <- "4"
		in <- "5"

		closeCh <- struct{}{}
		close(in)
	}(inCh)

	expected := []string{"_6", "_18", "_30"}
	actual := []string{}
	for i := range r {
		actual = append(actual, i.(string))
	}

	<-closeCh
	require.NoError(t, err)
	assert.Len(t, actual, 3)
	assert.Equal(t, expected, actual)
}
