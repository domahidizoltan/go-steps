package test

import (
	"reflect"
	"strconv"
	"testing"

	s "github.com/domahidizoltan/go-steps"
)

// op: 0.7ns 0B 0 alloc
func BenchmarkFn(b *testing.B) {
	b.ReportAllocs()

	f := func(a, b int) int {
		return a + b
	}

	var res int
	for i := 0; i < b.N; i++ {
		res = f(i, i+1)
	}
	_ = res
}

// op: 2.9ns 0B 0 alloc
func BenchmarkFnSlice(b *testing.B) {
	b.ReportAllocs()

	f := func(a []int) int {
		sum := 0
		for _, v := range a {
			sum += v
		}
		return sum
	}

	var res int
	for i := 0; i < b.N; i++ {
		res = f([]int{i, i + 1})
	}
	_ = res
}

// op: 443.0ns 47B 3 alloc
func BenchmarkReflect(b *testing.B) {
	b.ReportAllocs()

	def := func(a, b int) int {
		return a + b
	}

	in0 := reflect.TypeFor[int]()
	in1 := reflect.TypeFor[int]()
	out := reflect.TypeFor[int]()
	fn := reflect.FuncOf([]reflect.Type{in0, in1}, []reflect.Type{out}, false)

	f := reflect.ValueOf(def).Convert(fn)

	var res int
	for i := 0; i < b.N; i++ {
		r := f.Call([]reflect.Value{reflect.ValueOf(i), reflect.ValueOf(i + 1)})
		res = r[0].Interface().(int)
	}
	_ = res
}

// op: 443.0ns 72B 4 alloc
func BenchmarkReflectSlice(b *testing.B) {
	b.ReportAllocs()

	def := func(a [8]int) int {
		sum := 0
		for _, v := range a {
			sum += v
		}
		return sum
	}

	_ = def
	def2 := func(a int) int { return 0 }

	in0 := reflect.TypeFor[[8]int]()
	out := reflect.TypeFor[int]()
	fn := reflect.FuncOf([]reflect.Type{in0}, []reflect.Type{out}, false)

	_ = def2
	// _ = fn
	// in0 = reflect.TypeFor[int]()
	// fn2 := reflect.FuncOf([]reflect.Type{in0}, []reflect.Type{out}, false)
	f := reflect.ValueOf(def).Convert(fn)

	var res int
	for i := 0; i < b.N; i++ {
		// r := f.Call([]reflect.Value{reflect.ValueOf([]int{i, i + 1})})
		// res = r[0].Interface().(int)

		f.Call([]reflect.Value{reflect.ValueOf([8]int{i})})
	}
	_ = res
}

// op: 290ns 54B 4 alloc
func BenchmarkSimple(b *testing.B) {
	in := []string{"1", "2", "3", "4", "5"}
	b.ReportAllocs()
	x := []string{}
	for i := 0; i < b.N; i++ {
		res := []string{}
		for _, v := range in {
			a, _ := strconv.Atoi(v)
			if a%2 != 0 {
				continue
			}
			a = a * 3
			res = append(res, "_"+strconv.Itoa(a*2))
		}
		x = res

	}
	_ = x
}

// op: 22Kns 3169B 92 alloc
// with different id but prevalidated op: 15Kns 2392B 82 alloc
// with same ID op: 11Kns 2192B 79 alloc
//
// step2 op: 3450ns 264B 17 alloc
// step op: 1680ns 296B 14 alloc
func BenchmarkSlice(b *testing.B) {
	b.ReportAllocs()
	steps := s.Steps(
		s.Map(func(i string) (int, error) {
			return strconv.Atoi(i)
		}),
		s.Filter(func(i int) (bool, error) {
			return i%2 == 0, nil
		}),
		s.MultiplyBy(3),
		s.Map(func(i int) (string, error) {
			return "_" + strconv.Itoa(i*2), nil
		}),
	)
	// steps.Validate()
	b.ResetTimer()
	var x string
	for i := 0; i < b.N; i++ {
		r, err := s.Transform[string]([]string{"1", "2", "3", "4", "5"}).
			With(steps).AsRange()
		if err != nil {
			panic(err)
		}
		for i := range r {
			x = i.(string)
		}
	}
	_ = x
}

// op: 3480ns 800B 26 alloc
func BenchmarkWithBranches(b *testing.B) {
	b.ReportAllocs()
	type fizzBuzz uint8

	const (
		none fizzBuzz = iota
		fz
		bz
		fzbz
	)

	steps := s.Steps(
		s.Split(func(i int) (fizzBuzz, error) {
			switch {
			case i%6 == 0:
				return fzbz, nil
			case i%2 == 0:
				return fz, nil
			case i%3 == 0:
				return bz, nil
			}
			return none, nil
		}),
		s.WithBranches[int](
			s.Steps(s.Map(func(i int) (string, error) {
				return strconv.Itoa(i), nil
			})),
			s.Steps(s.Map(func(i int) (string, error) {
				return "fz:" + strconv.Itoa(i), nil
			})),
			s.Steps(s.Map(func(i int) (string, error) {
				return "bz:" + strconv.Itoa(i), nil
			})),
			s.Steps(s.Map(func(i int) (string, error) {
				return "fzbz:" + strconv.Itoa(i), nil
			})),
		),
		s.Zip(),
	)

	// steps.Validate()
	b.ResetTimer()
	var x string
	for i := 0; i < b.N; i++ {
		r, err := s.Transform[int]([]int{1, 2, 3, 4, 5, 6}).
			With(steps).AsRange()
		if err != nil {
			panic(err)
		}
		for i := range r {
			x = i.(string)
		}
	}
	_ = x
}

type testSt struct {
	err  error
	smg  *int
	args [4]any
	emit bool
}

func BenchmarkStruct(b *testing.B) {
	b.ReportAllocs()
	var res testSt
	x := 1
	for i := 0; i < b.N; i++ {
		st := testSt{
			args: [4]any{"1", "2", "3", "4"},
			emit: true,
			smg:  &x,
		}
		res = st
	}
	_ = res
}

// TODO refactor + validate
