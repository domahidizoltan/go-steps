# go-steps

**go-steps** is an experimental data transformation library with one by one processing. 

This means that the processing steps are executed sequentially for each input items, 
and the processing steps are ignored once any step defines this, 
so the transformation jumps to the processing of the next item.

```go
iter := Transform[int]([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
	With(Steps(
		Take[int](7), // 1, 2, 3, 4, 5, 6, 7
		Filter(func(i int) (bool, error) {
			return i%2 == 0, nil
		}), // 2, 4, 6
		Map(func(i int) (string, error) {
			return strconv.Itoa(i * i), nil
		}), // "4", "16", "36"
	)).
	AsRange()

for i := range iter {
	// process
}
```

As you can see a transformation chain consist of 3 main parts:
- input: passed into `Transform`
- steps: this defines the transformation steps for each input items (`With(Steps(...))`)
- output: the result of the transformation (`AsRange`)

<br/>

**Input** can be a slice or a channel, enabling processing of streaming data. 
```go
inputCh := make(chan int, 3)
iter := Transform[int](inputCh).
	With(Steps(
		Take[int](7), // 1, 2, 3, 4, 5, 6, 7
		Filter(func(i int) (bool, error) {
			return i%2 == 0, nil
		}), // 2, 4, 6
		Map(func(i int) (string, error) {
			return strconv.Itoa(i * i), nil
		}), // "4", "16", "36"
	)).
	AsRange()

go func() {
	for _, i := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
		inputCh <- i
	}
	close(inputCh)
}()

for i := range iter {
	// process
}
```

Besides the input values you can also set some options for the transformer.
```go
buf := bytes.NewBufferString("")

res := Transform[int]([]int{1, 2, 3}, WithLogWriter(buf)).
	WithSteps(
		Log("before filter"),
		Filter(func(i int) (bool, error) {
			return i%2 == 1, nil
		}),
		Log("after filter"),
	).AsSlice()

fmt.Println(res) // [1 3]
fmt.Prtinln(buf.String())
/*
before filter 	arg0: 1 
after filter 	arg0: 1 
before filter 	arg0: 2 
before filter 	arg0: 3 
after filter 	arg0: 3 
*/
	
```

<br/>

**Steps** are using types to help the developer always see the types of input and output values.

Some type information could be lost during the processing, and because of this a validation process 
is running for the first time, to help fail fast the processing. 

Since the processing steps could be detached from the transformer, this validation could be ran separate, 
before starting the processing.
```go
steps := Steps(
	Filter(func(i int) (bool, error) {
		return i%2 == 1, nil
	}),
	Map(func(in string) (int, error) {
		return strconv.Atoi(in)
	}),
)

if err := steps.Validate(); err != nil {
	panic(err) //step validation failed [Map:2]: incompatible input argument type [int!=string:1]  
}

_ := Transform[int]([]int{1, 2, 3}).
	With(steps).
	AsSlice()
```

Errors of the steps are handled inside the processing, but they are not instantly returned, 
to not make the output more verbose. By default errors are logged, but the error handler is configurable.

When error occurs the processing stops, and the developer can decide how to propagate the error.
```go
var transformErr error
errHandler := func(err error) {
	transformErr = err
}

iter := Transform[int]([]int{10, 0, -10}, WithErrorHandler(errHandler)).
	WithSteps(
		Map(func(i int) (int, error) {
			if i == 0 {
				return 0, fmt.Errorf("division by zero: input=%d", i)
			}
			return i / 10, nil
		}),
	).AsRange()

res := []int{}
for i := range iter {
	res = append(res, i.(int))
}
fmt.Println(res) // [1]
	
if transformErr != nil {
	fmt.Println(transformErr.Error()) //division by zero: input=0
}

```

Steps can also have an extra final step used to aggregate the results. 

