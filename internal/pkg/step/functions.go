package step

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var (
	stepPattern            = regexp.MustCompile(`^Step\[(.*),(.*)]$`)
	binaryInputStepPattern = regexp.MustCompile(`^BinaryInputStep\[(.*),(.*),(.*)]$`)
)

func ValidateStepsNew[T any](steps []T) ([]FnType, error) {
	if len(steps) == 0 {
		return nil, nil
	}

	fnTypes := make([]FnType, 0, len(steps))
	transformatorTypePkg := reflect.TypeFor[T]().PkgPath()
	prevOutType := reflect.TypeOf(steps[0]).In(0)
	for pos, s := range steps {
		stepType := reflect.TypeOf(s)

		if stepType.PkgPath() != transformatorTypePkg {
			return nil, fmt.Errorf("%w: [pos %d.] %s", ErrInvalidStepType, pos, stepType.Name())
		}

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
			return nil, err
		}

		stepFn := reflect.ValueOf(s).Convert(fnType)
		fnTypes = append(fnTypes, FnType{
			Type: fnType,
			Val:  stepFn,
		})

		prevOutType = out0
	}
	return fnTypes, nil
}

// TODO deprecated
func ValidateSteps[T any](t *Transformator) {
	transformatorTypePkg := t.StepsType.PkgPath()
	prevOutType := reflect.TypeFor[T]()
	for pos, s := range t.Steps {
		stepType := reflect.TypeOf(s)

		if stepType.PkgPath() != transformatorTypePkg {
			delete(FnCache, t.ID)
			t.Err = fmt.Errorf("%w: [pos %d.] %s", ErrInvalidStepType, pos, stepType.Name())
			return
		}

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
			delete(FnCache, t.ID)
			t.Err = err
			return
		}

		stepFn := reflect.ValueOf(s).Convert(fnType)
		if pos == 0 {
			FnCache[t.ID] = []FnType{} // TODO make
		}
		FnCache[t.ID] = append(FnCache[t.ID], FnType{
			Type: fnType,
			Val:  stepFn,
		})

		t.Steps = append(t.Steps, s)
		prevOutType = out0
	}
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

func Process[T any](i T, yield func(i any) bool, fns []FnType) bool {
	var in any = i
	var skipped bool
	for _, fn := range fns {
		args := []reflect.Value{reflect.ValueOf(in)}

		if fn.Type.NumIn() > 1 {
			args = append(args, reflect.ValueOf(-1))
		}
		res := fn.Val.Call(args)
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

func CreateCacheID() string {
	for {
		id := randomString(8)
		if _, ok := FnCache[id]; !ok {
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

func ToAnySlice[T any](steps []T) []any {
	anySteps := make([]any, 0, len(steps))
	for _, s := range steps {
		anySteps = append(anySteps, s)
	}
	return anySteps
}
