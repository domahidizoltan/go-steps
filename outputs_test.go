package steps

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	errTransformator = errors.New("transformator error")
	errStep          = errors.New("step error")
	errorFn          = Map(func(in int) (int, error) {
		return 0, errStep
	})
)

func TestAsRange_WithSlice(t *testing.T) {
	type scenario struct {
		name           string
		transformator  stepsTransformator[int, []int]
		expectedOutput []any
		expectedError  error
	}

	for _, sc := range []scenario{
		{
			name: "empty_input",
			transformator: stepsTransformator[int, []int]{
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
		}, {
			name: "empty_steps",
			transformator: stepsTransformator[int, []int]{
				input: []int{1, 2, 3, 4, 5},
			},
			expectedOutput: []any{1, 2, 3, 4, 5},
		}, {
			name: "step_error_returned",
			transformator: stepsTransformator[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, errorFn.StepFn},
				},
			},
			expectedError: errStep,
		}, {
			name: "transformator_error_returned",
			transformator: stepsTransformator[int, []int]{
				transformator: transformator{
					error: errTransformator,
				},
			},
			expectedError: errTransformator,
		}, {
			name: "input_processed",
			transformator: stepsTransformator[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
			expectedOutput: []any{3, 5},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			iter := sc.transformator.AsRange(func(err error) {
				assert.Equal(t, sc.expectedError, err)
			})

			var res []any
			for i := range iter {
				res = append(res, i)
			}
			assert.Equal(t, sc.expectedOutput, res)
		})
	}
}

func TestAsIndexedRange_WithSlice(t *testing.T) {
	type scenario struct {
		name           string
		transformator  stepsTransformator[int, []int]
		expectedOutput map[any]any
		expectedError  error
	}

	for _, sc := range []scenario{
		{
			name: "empty_input",
			transformator: stepsTransformator[int, []int]{
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
		}, {
			name: "empty_steps",
			transformator: stepsTransformator[int, []int]{
				input: []int{1, 2, 3, 4, 5},
			},
			expectedOutput: map[any]any{0: 1, 1: 2, 2: 3, 3: 4, 4: 5},
		}, {
			name: "step_error_returned",
			transformator: stepsTransformator[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, errorFn.StepFn},
				},
			},
			expectedError: errStep,
		}, {
			name: "transformator_error_returned",
			transformator: stepsTransformator[int, []int]{
				transformator: transformator{
					error: errTransformator,
				},
			},
			expectedError: errTransformator,
		}, {
			name: "input_processed",
			transformator: stepsTransformator[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
			expectedOutput: map[any]any{1: 3, 3: 5},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			iter := sc.transformator.AsIndexedRange(func(err error) {
				assert.Equal(t, sc.expectedError, err)
			})

			res := map[any]any{}
			for idx, i := range iter {
				res[idx] = i
			}

			if sc.expectedOutput == nil {
				sc.expectedOutput = map[any]any{}
			}
			assert.Equal(t, sc.expectedOutput, res)
		})
	}
}

func TestAsRange_WithChan(t *testing.T) {
	type scenario struct {
		name           string
		transformator  stepsTransformator[int, chan int]
		inputValues    []int
		expectedOutput []any
		expectedError  error
	}

	for _, sc := range []scenario{
		{
			name: "empty_input",
			transformator: stepsTransformator[int, chan int]{
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
		}, {
			name:           "empty_steps",
			transformator:  stepsTransformator[int, chan int]{},
			inputValues:    []int{1, 2, 3, 4, 5},
			expectedOutput: []any{1, 2, 3, 4, 5},
		}, {
			name: "step_error_returned",
			transformator: stepsTransformator[int, chan int]{
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, errorFn.StepFn},
				},
			},
			inputValues:   []int{1, 2, 3, 4, 5},
			expectedError: errStep,
		}, {
			name: "transformator_error_returned",
			transformator: stepsTransformator[int, chan int]{
				transformator: transformator{
					error: errTransformator,
				},
			},
			expectedError: errTransformator,
		}, {
			name: "input_processed",
			transformator: stepsTransformator[int, chan int]{
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
			inputValues:    []int{1, 2, 3, 4, 5},
			expectedOutput: []any{3, 5},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			inputCh := make(chan int, 10)
			sc.transformator.input = inputCh

			iter := sc.transformator.AsRange(func(err error) {
				assert.Equal(t, sc.expectedError, err)
			})

			go func(inputCh chan int) {
				for _, v := range sc.inputValues {
					inputCh <- v
				}
				close(inputCh)
			}(inputCh)

			var res []any
			for i := range iter {
				res = append(res, i)
			}

			assert.Equal(t, sc.expectedOutput, res)
		})
	}
}