There are no more processing steps allowed beyond this point.
```go
type (
	person struct {
		name   string
		age    int
		isMale bool
	}
	ageRange uint8
)

const (
	young ageRange = iota
	mature
	old
)

persons := []person{
	{"John", 30, true},
	{"Jane", 25, false},
	{"Bob", 75, true},
	{"Alice", 28, false},
	{"Charlie", 17, true},
	{"Frank", 81, true},
	{"Bill", 45, true},
}

res := Transform[person](persons).
	With(Steps(
		Filter(func(p person) (bool, error) {
			return p.isMale, nil
		}),
	).Aggregate(
		GroupBy(func(p person) (ageRange, string, error) {
			switch {
			case p.age <= 18:
				return young, p.name, nil
			case p.age <= 65:
				return mature, p.name, nil
			default:
				return old, p.name, nil
			}
		}),
	),
	).
	AsMultiMap()

fmt.Println(res) //map[0:[Charlie] 1:[John Bill] 2:[Bob Frank]]
```

<br/>

**Output** is the result of the transformation. It can return an iterator (`AsRange`) 
where the resulting items are passed one by one to the `range` keyword, but they could also be 
collected and returned as a single result (`AsSlice` or `AsMap`)
```go 
inputCh := make(chan int, 3)
go func() {
	for _, i := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
		inputCh <- i
	}
	close(inputCh)
}()

res := Transform[int](inputCh).
	With(Steps(
		Take[int](7), // 1, 2, 3, 4, 5, 6, 7
		Filter(func(i int) (bool, error) {
			return i%2 == 0, nil
		}), // 2, 4, 6
		Map(func(i int) (string, error) {
			return strconv.Itoa(i * i), nil
		}), // "4", "16", "36"
	)).
	AsMap()

fmt.Println(res) //map[1:4 3:16 5:36]

```

Custom inputs and steps could be also defined:
```go
func multiplyBy[IN0 ~int](multiplier IN0) StepWrapper {
	return StepWrapper{
		Name: "multiplyBy",
		StepFn: func(in StepInput) StepOutput {
			return StepOutput{
				Args:    Args{in.Args[0].(IN0) * multiplier},
				ArgsLen: 1,
				Error:   nil,
				Skip:    false,
			}
		},
		Validate: func(prevStepOut ArgTypes) (ArgTypes, error) {
			return ArgTypes{reflect.TypeFor[IN0]()}, nil
		},
	}
}

func primeInput(opts TransformerOptions) chan int {
	inputCh := make(chan int, 3)
	go func() {
		for _, i := range []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29} {
			inputCh <- i
		}
		close(inputCh)
	}()
	return inputCh
}


func main() {
    res := TransformFn[int](primeInput).
		WithSteps(multiplyBy(3)).
		AsSlice()
	fmt.Println(res) //[6 9 15 21 33 39 51 57 69 87]
}
```

It is also possible to split up a transformation chain into branches, but they must be merged back 
before returning the output.
```go
res := Transform[int]([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
	WithSteps(
		Split(func(in int) (uint8, error) {
			return uint8(in % 2), nil
		}),
		WithBranches[int](
			Steps(
				Map(func(in int) (string, error) {
					return strconv.Itoa(-in), nil
				}),
			),
			Steps(
    			Map(func(in int) (string, error) {
					return strconv.Itoa(in * 10), nil
				}),
			),
		),
		Merge(),
	).
	AsSlice()

fmt.Printf("%#v", res) //[]interface {}{"10", "-2", "30", "-4", "50", "-6", "70", "-8", "90", "-10"}
```
The branch steps are also validated, and could be explicitly validated as well.

