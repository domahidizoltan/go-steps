package test

import (
	"bytes"
	"encoding/csv"
	"os"
	"strconv"
	"testing"

	s "github.com/domahidizoltan/go-steps"
	"github.com/domahidizoltan/go-steps/test/customwrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformSliceAsRange(t *testing.T) {
	errorHandler := func(err error) {
		require.NoError(t, err)
	}

	transformator := s.Transform[string]([]string{"1", "2", "3", "4", "5"}).
		With(s.Steps(
			s.Map(func(i string) (int, error) {
				return strconv.Atoi(i)
			}),
			s.Filter(func(i int) (bool, error) {
				return i%2 == 0, nil
			}),
			customwrapper.MultiplyBy(3),
			s.Map(func(i int) (string, error) {
				return "_" + strconv.Itoa(i*2), nil
			}),
		))

	expected := []string{"_12", "_24"}
	actual := []string{}
	for i := range transformator.AsRange(errorHandler) {
		actual = append(actual, i.(string))
	}
	assert.Len(t, actual, 2)
	assert.Equal(t, expected, actual)
}

type employee struct {
	ID         int
	Name       string
	Age        int
	Department string
	Salary     int
	City       string
}

func loadCsv(filePath string) ([]employee, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(b))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var employees []employee
	for i, row := range rows {
		if i == 0 {
			continue
		}

		emp := employee{
			Name:       row[1],
			Department: row[3],
			City:       row[5],
		}

		var err error
		emp.ID, err = strconv.Atoi(row[0])
		if err != nil {
			return nil, err
		}
		emp.Age, err = strconv.Atoi(row[2])
		if err != nil {
			return nil, err
		}
		emp.Salary, err = strconv.Atoi(row[4])
		if err != nil {
			return nil, err
		}

		employees = append(employees, emp)
	}
	return employees, nil
}

func TestAggregateCsv(t *testing.T) {
	type ageRange uint8

	const (
		young ageRange = iota
		midAge
		mature
	)
	employees, err := loadCsv("testdata/salaries.csv")
	require.NoError(t, err)

	errorHandler := func(err error) {
		require.NoError(t, err)
	}
	transformator := s.Transform[employee](employees).
		With(s.Steps(
			s.Filter(func(e employee) (bool, error) {
				// if e.City == "New York" && e.Department == "Engineering" {
				// 	fmt.Printf("%+v\n", e)
				// }
				return e.City == "New York" && e.Department == "Engineering", nil
			}),
		).Aggregate(s.GroupBy(func(e employee) (ageRange, int, error) {
			switch {
			case e.Age <= 30:
				return young, e.Salary, nil
			case e.Age <= 45:
				return midAge, e.Salary, nil
			default:
				return mature, e.Salary, nil
			}
		})),
		)

	expected := [3]float64{68000, 74500, 0}
	sum := func(values []any) float64 {
		var sum float64
		for _, v := range values {
			sum += float64(v.(int))
		}
		return sum / float64(len(values))
	}

	actual := [3]float64{}
	for k, v := range transformator.AsMultiMap(errorHandler) {
		switch k.(ageRange) {
		case young:
			actual[0] = sum(v)
		case midAge:
			actual[1] = sum(v)
		case mature:
			actual[2] = sum(v)
		}
	}

	assert.Equal(t, expected, actual)
}

func TestSplitAndZip(t *testing.T) {
	type fizzBuzz uint8

	const (
		none fizzBuzz = iota
		fz
		bz
		fzbz
	)
	prefixWith := func(prefix string) s.StepWrapper {
		return s.Map(func(i int) (string, error) {
			return prefix + strconv.Itoa(i), nil
		})
	}

	errorHandler := func(err error) {
		require.NoError(t, err)
	}
	transformator := s.Transform[int]([]int{1, 2, 3, 4, 5, 6}).
		With(s.Steps(
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
				s.Steps(prefixWith("")),
				s.Steps(prefixWith("fz:")),
				s.Steps(prefixWith("bz:")),
				s.Steps(prefixWith("fzbz:")),
			),
			s.Merge(),
		))

	expected := []string{"1", "fz:2", "bz:3", "fz:4", "5", "fzbz:6"}
	actual := []string{}
	for i := range transformator.AsRange(errorHandler) {
		actual = append(actual, i.(string))
	}
	assert.Equal(t, expected, actual)
}
