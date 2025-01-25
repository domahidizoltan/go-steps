package test

import (
	"strconv"
	"testing"

	s "github.com/domahidizoltan/go-steps"
	"github.com/domahidizoltan/go-steps/test/customwrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformChanAsRange(t *testing.T) {
	closeCh := make(chan struct{}, 1)
	inCh := make(chan string, 5)

	r := s.Transform[string](inCh).
		With(s.Steps(
			s.Map(func(i string) (int, error) {
				return strconv.Atoi(i)
			}),
			s.Filter(func(i int) (bool, error) {
				return i%2 == 0, nil
			}),
			customwrapper.MultiplyBy(3),
			s.Map(func(i int) (string, error) {
				return "_" + strconv.Itoa(i*2), nil
			}),
		)).AsRange(func(err error) {
		require.NoError(t, err)
	})

	go func(in chan string) {
		in <- "1"
		in <- "2"
		in <- "3"
		in <- "4"
		in <- "5"

		closeCh <- struct{}{}
		close(in)
	}(inCh)

	expected := []string{"_12", "_24"}
	actual := []string{}
	for i := range r {
		actual = append(actual, i.(string))
	}

	<-closeCh
	assert.Len(t, actual, 2)
	assert.Equal(t, expected, actual)
}
