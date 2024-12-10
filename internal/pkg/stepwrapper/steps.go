package stepwrapper

import (
	"reflect"

	"github.com/domahidizoltan/go-steps/internal/pkg/step"
)

func Map[IN0, OUT0 any](fn func(in IN0) (OUT0, error)) step.StepWrapper {
	return step.StepWrapper{
		InTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[OUT0](),
		},
		StepFn: func(in step.StepInput) step.StepOutput {
			out, err := fn(in.Args[0].(IN0))
			return step.StepOutput{
				Args:    [step.MaxArgs]any{out},
				ArgsLen: 1,
				Error:   err,
				Skip:    err != nil,
			}
		},
	}
}

func Filter[IN0 any](fn func(in IN0) (bool, error)) step.StepWrapper {
	return step.StepWrapper{
		InTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		StepFn: func(in step.StepInput) step.StepOutput {
			ok, err := fn(in.Args[0].(IN0))
			return step.StepOutput{
				Args:    [step.MaxArgs]any{in.Args[0].(IN0)},
				ArgsLen: 1,
				Error:   err,
				Skip:    !ok,
			}
		},
	}
}

type kv struct {
	value any
	t     reflect.Type
	key   uint8
}

func Split[IN0 any, OUT0 ~uint8](fn func(in IN0) (OUT0, error)) step.StepWrapper {
	return step.StepWrapper{
		InTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[kv](),
		},
		StepFn: func(in step.StepInput) step.StepOutput {
			idx, err := fn(in.Args[0].(IN0))
			out := kv{key: uint8(idx), value: in.Args[0].(IN0), t: reflect.TypeOf(in.Args[0])}
			return step.StepOutput{
				Args:    [step.MaxArgs]any{out},
				ArgsLen: 1,
				Error:   err,
			}
		},
	}
}

func WithBranches[IN0 any](steps ...step.StepsContainer) step.StepWrapper {
	return step.StepWrapper{
		InTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[kv](),
		},
		OutTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[kv](),
		},
		StepFn: func(in step.StepInput) step.StepOutput {
			keyVal := in.Args[0].(kv)
			val := step.StepInput{
				Args:    [step.MaxArgs]any{keyVal.value},
				ArgsLen: 1,
			}
			var out step.StepOutput
			for _, stepWrapper := range steps[int(keyVal.key)].StepWrappers {
				out = stepWrapper.StepFn(val)
				val = step.StepInput{
					Args:    out.Args,
					ArgsLen: out.ArgsLen,
				}
				if out.Skip || out.Error != nil {
					break
				}
			}
			return step.StepOutput{
				Args:    [step.MaxArgs]any{kv{key: keyVal.key, value: out.Args[0], t: reflect.TypeOf(out.Args[0])}},
				ArgsLen: 1,
				Error:   out.Error,
				Skip:    out.Skip,
			}
		},
	}
}

func Zip() step.StepWrapper {
	return step.StepWrapper{
		InTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[kv](),
		},
		OutTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[any](), // should not be any
		},
		StepFn: func(in step.StepInput) step.StepOutput {
			return step.StepOutput{
				Args:    [step.MaxArgs]any{in.Args[0].(kv).value},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
	}
}

// TODO Added only for testing purposes

func MultiplyBy[IN0 ~int](multiplier IN0) step.StepWrapper {
	return step.StepWrapper{
		InTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		OutTypes: [step.MaxArgs]reflect.Type{
			reflect.TypeFor[IN0](),
		},
		StepFn: func(in step.StepInput) step.StepOutput {
			return step.StepOutput{
				Args:    [step.MaxArgs]any{in.Args[0].(IN0) * multiplier},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
	}
}
