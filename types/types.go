package types

type (
	Step[U, V comparable]               func(in U) (V, bool, error)
	BinaryInputStep[U, V, W comparable] func(in1 U, in2 V) (W, bool, error)
)
