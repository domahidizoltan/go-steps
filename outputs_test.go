package steps

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const tfName = "TFName"

var (
	errTransformer = errors.New("transformer error")
	errStep        = errors.New("step error")
	errorFn        = Map(func(in int) (int, error) {
		return 0, errStep
	})
	opts = TransformerOptions{Name: tfName}
)

func TestAsRange_WithSlice(t *testing.T) {
	type scenario struct {
		name           string
		transformer    stepsTransformer[int, []int]
		expectedOutput []any
		expectedError  error
	}

	for _, sc := range []scenario{
		{
			name: "empty_input",
			transformer: stepsTransformer[int, []int]{
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
		}, {
			name: "empty_steps",
			transformer: stepsTransformer[int, []int]{
				transformer: transformer{},
				input:       []int{1, 2, 3, 4, 5},
			},
			expectedOutput: []any{1, 2, 3, 4, 5},
		}, {
			name: "step_error_returned",
			transformer: stepsTransformer[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, errorFn.StepFn},
				},
			},
			expectedError: errStep,
		}, {
			name: "transformer_error_returned",
			transformer: stepsTransformer[int, []int]{
				transformer: transformer{
					options: opts,
					error:   errTransformer,
				},
			},
			expectedError: errTransformer,
		}, {
			name: "input_processed",
			transformer: stepsTransformer[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
			expectedOutput: []any{3, 5},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			sc.transformer.options.ErrorHandler = func(err error) {
				assert.True(t, strings.HasPrefix(err.Error(), "["+tfName))
				assert.ErrorIs(t, err, sc.expectedError)
			}
			iter := sc.transformer.AsRange()

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
		transformer    stepsTransformer[int, []int]
		expectedOutput map[any]any
		expectedError  error
	}

	for _, sc := range []scenario{
		{
			name: "empty_input",
			transformer: stepsTransformer[int, []int]{
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
		}, {
			name: "empty_steps",
			transformer: stepsTransformer[int, []int]{
				input: []int{1, 2, 3, 4, 5},
			},
			expectedOutput: map[any]any{0: 1, 1: 2, 2: 3, 3: 4, 4: 5},
		}, {
			name: "step_error_returned",
			transformer: stepsTransformer[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, errorFn.StepFn},
				},
			},
			expectedError: errStep,
		}, {
			name: "transformer_error_returned",
			transformer: stepsTransformer[int, []int]{
				transformer: transformer{
					options: opts,
					error:   errTransformer,
				},
			},
			expectedError: errTransformer,
		}, {
			name: "input_processed",
			transformer: stepsTransformer[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
			expectedOutput: map[any]any{1: 3, 3: 5},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			sc.transformer.options.ErrorHandler = func(err error) {
				assert.True(t, strings.HasPrefix(err.Error(), "["+tfName))
				assert.ErrorIs(t, err, sc.expectedError)
			}
			iter := sc.transformer.AsIndexedRange()

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
		transformer    stepsTransformer[int, chan int]
		inputValues    []int
		expectedOutput []any
		expectedError  error
	}

	for _, sc := range []scenario{
		{
			name: "empty_input",
			transformer: stepsTransformer[int, chan int]{
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
		}, {
			name:           "empty_steps",
			transformer:    stepsTransformer[int, chan int]{},
			inputValues:    []int{1, 2, 3, 4, 5},
			expectedOutput: []any{1, 2, 3, 4, 5},
		}, {
			name: "step_error_returned",
			transformer: stepsTransformer[int, chan int]{
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, errorFn.StepFn},
				},
			},
			inputValues:   []int{1, 2, 3, 4, 5},
			expectedError: errStep,
		}, {
			name: "transformer_error_returned",
			transformer: stepsTransformer[int, chan int]{
				transformer: transformer{
					options: opts,
					error:   errTransformer,
				},
			},
			expectedError: errTransformer,
		}, {
			name: "input_processed",
			transformer: stepsTransformer[int, chan int]{
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
			inputValues:    []int{1, 2, 3, 4, 5},
			expectedOutput: []any{3, 5},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			inputCh := make(chan int, 10)
			sc.transformer.input = inputCh
			sc.transformer.options.ErrorHandler = func(err error) {
				assert.True(t, strings.HasPrefix(err.Error(), "["+tfName))
				assert.ErrorIs(t, err, sc.expectedError)
			}

			iter := sc.transformer.AsRange()

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
		transformer    stepsTransformer[int, chan int]
		inputValues    []int
		expectedOutput map[any]any
		expectedError  error
	}

	for _, sc := range []scenario{
		{
			name: "empty_input",
			transformer: stepsTransformer[int, chan int]{
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
		}, {
			name:           "empty_steps",
			transformer:    stepsTransformer[int, chan int]{},
			inputValues:    []int{1, 2, 3, 4, 5},
			expectedOutput: map[any]any{0: 1, 1: 2, 2: 3, 3: 4, 4: 5},
		}, {
			name: "step_error_returned",
			transformer: stepsTransformer[int, chan int]{
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, errorFn.StepFn},
				},
			},
			inputValues:   []int{1, 2, 3, 4, 5},
			expectedError: errStep,
		}, {
			name: "transformer_error_returned",
			transformer: stepsTransformer[int, chan int]{
				transformer: transformer{
					options: opts,
					error:   errTransformer,
				},
			},
			expectedError: errTransformer,
		}, {
			name: "input_processed",
			transformer: stepsTransformer[int, chan int]{
				transformer: transformer{
					options: opts,
					steps:   []StepFn{filterFn.StepFn, mapFn.StepFn},
				},
			},
			inputValues:    []int{1, 2, 3, 4, 5},
			expectedOutput: map[any]any{1: 3, 3: 5},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			inputCh := make(chan int, 10)
			sc.transformer.input = inputCh
			sc.transformer.options.ErrorHandler = func(err error) {
				assert.True(t, strings.HasPrefix(err.Error(), "["+tfName))
				assert.ErrorIs(t, err, sc.expectedError)
			}

			iter := sc.transformer.AsIndexedRange()

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
	transformer := stepsTransformer[int, []int]{
		input: []int{1, 2, 3, 4, 5},
		transformer: transformer{
			options: opts,
			steps:   []StepFn{errorFn.StepFn},
		},
	}
	transformer.options.ErrorHandler = func(err error) {}
	var res []any
	for i := range transformer.AsRange() {
		res = append(res, i)
	}
	assert.Nil(t, res)
}

func TestAsIndexedRange_WithSlice_WithoutErrorHandler(t *testing.T) {
	transformer := stepsTransformer[int, []int]{
		input: []int{1, 2, 3, 4, 5},
		transformer: transformer{
			options: opts,
			steps:   []StepFn{errorFn.StepFn},
		},
	}
	transformer.options.ErrorHandler = func(err error) {}
	res := map[any]any{}
	for idx, i := range transformer.AsIndexedRange() {
		res[idx] = i
	}
	assert.Empty(t, res)
}

func TestAsRange_WithChan_WithoutErrorHandler(t *testing.T) {
	inputCh := make(chan int, 10)
	transformer := stepsTransformer[int, chan int]{
		input: inputCh,
		transformer: transformer{
			options: opts,
			steps:   []StepFn{errorFn.StepFn},
		},
	}
	transformer.options.ErrorHandler = func(err error) {}
	go func(inputCh chan int) {
		for _, v := range []int{1, 2, 3, 4, 5} {
			inputCh <- v
		}
		close(inputCh)
	}(inputCh)

	var res []any
	for i := range transformer.AsRange() {
		res = append(res, i)
	}
	assert.Nil(t, res)
}

func TestAsIndexedRange_WithChan_WithoutErrorHandler(t *testing.T) {
	inputCh := make(chan int, 10)
	transformer := stepsTransformer[int, chan int]{
		input: inputCh,
		transformer: transformer{
			options: opts,
			steps:   []StepFn{errorFn.StepFn},
		},
	}
	transformer.options.ErrorHandler = func(err error) {}
	go func(inputCh chan int) {
		for _, v := range []int{1, 2, 3, 4, 5} {
			inputCh <- v
		}
		close(inputCh)
	}(inputCh)

	var res []any
	for i := range transformer.AsIndexedRange() {
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
// 			transformer := stepsTransformer[int, []int]{
// 				input: []int{1, 2, 3, 4, 5},
// 				transformer: transformer{
// 					aggregator: sc.reducer,
// 				},
// 			}
//
// 			res := transformer.AsMultiMap(sc.errorHandler)
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
				assert.True(t, strings.HasPrefix(err.Error(), "["+tfName))
				assert.ErrorIs(t, err, errStep)
			},
			steps: []StepFn{errorFn.StepFn},
		},
		{
			name:  "input_processed",
			steps: []StepFn{filterFn.StepFn},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			transformer := stepsTransformer[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformer: transformer{
					options: optsWithErrHandler(sc.errorHandler),
					steps:   sc.steps,
				},
			}

			res := transformer.AsMap()
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
				assert.True(t, strings.HasPrefix(err.Error(), "["+tfName))
				assert.ErrorIs(t, err, errStep)
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
			transformer := stepsTransformer[int, chan int]{
				input: inputCh,
				transformer: transformer{
					options: optsWithErrHandler(sc.errorHandler),
					steps:   sc.steps,
				},
			}

			go func(inputCh chan int) {
				for _, v := range []int{1, 2, 3, 4, 5} {
					inputCh <- v
				}
				close(inputCh)
			}(inputCh)

			res := transformer.AsMap()
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
				assert.True(t, strings.HasPrefix(err.Error(), "["+tfName))
				assert.ErrorIs(t, err, errStep)
			},
			steps: []StepFn{errorFn.StepFn},
		},
		{
			name:  "input_processed",
			steps: []StepFn{filterFn.StepFn},
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			transformer := stepsTransformer[int, []int]{
				input: []int{1, 2, 3, 4, 5},
				transformer: transformer{
					options: optsWithErrHandler(sc.errorHandler),
					steps:   sc.steps,
				},
			}
			res := transformer.AsSlice()
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
				assert.True(t, strings.HasPrefix(err.Error(), "["+tfName))
				assert.ErrorIs(t, err, errStep)
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
			transformer := stepsTransformer[int, chan int]{
				input: inputCh,
				transformer: transformer{
					options: optsWithErrHandler(sc.errorHandler),
					steps:   sc.steps,
				},
			}

			go func(inputCh chan int) {
				for _, v := range []int{1, 2, 3, 4, 5} {
					inputCh <- v
				}
				close(inputCh)
			}(inputCh)

			res := transformer.AsSlice()
			if sc.errorHandler == nil {
				assert.Equal(t, []any{2, 4}, res)
			}
		})
	}
}

func optsWithErrHandler(errHandler func(error)) TransformerOptions {
	opts := opts
	opts.ErrorHandler = errHandler
	return opts
}
