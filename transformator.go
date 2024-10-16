package steps

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type transformatorInput[U comparable] struct {
	inCh chan U
	in   []U
}

type Transformator[U comparable] struct {
	err   error
	id    string
	inCh  chan U
	in    []U
	steps []any
}

func Transform[U comparable](in []U) transformatorInput[U] {
	return transformatorInput[U]{
		in: in,
	}
}

func TransformChan[U comparable](inCh chan U) transformatorInput[U] {
	return transformatorInput[U]{
		inCh: inCh,
	}
}

var (
	stepPattern            = regexp.MustCompile(`^Step\[(.*),(.*)]$`)
	binaryInputStepPattern = regexp.MustCompile(`^BinaryInputStep\[(.*),(.*),(.*)]$`)
	stepFnCache            = map[string][]reflect.Value{}
)

func (p transformatorInput[U]) With(steps ...any) Transformator[U] {
	t := Transformator[U]{
		id:    createCacheID(),
		in:    p.in,
		inCh:  p.inCh,
		steps: []any{},
	}

	prevOutType := reflect.TypeFor[U]()
	for pos, step := range steps {
		stepType := reflect.TypeOf(step)

		var fnType reflect.Type
		var out0 reflect.Type
		var err error

		switch {
		case strings.HasPrefix(stepType.Name(), "BinaryInputStep"):
			fnType, out0, err = binaryInputStepParser(pos, prevOutType, stepType)
		default:
			fnType, out0, err = stepParser(pos, prevOutType, stepType)
		}

		if err != nil {
			delete(stepFnCache, t.id)
			t.err = err
			return t
		}

		stepFn := reflect.ValueOf(step).Convert(fnType)

		if pos == 0 {
			stepFnCache[t.id] = []reflect.Value{}
		}
		stepFnCache[t.id] = append(stepFnCache[t.id], stepFn)

		t.steps = append(t.steps, step)
		prevOutType = out0
	}

	fmt.Println("done With")

	return t
}

func stepParser(pos int, prevOutType reflect.Type, stepType reflect.Type) (reflect.Type, reflect.Type, error) {
	if !stepPattern.MatchString(stepType.Name()) {
		return nil, nil, fmt.Errorf("%w: [pos %d.] %s", ErrNotAStep, pos, stepType.Name())
	}

	in0 := stepType.In(0)
	if in0 != prevOutType {
		if pos == 0 {
			return nil, nil, fmt.Errorf("%w: [pos %d.] input type was %s instead of %s from the generic type", ErrInvalidInputType, pos, in0.Name(), prevOutType.Name())
		}

		return nil, nil, fmt.Errorf("%w: [pos %d.] input type was %s instead of %s from the previous output", ErrInvalidInputType, pos, in0.Name(), prevOutType.Name())
	}

	out0 := stepType.Out(0)
	out1 := reflect.TypeFor[bool]()
	out2 := reflect.TypeFor[error]()

	fnType := reflect.FuncOf([]reflect.Type{in0}, []reflect.Type{out0, out1, out2}, false)
	return fnType, out0, nil
}

func binaryInputStepParser(pos int, prevOutType reflect.Type, stepType reflect.Type) (reflect.Type, reflect.Type, error) {
	if !binaryInputStepPattern.MatchString(stepType.Name()) {
		return nil, nil, fmt.Errorf("%w: [pos %d.] %s", ErrNotAStep, pos, stepType.Name())
	}

	in0 := stepType.In(0)
	in1 := stepType.In(1)
	if in0 != prevOutType {
		if pos == 0 {
			return nil, nil, fmt.Errorf("%w: [pos %d.] input type was %s instead of %s from the generic type", ErrInvalidInputType, pos, in0.Name(), prevOutType.Name())
		}

		return nil, nil, fmt.Errorf("%w: [pos %d.] input type was %s instead of %s from the previous output", ErrInvalidInputType, pos, in0.Name(), prevOutType.Name())
	}

	out0 := stepType.Out(0)
	out1 := reflect.TypeFor[bool]()
	out2 := reflect.TypeFor[error]()

	fnType := reflect.FuncOf([]reflect.Type{in0, in1}, []reflect.Type{out0, out1, out2}, false)
	return fnType, out0, nil
}

func (t Transformator[U]) AsRange() (func(yield func(i any) bool), error) {
	fmt.Println("AsRange")

	if t.err != nil {
		delete(stepFnCache, t.id)
		return nil, t.err
	}

	fns := stepFnCache[t.id]
	return func(yield func(i any) bool) {
		process := func(i U) bool {
			var in any = i
			var skipped bool
			for i, fn := range fns {
				args := []reflect.Value{reflect.ValueOf(in)}

				// TODO make stepFnCache a struct for extra arguments
				if i == 2 {
					args = append(args, reflect.ValueOf(-1))
				}
				res := fn.Call(args)
				out, skip, err := res[0].Interface(), res[1].Bool(), res[2].Interface()
				if err != nil {
					// TODO return error
				}

				if skip || err != nil {
					skipped = true
					break
				}

				in = out
			}

			if !skipped && !yield(in) {
				return true
			}

			return false
		}

		if t.inCh != nil {
			for i := range t.inCh {
				if process(i) {
					break
				}
			}
		} else {
			for _, i := range t.in {
				if process(i) {
					break
				}
			}
		}
	}, nil
}

func createCacheID() string {
	for {
		id := randomString(8)
		if _, ok := stepFnCache[id]; !ok {
			fmt.Println("id", id)
			return id
		}
	}
}

func randomString(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
