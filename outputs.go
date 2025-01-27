package steps

import (
	"fmt"
	"iter"
	"reflect"
)

func emptyErrorHandler[T any, IT inputType[T]](t stepsTransformer[T, IT]) func(error) {
	return func(err error) {
		_ = t
		fmt.Println("error occured:", err)
	}
}

func (t stepsTransformer[T, IT]) AsRange(errorHandler func(error)) iter.Seq[any] {
	if errorHandler == nil {
		errorHandler = emptyErrorHandler(t)
	}

	return func(yield func(any) bool) {
		if t.error != nil {
			errorHandler(t.error)
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
						errorHandler(err)
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
						errorHandler(err)
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
		errorHandler = emptyErrorHandler(t)
	}

	return func(yield func(any, any) bool) {
		if t.error != nil {
			errorHandler(t.error)
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
						errorHandler(err)
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
						errorHandler(err)
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
		fmt.Printf("__ %+v\n", reflect.ValueOf(v))
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
