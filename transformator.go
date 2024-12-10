package steps

import (
	"github.com/domahidizoltan/go-steps/internal/pkg/step"
)

type (
	inputType[T any] interface {
		chan T | []T
	}

	input[T any, I inputType[T]] struct {
		data I
	}

	stepsContainer step.StepsContainer

	transformator[T any, I inputType[T]] struct {
		in I
		step.Transformator
	}
)

func Transform[T any, I inputType[T]](in I) input[T, I] {
	return input[T, I]{in}
}

func Steps(s ...step.StepWrapper) step.StepsContainer {
	return step.Steps(s...)
}

func (s *stepsContainer) Validate() error {
	if s.Error != nil {
		return s.Error
	}

	s.Steps, s.Error = step.GetValidatedSteps[stepsContainer](s.StepWrappers)
	return s.Error
}

// TODO .WithOptions to add debug options, transformator name, etc
//.

// func (i input[T]) WithSteps(steps ...types.StepWrapper) transformator[T] {
// 	// validate if input T matches first step input type
// 	return i.With(Steps(steps...))
// }

func (i input[T, I]) With(steps step.StepsContainer) transformator[T, I] {
	// validate if input T matches first step input type

	_steps := stepsContainer{
		StepWrappers: steps.StepWrappers,
		Aggregator:   steps.Aggregator,
	}
	t := transformator[T, I]{
		Transformator: step.Transformator{
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
