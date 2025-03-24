package test

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"testing"

	. "github.com/domahidizoltan/go-steps"
	"github.com/jszwec/csvutil"
	"github.com/samber/lo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkNativeSimpleStep(b *testing.B) {
	b.StopTimer()

	var res []int
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		res = res[:0]
		for _, v := range input {
			if v%2 != 0 {
				continue
			}
			res = append(res, v)
		}
	}
	assert.Equal(b, []int{2, 4, 6, 8, 10}, res)
}

func BenchmarkLoSimpleStep(b *testing.B) {
	b.StopTimer()

	var res []int
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		res = res[:0]
		res = lo.Filter(input, func(v int, _ int) bool {
			return v%2 == 0
		})
	}

	assert.Equal(b, []int{2, 4, 6, 8, 10}, res)
}

func BenchmarkTransformerSimpleStep(b *testing.B) {
	b.StopTimer()

	steps := Steps(Filter(func(i int) (bool, error) {
		return i%2 == 0, nil
	}))
	require.NoError(b, steps.Validate())

	var res []any
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		res = res[:0]
		iter := Transform[int](input).
			With(steps).
			AsRange()

		for i := range iter {
			res = append(res, i)
		}
	}
	assert.Equal(b, []any{2, 4, 6, 8, 10}, res)
}

func BenchmarkNativeMultipleSteps(b *testing.B) {
	b.StopTimer()

	res := 0
	input := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for idx, val := range input {
			if idx < 3 {
				continue
			}

			num, err := strconv.Atoi(val)
			if err != nil {
				require.NoError(b, err)
			}

			if num%2 != 0 {
				continue
			}

			sum += num
		}
		res = sum
	}
	assert.Equal(b, 28, res)
}

func BenchmarkLoMultipleSteps(b *testing.B) {
	b.StopTimer()

	var res int
	input := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		skippedInput := input[3:]
		mappedInput := lo.Map(skippedInput, func(item string, _ int) int {
			num, _ := strconv.Atoi(item)
			return num
		})
		filteredInput := lo.Filter(mappedInput, func(v int, _ int) bool {
			return v%2 == 0
		})
		res = lo.Sum(filteredInput)
	}
	assert.Equal(b, 28, res)
}

func BenchmarkTransformerMultipleSteps(b *testing.B) {
	b.StopTimer()

	steps := Steps(
		Skip[string](3),
		Map(func(i string) (int, error) {
			return strconv.Atoi(i)
		}),
		Filter(func(i int) (bool, error) {
			return i%2 == 0, nil
		}),
	).Aggregate(Sum[int]())
	require.NoError(b, steps.Validate())

	var res []any
	input := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		res = res[:0]
		iter := Transform[string](input).
			With(steps).
			AsRange()
		for i := range iter {
			res = append(res, i)
		}
	}
	assert.Equal(b, 28, res[0])
}

const expectedJson = "[{\"id\":1,\"name\":\"John Doe\",\"age\":25,\"department\":\"Engineering\",\"salary\":60000,\"city\":\"New York\"},{\"id\":21,\"name\":\"Mark King\",\"age\":36,\"department\":\"Engineering\",\"salary\":78000,\"city\":\"New York\"}]"

func BenchmarkNativeCsvToJsonSteps(b *testing.B) {
	b.StopTimer()

	type salary struct {
		ID         int    `csv:"ID" json:"id"`
		Name       string `csv:"Name" json:"name"`
		Age        int    `csv:"Age" json:"age"`
		Department string `csv:"Department" json:"department"`
		Salary     int    `csv:"Salary" json:"salary"`
		City       string `csv:"City" json:"city"`
	}

	reader, err := os.Open("testdata/salaries.csv")
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		panic(err)
	}

	var rows []salary
	var transformRes []salary
	res := ""

	b.ReportAllocs()
	b.StartTimer()
	if err := csvutil.Unmarshal(data, &rows); err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		count := 0
		transformRes = transformRes[:0]
		for _, s := range rows {
			if !(s.City == "New York" && s.Department == "Engineering") {
				continue
			}

			if count >= 2 {
				break
			}
			count++

			transformRes = append(transformRes, s)
		}

		resBytes, err := json.Marshal(transformRes)
		if err != nil {
			panic(err)
		}
		res = string(resBytes)
	}

	assert.Equal(b, expectedJson, res)
}

func BenchmarkLoCsvToJsonSteps(b *testing.B) {
	b.StopTimer()

	type salary struct {
		ID         int    `csv:"ID" json:"id"`
		Name       string `csv:"Name" json:"name"`
		Age        int    `csv:"Age" json:"age"`
		Department string `csv:"Department" json:"department"`
		Salary     int    `csv:"Salary" json:"salary"`
		City       string `csv:"City" json:"city"`
	}

	reader, err := os.Open("testdata/salaries.csv")
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		panic(err)
	}

	var rows []salary
	var transformRes []salary
	res := ""

	b.ReportAllocs()
	b.StartTimer()
	if err := csvutil.Unmarshal(data, &rows); err != nil {
		panic(err)
	}

	for i := 0; i < b.N; i++ {
		transformRes = transformRes[:0]

		filteredRows := lo.Filter(rows, func(item salary, _ int) bool {
			return item.City == "New York" && item.Department == "Engineering"
		})

		transformRes = append(transformRes, filteredRows[:2]...)

		resBytes, err := json.Marshal(transformRes)
		if err != nil {
			panic(err)
		}
		res = string(resBytes)
	}

	assert.Equal(b, expectedJson, res)
}

func BenchmarkTransformerCsvToJsonSteps(b *testing.B) {
	b.StopTimer()

	type salary struct {
		ID         int    `csv:"ID" json:"id"`
		Name       string `csv:"Name" json:"name"`
		Age        int    `csv:"Age" json:"age"`
		Department string `csv:"Department" json:"department"`
		Salary     int    `csv:"Salary" json:"salary"`
		City       string `csv:"City" json:"city"`
	}

	steps := Steps(
		Filter(func(s salary) (bool, error) {
			return s.City == "New York" && s.Department == "Engineering", nil
		}),
		Take[salary](2),
	)
	require.NoError(b, steps.Validate())

	res := ""
	input := FromCsv[salary](File("testdata/salaries.csv"))

	b.ReportAllocs()
	b.StartTimer()

	transform := TransformFn[salary](input)
	for i := 0; i < b.N; i++ {
		res = transform.
			With(steps).
			AsJson()
	}

	assert.Equal(b, expectedJson, res)
}
