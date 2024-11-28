package steps

import (
	"iter"
	"reflect"

	is "github.com/domahidizoltan/go-steps/internal/pkg/step"
)

func (t transformator[T, I]) AsRange() (iter.Seq[any], error) {
	if t.Error != nil {
		return nil, t.Error
	}

	return func(yield func(any) bool) {
		switch in := any(t.in).(type) {
		case chan T:
			// TODO check if closed for lastItem
			for v := range in {
				if is.Process(v, yield, &t.Transformator, false) {
					break
				}
			}
		case []T:
			lastIdx := len(in) - 1
			for idx, v := range in {
				if is.Process(v, yield, &t.Transformator, idx == lastIdx) {
					break
				}
			}
		default:
			panic("unsupported input type")
		}
	}, nil
}

// TODO alias for AsKeyValueRange()
func (t transformator[T, I]) AsIndexedRange() (iter.Seq2[any, any], error) {
	if t.Error != nil {
		return nil, t.Error
	}

	return func(yield func(any, any) bool) {
		switch in := any(t.in).(type) {
		case chan T:
			idx := 0
			for v := range in {
				if is.ProcessIndexed(idx, v, yield, &t.Transformator, false) {
					break
				}
			}
		case []T:
			lastIdx := len(in) - 1
			for idx, v := range in {
				if is.ProcessIndexed(idx, v, yield, &t.Transformator, idx == lastIdx) {
					break
				}
			}
		default:
			panic("unsupported input type")
		}
	}, nil
}

func (t transformator[T, I]) AsMap() (map[any][]any, error) {
	r, err := t.AsIndexedRange()
	if err != nil {
		return nil, err
	}

	var acc any
	for _, v := range r {
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
	return res, nil
}
