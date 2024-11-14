package types

import "reflect"

const MaxArgs = 4

type (
	StepInput struct {
		Args    [MaxArgs]any
		ArgsLen uint8
	}

	StepOutput struct {
		Error   error
		Args    [MaxArgs]any
		ArgsLen uint8
		Skip    bool
	}

	StepWrapper struct {
		InTypes  [MaxArgs]reflect.Type
		OutTypes [MaxArgs]reflect.Type
		StepFn   StepFn
	}

	StepFn func(StepInput) StepOutput
)
