package step

import (
	"errors"

	"github.com/domahidizoltan/go-steps/types"
)

type (
	Transformator struct {
		Error error
		Steps []types.StepFn
	}
)

var ErrInvalidInputType = errors.New("Invalid input type")