**go-steps** supports CSV (using https://github.com/jszwec/csvutil) and JSON input and output.
```go
type salary struct {
	ID         string `csv:"ID" json:"id"`
	Name       string `csv:"Name" json:"name"`
	Age        int    `csv:"Age" json:"age"`
	Department string `csv:"Department" json:"department"`
	Salary     int    `csv:"Salary" json:"salary"`
	City       string `csv:"City" json:"city"`
}

res := TransformFn[salary](FromStreamingCsv[salary](File("testdata/salaries.csv"), false)).
	WithSteps(
		Filter(func(s salary) (bool, error) {
			return s.City == "New York" && s.Department == "Engineering", nil
		}),
		Take[salary](2),
	).
	AsJson()

fmt.Println(res) //[{"id":1,"name":"John Doe","age":25,"department":"Engineering","salary":60000,"city":"New York"},{"lary":78000,"city":"New York"}]
```


See the [samples](https://github.com/domahidizoltan/go-steps/blob/master/test/samples_test.go) for more details

For documentation open the godoc with `m=all` option:
```bash
godoc -http=:6060
# http://localhost:6060/pkg/github.com/domahidizoltan/go-steps?m=all
```

## Benchmarks
Since the library is using a lot more steps and reflection, it's obvious that it is slower than the native Go version.  

As more complex the transformation chain gets, the difference gets bigger.  

Here are some benchmarks for reference:  
- SimpleStep: using a simple step on a slice
- MultipleSteps: using multiple three steps and an aggregation on a slice
- CsvToJsonSteps: reading a CSV file and converting it to JSON while doing some minimal transformation on the data

```bash
goos: linux
goarch: amd64
cpu: AMD Ryzen 7 7840HS with Radeon 780M Graphics
                  │ tmp/native.txt │               tmp/lo.txt               │           tmp/transformer.txt           │
                  │     sec/op     │    sec/op      vs base                 │    sec/op      vs base                  │
SimpleStep-16          6.544n ± 2%   44.820n ± 19%  +584.90% (p=0.000 n=10)   534.750n ± 3%  +8071.61% (p=0.000 n=10)
MultipleSteps-16       21.79n ± 0%    88.43n ±  1%  +305.83% (p=0.000 n=10)   1392.00n ± 1%  +6288.25% (p=0.000 n=10)
CsvToJsonSteps-16      522.6n ± 3%   2891.0n ± 20%  +453.20% (p=0.000 n=10)   13404.5n ± 1%  +2464.96% (p=0.000 n=10)
geomean                42.08n         225.4n        +435.73%                    2.153µ       +5015.92%

                  │ tmp/native.txt │              tmp/lo.txt               │          tmp/transformer.txt           │
                  │      B/op      │    B/op      vs base                  │     B/op      vs base                  │
SimpleStep-16          0.00 ± 0%      80.00 ± 0%          ? (p=0.000 n=10)    104.00 ± 0%          ? (p=0.000 n=10)
MultipleSteps-16        0.0 ± 0%      128.0 ± 0%          ? (p=0.000 n=10)     464.0 ± 0%          ? (p=0.000 n=10)
CsvToJsonSteps-16     440.0 ± 0%     8639.0 ± 0%  +1863.41% (p=0.000 n=10)   16856.0 ± 0%  +3730.91% (p=0.000 n=10)
geomean                          ¹    445.6       ?                            933.5       ?
¹ summaries must be >0 to compute geomean

                  │ tmp/native.txt │             tmp/lo.txt             │          tmp/transformer.txt           │
                  │   allocs/op    │ allocs/op   vs base                │  allocs/op    vs base                  │
SimpleStep-16         0.000 ± 0%     1.000 ± 0%        ? (p=0.000 n=10)     3.000 ± 0%          ? (p=0.000 n=10)
MultipleSteps-16      0.000 ± 0%     2.000 ± 0%        ? (p=0.000 n=10)    25.000 ± 0%          ? (p=0.000 n=10)
CsvToJsonSteps-16     3.000 ± 0%     4.000 ± 0%  +33.33% (p=0.000 n=10)   211.000 ± 0%  +6933.33% (p=0.000 n=10)
geomean                          ¹   2.000       ?                          25.11       ?
¹ summaries must be >0 to compute geomean
```
See the [benchmarks](https://github.com/domahidizoltan/go-steps/blob/master/test/benchmarks_test.go) for more details

