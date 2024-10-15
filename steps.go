package steps

type Step[U, V comparable] func(in U) (V, bool, error)

// TODO add partial Step[U ,V ,W comparable]
func Map[U, V comparable](fn func(in U) (V, error)) Step[U, V] {
	return func(in U) (V, bool, error) {
		out, err := fn(in)
		return out, err != nil, err
	}
}

func Filter[U comparable](fn func(in U) (bool, error)) Step[U, U] {
	return func(in U) (U, bool, error) {
		ok, err := fn(in)
		return in, ok, err
	}
}

// TODO Added only for testing purposes
type BinaryInputStep[U, V, W comparable] func(in1 U, in2 V) (W, bool, error)

func MultiplyBy[U ~int](multiplier U) BinaryInputStep[U, U, U] {
	return func(in U, _ U) (U, bool, error) {
		return in * multiplier, false, nil
	}
}
