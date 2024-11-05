package chans

import (
	"github.com/domahidizoltan/go-steps/internal/pkg/step"
	"github.com/domahidizoltan/go-steps/types"
)

type AnyStep any

type Step[U, V comparable] types.Step[U, V]

func Map[U, V comparable](fn func(in U) (V, error)) Step[U, V] {
	return Step[U, V](step.Map(fn))
}

func Filter[U comparable](fn func(in U) (bool, error)) Step[U, U] {
	return Step[U, U](step.Filter(fn))
}

type BinaryInputStep[U, V, W comparable] types.BinaryInputStep[U, V, W]

func MultiplyBy[U ~int](multiplier U) BinaryInputStep[U, U, U] {
	return BinaryInputStep[U, U, U](step.MultiplyBy(multiplier))
}