func TestAsIndexedRange_WithChan(t *testing.T) {
	type scenario struct {
		name           string
		transformator  stepsTransformator[int, chan int]
		inputValues    []int
		expectedOutput map[any]any
		expectedError  error
	}

	for _, sc := range []scenario{
		{
			name: "empty_input",
			transformator: stepsTransformator[int, chan int]{
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
		}, {
			name:           "empty_steps",
			transformator:  stepsTransformator[int, chan int]{},
			inputValues:    []int{1, 2, 3, 4, 5},
			expectedOutput: map[any]any{0: 1, 1: 2, 2: 3, 3: 4, 4: 5},
		}, {
			name: "step_error_returned",
			transformator: stepsTransformator[int, chan int]{
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, errorFn.StepFn},
				},
			},
			inputValues:   []int{1, 2, 3, 4, 5},
			expectedError: errStep,
		}, {
			name: "transformator_error_returned",
			transformator: stepsTransformator[int, chan int]{
				transformator: transformator{
					error: errTransformator,
				},
			},
			expectedError: errTransformator,
		}, {
			name: "input_processed",
			transformator: stepsTransformator[int, chan int]{
				transformator: transformator{
					steps: []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
			inputValues:    []int{1, 2, 3, 4, 5},
			expectedOutput: map[any]any{1: 3, 3: 5},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			inputCh := make(chan int, 10)
			sc.transformator.input = inputCh

			iter := sc.transformator.AsIndexedRange(func(err error) {
				assert.Equal(t, sc.expectedError, err)
			})

			go func(inputCh chan int) {
				for _, v := range sc.inputValues {
					inputCh <- v
				}
				close(inputCh)
			}(inputCh)

			res := map[any]any{}
			for idx, i := range iter {
				res[idx] = i
			}

			if sc.expectedOutput == nil {
				sc.expectedOutput = map[any]any{}
			}
			assert.Equal(t, sc.expectedOutput, res)
		})
	}
}

func TestAsRange_WithSlice_WithoutErrorHandler(t *testing.T) {
	transformator := stepsTransformator[int, []int]{
		input: []int{1, 2, 3, 4, 5},
		transformator: transformator{
			steps: []StepFn{errorFn.StepFn},
		},
	}
	var res []any
	for i := range transformator.AsRange(nil) {
		res = append(res, i)
	}
	assert.Nil(t, res)
}

func TestAsIndexedRange_WithSlice_WithoutErrorHandler(t *testing.T) {
	transformator := stepsTransformator[int, []int]{
		input: []int{1, 2, 3, 4, 5},
		transformator: transformator{
			steps: []StepFn{errorFn.StepFn},
		},
	}
	res := map[any]any{}
	for idx, i := range transformator.AsIndexedRange(nil) {
		res[idx] = i
	}
	assert.Empty(t, res)
}

func TestAsRange_WithChan_WithoutErrorHandler(t *testing.T) {
	inputCh := make(chan int, 10)
	transformator := stepsTransformator[int, chan int]{
		input: inputCh,
		transformator: transformator{
			steps: []StepFn{errorFn.StepFn},
		},
	}
	go func(inputCh chan int) {
		for _, v := range []int{1, 2, 3, 4, 5} {
			inputCh <- v
		}
		close(inputCh)
	}(inputCh)

	var res []any
	for i := range transformator.AsRange(nil) {
		res = append(res, i)
	}
	assert.Nil(t, res)
}

func TestAsIndexedRange_WithChan_WithoutErrorHandler(t *testing.T) {
	inputCh := make(chan int, 10)
	transformator := stepsTransformator[int, chan int]{
		input: inputCh,
		transformator: transformator{
			steps: []StepFn{errorFn.StepFn},
		},
	}
	go func(inputCh chan int) {
		for _, v := range []int{1, 2, 3, 4, 5} {
			inputCh <- v
		}
		close(inputCh)
	}(inputCh)

	var res []any
	for i := range transformator.AsIndexedRange(nil) {
		res = append(res, i)
	}
	assert.Nil(t, res)
}

