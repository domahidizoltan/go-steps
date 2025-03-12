package test

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	. "github.com/domahidizoltan/go-steps"
	"github.com/stretchr/testify/assert"
)

func TestOneByOneProcessing(t *testing.T) {
	iter := Transform[int]([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
		With(Steps(
			Take[int](7), // 1, 2, 3, 4, 5, 6, 7
			Filter(func(i int) (bool, error) {
				return i%2 == 0, nil
			}), // 2, 4, 6
			Map(func(i int) (string, error) {
				return strconv.Itoa(i * i), nil
			}), // "4", "16", "36"
		)).
		AsRange()

	res := []string{}
	for i := range iter {
		res = append(res, i.(string))
	}

	assert.Equal(t, []string{"4", "16", "36"}, res)
}

func TestOneByOneProcessing_WithChannel(t *testing.T) {
	inputCh := make(chan int, 3)
	iter := Transform[int](inputCh).
		With(Steps(
			Take[int](7), // 1, 2, 3, 4, 5, 6, 7
			Filter(func(i int) (bool, error) {
				return i%2 == 0, nil
			}), // 2, 4, 6
			Map(func(i int) (string, error) {
				return strconv.Itoa(i * i), nil
			}), // "4", "16", "36"
		)).
		AsRange()

	go func() {
		for _, i := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
			inputCh <- i
		}
		close(inputCh)
	}()

	res := []string{}
	for i := range iter {
		res = append(res, i.(string))
	}

	assert.Equal(t, []string{"4", "16", "36"}, res)
}

func TestTransformerOptions(t *testing.T) {
	buf := bytes.NewBufferString("")

	res := Transform[int]([]int{1, 2, 3}, WithLogWriter(buf)).
		WithSteps(
			Log("before filter"),
			Filter(func(i int) (bool, error) {
				return i%2 == 1, nil
			}),
			Log("after filter"),
		).AsSlice()

	assert.Equal(t, []any{1, 3}, res)
	expectedLogs := `before filter 	arg0: 1 
after filter 	arg0: 1 
before filter 	arg0: 2 
before filter 	arg0: 3 
after filter 	arg0: 3 
`
	assert.Equal(t, expectedLogs, buf.String())
}

func TestValidation(t *testing.T) {
	steps := Steps(
		Filter(func(i int) (bool, error) {
			return i%2 == 1, nil
		}),
		Map(func(in string) (int, error) {
			return strconv.Atoi(in)
		}),
	)

	if err := steps.Validate(); err != nil {
		assert.Error(t, err)
	}

	res := Transform[int]([]int{1, 2, 3}).
		With(steps).
		AsSlice()

	assert.Nil(t, res)
}

func TestErrorHandler(t *testing.T) {
	var transformErr error
	errHandler := func(err error) {
		transformErr = err
	}

	iter := Transform[int]([]int{10, 0, -10}, WithErrorHandler(errHandler)).
		WithSteps(
			Map(func(i int) (int, error) {
				if i == 0 {
					return 0, fmt.Errorf("division by zero: input=%d", i)
				}
				return i / 10, nil
			}),
		).AsRange()

	res := []int{}
	for i := range iter {
		res = append(res, i.(int))
	}
	assert.Equal(t, []int{1}, res)

	if transformErr != nil {
		assert.ErrorContains(t, transformErr, "division by zero: input=0")
	}
}

func TestGroupBy(t *testing.T) {
	type (
		person struct {
			name   string
			age    int
			isMale bool
		}
		ageRange uint8
	)

	const (
		young ageRange = iota
		mature
		old
	)

	persons := []person{
		{"John", 30, true},
		{"Jane", 25, false},
		{"Bob", 75, true},
		{"Alice", 28, false},
		{"Charlie", 17, true},
		{"Frank", 81, true},
		{"Bill", 45, true},
	}

	res := Transform[person](persons).
		With(Steps(
			Filter(func(p person) (bool, error) {
				return p.isMale, nil
			}),
		).Aggregate(
			GroupBy(func(p person) (ageRange, string, error) {
				switch {
				case p.age <= 18:
					return young, p.name, nil
				case p.age <= 65:
					return mature, p.name, nil
				default:
					return old, p.name, nil
				}
			}),
		),
		).
		AsMultiMap()

	expected := map[any][]any{
		young:  {"Charlie"},
		mature: {"John", "Bill"},
		old:    {"Bob", "Frank"},
	}
	assert.Equal(t, expected, res)
}

func TestAsMap(t *testing.T) {
	inputCh := make(chan int, 3)
	go func() {
		for _, i := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
			inputCh <- i
		}
		close(inputCh)
	}()

	res := Transform[int](inputCh).
		With(Steps(
			Take[int](7), // 1, 2, 3, 4, 5, 6, 7
			Filter(func(i int) (bool, error) {
				return i%2 == 0, nil
			}), // 2, 4, 6
			Map(func(i int) (string, error) {
				return strconv.Itoa(i * i), nil
			}), // "4", "16", "36"
		)).
		AsMap()

	assert.Equal(t, map[any]any{1: "4", 3: "16", 5: "36"}, res)
}

func multiplyBy[IN0 ~int](multiplier IN0) StepWrapper {
	return StepWrapper{
		Name: "multiplyBy",
		StepFn: func(in StepInput) StepOutput {
			return StepOutput{
				Args:    Args{in.Args[0].(IN0) * multiplier},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
		Validate: func(prevStepOut ArgTypes) (ArgTypes, error) {
			return ArgTypes{reflect.TypeFor[IN0]()}, nil
		},
	}
}

func primeInput(opts TransformerOptions) chan int {
	inputCh := make(chan int, 3)
	go func() {
		for _, i := range []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29} {
			inputCh <- i
		}
		close(inputCh)
	}()
	return inputCh
}

func TestCustomInputAndStep(t *testing.T) {
	res := TransformFn[int](primeInput).
		WithSteps(multiplyBy(3)).
		AsSlice()

	assert.Equal(t, []any{6, 9, 15, 21, 33, 39, 51, 57, 69, 87}, res)
}

func TestBranch(t *testing.T) {
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

	assert.Equal(t, []any{"10", "-2", "30", "-4", "50", "-6", "70", "-8", "90", "-10"}, res)
}

func TestCsvAndJson(t *testing.T) {
	type salary struct {
		ID         int    `csv:"ID" json:"id"`
		Name       string `csv:"Name" json:"name"`
		Age        int    `csv:"Age" json:"age"`
		Department string `csv:"Department" json:"department"`
		Salary     int    `csv:"Salary" json:"salary"`
		City       string `csv:"City" json:"city"`
	}

	res := TransformFn[salary](FromStreamingCsv[salary](File("testdata/salaries.csv"), false)).
		WithSteps(
			Filter(func(s salary) (bool, error) {
				return s.City == "New York" && s.Department == "Engineering", nil
			}),
			Take[salary](2),
		).
		AsJson()

	expected := "[{\"id\":1,\"name\":\"John Doe\",\"age\":25,\"department\":\"Engineering\",\"salary\":60000,\"city\":\"New York\"},{\"id\":21,\"name\":\"Mark King\",\"age\":36,\"department\":\"Engineering\",\"salary\":78000,\"city\":\"New York\"}]"
	assert.Equal(t, expected, res)
}
