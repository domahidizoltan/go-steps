package steps

import (
	"fmt"
	"iter"
	"reflect"
)

func handleErrWithTrName[T any, IT inputType[T]](t stepsTransformer[T, IT], err error, errorHandler func(error)) {
	if len(t.options.Name) != 0 {
		err = fmt.Errorf("[%s] %w", t.options.Name, err)
	}
	errorHandler(err)
}

func emptyErrorHandler() func(error) {
	return func(err error) {
		fmt.Println("error occured:", err)
	}
}

func (t stepsTransformer[T, IT]) AsRange(errorHandler func(error)) iter.Seq[any] {
	if errorHandler == nil {
		errorHandler = emptyErrorHandler()
	}

	return func(yield func(any) bool) {
		if t.error != nil {
			handleErrWithTrName(t, t.error, errorHandler)
			return
		}

		var terminated bool
		var err error
		switch in := any(t.input).(type) {
		case chan T:
			// TODO check if closed for lastItem
			for v := range in {
				_, terminated, err = process(v, yield, &t.transformer, false)
				if terminated || err != nil {
					if err != nil {
						handleErrWithTrName(t, err, errorHandler)
					}
					break
				}
			}
		case []T:
			lastIdx := len(in) - 1
			for idx, v := range in {
				_, terminated, err = process(v, yield, &t.transformer, idx == lastIdx)
				if terminated || err != nil {
					if err != nil {
						handleErrWithTrName(t, err, errorHandler)
					}
					break
				}
			}
		default:
			panic("unsupported input type")
		}
	}
}

func (t stepsTransformer[T, IT]) AsKeyValueRange(errorHandler func(error)) iter.Seq2[any, any] {
	return t.AsIndexedRange(errorHandler)
}

func (t stepsTransformer[T, IT]) AsIndexedRange(errorHandler func(error)) iter.Seq2[any, any] {
	if errorHandler == nil {
		errorHandler = emptyErrorHandler()
	}

	return func(yield func(any, any) bool) {
		if t.error != nil {
			handleErrWithTrName(t, t.error, errorHandler)
			return
		}

		var terminated bool
		var err error
		switch in := any(t.input).(type) {
		case chan T:
			idx := 0
			for v := range in {
				_, terminated, err = processIndexed(idx, v, yield, &t.transformer, false)
				if terminated || err != nil {
					if err != nil {
						handleErrWithTrName(t, err, errorHandler)
					}
					break
				}
				idx++
			}
		case []T:
			lastIdx := len(in) - 1
			for idx, v := range in {
				_, terminated, err = processIndexed(idx, v, yield, &t.transformer, idx == lastIdx)
				if terminated || err != nil {
					if err != nil {
						handleErrWithTrName(t, err, errorHandler)
					}
					break
				}
			}
		default:
			panic("unsupported input type")
		}
	}
}

func (t stepsTransformer[T, IT]) AsMultiMap(errorHandler func(error)) map[any][]any {
	var acc any
	for _, v := range t.AsIndexedRange(errorHandler) {
		acc = v
	}

	res := map[any][]any{}
	iter := reflect.ValueOf(acc).MapRange()
	for iter.Next() {
		v := iter.Value()
		vRes := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			vRes[i] = v.Index(i).Interface()
		}
		k := iter.Key().Interface()
		res[k] = vRes
	}
	return res
}

func (t stepsTransformer[T, IT]) AsMap(errorHandler func(error)) map[any]any {
	res := map[any]any{}
	for k, v := range t.AsIndexedRange(errorHandler) {
		res[k] = v
	}
	return res
}

func (t stepsTransformer[T, IT]) AsSlice(errorHandler func(error)) []any {
	var res []any
	for v := range t.AsRange(errorHandler) {
		res = append(res, v)
	}

	return res
}
