package steps

import (
	"iter"

	is "github.com/domahidizoltan/go-steps/internal/pkg/step"
)

func (t transformator[T, I]) AsRange() (iter.Seq[any], error) {
	if t.Error != nil {
		return nil, t.Error
	}

	return func(yield func(any) bool) {
		switch in := any(t.in).(type) {
		case chan T:
			for v := range in {
				if is.Process(v, yield, t.Transformator) {
					break
				}
			}
		case []T:
			for _, v := range in {
				if is.Process(v, yield, t.Transformator) {
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
				if is.ProcessIndexed(idx, v, yield, t.Transformator) {
					break
				}
			}
		case []T:
			for idx, v := range in {
				if is.ProcessIndexed(idx, v, yield, t.Transformator) {
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

	acc := map[any][]any{}
	for k, v := range r {
		acc[k] = append(acc[k], v)
	}
	return acc, nil
}
