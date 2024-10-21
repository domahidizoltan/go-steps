package step

import (
	"errors"
	"reflect"
)

type (
	Transformator struct {
		ID    string
		Err   error
		Steps []any
	}

	Data struct {
		Type reflect.Type
		Val  reflect.Value
	}
)

var (
	ErrNotAStep          = errors.New("Not a Step function")
	ErrInvalidInputType  = errors.New("Invalid input type")
	ErrInvalidOutputType = errors.New("Invalid output type")
)
