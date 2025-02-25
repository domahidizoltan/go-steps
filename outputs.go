package steps

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"reflect"

	"github.com/jszwec/csvutil"
)

func handleErrWithTrName[T any, IT inputType[T]](t stepsTransformer[T, IT], err error, errorHandler func(error)) {
	if len(t.options.Name) != 0 {
		err = fmt.Errorf("[%s] %w", t.options.Name, err)
	}
	errorHandler(err)
}

func (t stepsTransformer[T, IT]) AsRange() iter.Seq[any] {
	return func(yield func(any) bool) {
		if t.error != nil {
			handleErrWithTrName(t, t.error, t.options.ErrorHandler)
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
						handleErrWithTrName(t, err, t.options.ErrorHandler)
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
						handleErrWithTrName(t, err, t.options.ErrorHandler)
					}
					break
				}
			}
		default:
			panic("unsupported input type")
		}
	}
}

func (t stepsTransformer[T, IT]) AsKeyValueRange() iter.Seq2[any, any] {
	return t.AsIndexedRange()
}

func (t stepsTransformer[T, IT]) AsIndexedRange() iter.Seq2[any, any] {
	return func(yield func(any, any) bool) {
		if t.error != nil {
			handleErrWithTrName(t, t.error, t.options.ErrorHandler)
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
						handleErrWithTrName(t, err, t.options.ErrorHandler)
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
						handleErrWithTrName(t, err, t.options.ErrorHandler)
					}
					break
				}
			}
		default:
			panic("unsupported input type")
		}
	}
}

func (t stepsTransformer[T, IT]) AsMultiMap() map[any][]any {
	var acc any
	for _, v := range t.AsIndexedRange() {
		acc = v
	}

	if acc == nil {
		return nil
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

func (t stepsTransformer[T, IT]) AsMap() map[any]any {
	res := map[any]any{}
	for k, v := range t.AsIndexedRange() {
		res[k] = v
	}
	return res
}

func (t stepsTransformer[T, IT]) AsSlice() []any {
	var res []any
	for v := range t.AsRange() {
		res = append(res, v)
	}

	return res
}

func (t stepsTransformer[T, IT]) AsCsv() string {
	data := t.AsSlice()
	if len(data) == 0 {
		return ""
	}

	dataType := reflect.Indirect(reflect.ValueOf(data[0])).Type()
	typedSlice := reflect.MakeSlice(reflect.SliceOf(dataType), len(data), len(data))

	for i, v := range data {
		typedSlice.Index(i).Set(reflect.ValueOf(v))
	}

	res, err := csvutil.Marshal(typedSlice.Interface())
	if err != nil {
		handleErrWithTrName(t, err, t.options.ErrorHandler)
	}
	return string(res)
}

func (t stepsTransformer[T, IT]) ToStreamingCsv(writer io.Writer) {
	w := csv.NewWriter(writer)
	enc := csvutil.NewEncoder(w)

	for record := range t.AsRange() {
		err := enc.Encode(record)
		if err != nil {
			handleErrWithTrName(t, err, t.options.ErrorHandler)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		handleErrWithTrName(t, err, t.options.ErrorHandler)
	}
}

func (t stepsTransformer[T, IT]) AsJson() string {
	data := t.AsSlice()
	res, err := json.Marshal(data)
	if err != nil {
		handleErrWithTrName(t, err, t.options.ErrorHandler)
	}
	return string(res)
}

func (t stepsTransformer[T, IT]) ToStreamingJson(writer io.Writer) {
	enc := json.NewEncoder(writer)

	for record := range t.AsRange() {
		err := enc.Encode(record)
		if err != nil {
			handleErrWithTrName(t, err, t.options.ErrorHandler)
		}
	}
}