//TODO + chan
// func TestAsMultiMap_WithSlice(t *testing.T) {
// 	type scenario struct {
// 		name         string
// 		errorHandler func(error)
// 		reducer      ReducerFn
// 	}
//
// 	for _, sc := range []scenario{
// 		// {
// 		// 	name: "error_handler_passed",
// 		// 	errorHandler: func(err error) {
// 		// 		assert.Equal(t, errStep, err)
// 		// 	},
// 		// 	steps: []StepFn{errorFn.StepFn},
// 		// },
// 		{
// 			name: "input_processed",
// 			reducer: GroupBy(func(in int) (bool, int, error) {
// 				return in%2 == 0, in, nil
// 			}).ReducerFn,
// 		},
// 	} {
// 		t.Run(sc.name, func(t *testing.T) {
// 			transformator := stepsTransformator[int, []int]{
// 				input: []int{1, 2, 3, 4, 5},
// 				transformator: transformator{
// 					aggregator: sc.reducer,
// 				},
// 			}
//
// 			res := transformator.AsMultiMap(sc.errorHandler)
// 			if sc.errorHandler == nil {
// 				assert.Equal(t, map[any][]any{1: {2}, 3: {4}}, res)
// 			}
// 		})
// 	}
// }
//

func TestAsMap_WithSlice(t *testing.T) {
	type scenario struct {
		name         string
		errorHandler func(error)
		steps        []StepFn
	}

	for _, sc := range []scenario{
		{
			name: "error_handler_passed",
			errorHandler: func(err error) {
				assert.Equal(t, errStep, err)
			},
			steps: []StepFn{errorFn.StepFn},
		},
		{
			name:  "input_processed",
			steps: []StepFn{filterFn.StepFn},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			transformator := stepsTransformator[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformator: transformator{
					steps: sc.steps,
				},
			}

			res := transformator.AsMap(sc.errorHandler)
			if sc.errorHandler == nil {
				assert.Equal(t, map[any]any{1: 2, 3: 4}, res)
			}
		})
	}
}

func TestAsMap_WithChan(t *testing.T) {
	type scenario struct {
		name         string
		errorHandler func(error)
		steps        []StepFn
	}

	for _, sc := range []scenario{
		{
			name: "error_handler_passed",
			errorHandler: func(err error) {
				assert.Equal(t, errStep, err)
			},
			steps: []StepFn{errorFn.StepFn},
		},
		{
			name:  "input_processed",
			steps: []StepFn{filterFn.StepFn},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			inputCh := make(chan int, 10)
			transformator := stepsTransformator[int, chan int]{
				input: inputCh,
				transformator: transformator{
					steps: sc.steps,
				},
			}

			go func(inputCh chan int) {
				for _, v := range []int{1, 2, 3, 4, 5} {
					inputCh <- v
				}
				close(inputCh)
			}(inputCh)

			res := transformator.AsMap(sc.errorHandler)
			if sc.errorHandler == nil {
				assert.Equal(t, map[any]any{1: 2, 3: 4}, res)
			}
		})
	}
}

func TestAsSlice_WithSlice(t *testing.T) {
	type scenario struct {
		name         string
		errorHandler func(error)
		steps        []StepFn
	}

	for _, sc := range []scenario{
		{
			name: "error_handler_passed",
			errorHandler: func(err error) {
				assert.Equal(t, errStep, err)
			},
			steps: []StepFn{errorFn.StepFn},
		},
		{
			name:  "input_processed",
			steps: []StepFn{filterFn.StepFn},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			transformator := stepsTransformator[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformator: transformator{
					steps: sc.steps,
				},
			}
			res := transformator.AsSlice(sc.errorHandler)
			if sc.errorHandler == nil {
				assert.Equal(t, []any{2, 4}, res)
			}
		})
	}
}

func TestAsSlice_WithChan(t *testing.T) {
	type scenario struct {
		name         string
		errorHandler func(error)
		steps        []StepFn
	}

	for _, sc := range []scenario{
		{
			name: "error_handler_passed",
			errorHandler: func(err error) {
				assert.Equal(t, errStep, err)
			},
			steps: []StepFn{errorFn.StepFn},
		},
		{
			name:  "input_processed",
			steps: []StepFn{filterFn.StepFn},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			inputCh := make(chan int, 10)
			transformator := stepsTransformator[int, chan int]{
				input: inputCh,
				transformator: transformator{
					steps: sc.steps,
				},
			}

			go func(inputCh chan int) {
				for _, v := range []int{1, 2, 3, 4, 5} {
					inputCh <- v
				}
				close(inputCh)
			}(inputCh)

			res := transformator.AsSlice(sc.errorHandler)
			if sc.errorHandler == nil {
				assert.Equal(t, []any{2, 4}, res)
			}
		})
	}
}
