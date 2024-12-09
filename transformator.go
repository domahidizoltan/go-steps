package steps

import (
	is "github.com/domahidizoltan/go-steps/internal/pkg/step"
	"github.com/domahidizoltan/go-steps/types"
)

type (
	inputType[T any] interface {
		chan T | []T
	}

	input[T any, I inputType[T]] struct {
		data I
	}

	tempSteps is.TempSteps

	transformator[T any, I inputType[T]] struct {
		in I
		is.Transformator
	}
)

func Transform[T any, I inputType[T]](in I) input[T, I] {
	return input[T, I]{in}
}

func Steps(s ...types.StepWrapper) is.TempSteps {
	return is.Steps(s...)
}

func (s *tempSteps) Validate() error {
	if s.Error != nil {
		return s.Error
	}

	s.Steps, s.Error = is.GetValidatedSteps[tempSteps](s.StepWrappers)
	return s.Error
}

// TODO .WithOptions to add debug options, transformator name, etc
//.

// func (i input[T]) WithSteps(steps ...types.StepWrapper) transformator[T] {
// 	// validate if input T matches first step input type
// 	return i.With(Steps(steps...))
// }

func (i input[T, I]) With(steps is.TempSteps) transformator[T, I] {
	// validate if input T matches first step input type

	_steps := tempSteps{
		StepWrappers: steps.StepWrappers,
		Aggregator:   steps.Aggregator,
	}
	t := transformator[T, I]{
		Transformator: is.Transformator{
			Error: _steps.Validate(),
		},
	}
	if t.Error != nil {
		return t
	}

	// TODO in type must match first step first input type
	t.in = i.data
	t.Steps = _steps.Steps
	t.Aggregator = _steps.Aggregator
	return t
}
