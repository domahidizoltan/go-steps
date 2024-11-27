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

// TODO can I proxy some of these functions as well?
func Steps(s ...types.StepWrapper) tempSteps {
	// fmt.Println("addSteps")
	return tempSteps{
		StepWrappers: s,
	}
}

func (t tempSteps) Aggregate(fn types.ReducerWrapper) tempSteps { //?
	return tempSteps{
		StepWrappers: t.StepWrappers,
		Aggregator:   fn.ReducerFn,
	}
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

func (i input[T, I]) With(steps tempSteps) transformator[T, I] {
	// validate if input T matches first step input type
	t := transformator[T, I]{
		Transformator: is.Transformator{
			Error: steps.Validate(),
		},
	}

	if t.Error != nil {
		return t
	}

	// TODO in type must match first step first input type
	t.in = i.data
	t.Steps = append(t.Steps, steps.Steps...)
	t.Aggregator = steps.Aggregator
	return t
}
