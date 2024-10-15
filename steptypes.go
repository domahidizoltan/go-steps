package steps

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

var (
	stepsTypeCache = map[string][]reflect.Type{}

	pattern = regexp.MustCompile("")

	ErrNotAStep          = errors.New("Not a Step function")
	ErrInvalidInputType  = errors.New("Invalid input type")
	ErrInvalidOutputType = errors.New("Invalid output type")
)

func Validate[U comparable](steps ...[]any) error {
	for idx, step := range steps {
		t := reflect.TypeOf(step)
		if pattern.MatchString(t.Name()) {
			return fmt.Errorf("%w: %s (at position %d)", ErrNotAStep, t.Name(), idx)
		}

		// check prev and next and add to cache
		if idx == 0 {
			// steps[0]
			// in param type == U
		}
	}
	return nil
}
