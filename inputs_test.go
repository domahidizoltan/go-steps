package steps

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testPerson struct {
	Id   int        `csv:"-" json:"-"`
	Name string     `csv:"name" json:"name"`
	Dob  *time.Time `csv:"dob,omitempty" json:"dob,omitempty"`
	Code int        `csv:"code" json:"code"`
}

type failReader struct {
	err error
}

func (f failReader) Read([]byte) (int, error) {
	return 0, f.err
}

func TestFromCsv(t *testing.T) {
	for _, sc := range []struct {
		name        string
		input       io.Reader
		expected    []testPerson
		expectedErr string
	}{
		{
			name: "parse_csv",
			input: strings.NewReader(`
				id,name,dob,code
				1,John Doe,2006-01-02T15:04:05+07:00,11
				2,Jane Doe,1992-05-23T08:01:00Z,22
				3,Doe,,33`),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Jane Doe", Dob: mustParseTime("1992-05-23T08:01:00Z"), Code: 22},
				{Name: "Doe", Code: 33},
			},
		}, {
			name:        "reader_error",
			input:       &failReader{errors.New("reader error")},
			expectedErr: "reader error",
		}, {
			name: "parse_error",
			input: strings.NewReader(`
				id,name,dob,code
				1,John Doe,2006-01-02T15:04:05+07:00,11
				2,Jane Doe,1992-05-23T08:01:00Z,22
				xxxx`),
			expectedErr: "record on line 5: wrong number of fields",
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			opts := TransformerOptions{
				PanicHandler: func(err error) {
					assert.ErrorContains(t, err, sc.expectedErr)
				},
			}
			actual := FromCsv[testPerson](sc.input)(opts)
			assert.Equal(t, sc.expected, actual)
		})
	}
}

func TestStreamingCsv(t *testing.T) {
	for _, sc := range []struct {
		name           string
		input          io.Reader
		withoutHeaders bool
		expected       []testPerson
		expectedErr    string
		cancelled      bool
	}{
		{
			name: "parse_csv",
			input: bufio.NewReaderSize(strings.NewReader(`
				id,name,dob,code
				1,John Doe,2006-01-02T15:04:05+07:00,11
				2,Jane Doe,1992-05-23T08:01:00Z,22
				3,Doe,,33`), 16),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Jane Doe", Dob: mustParseTime("1992-05-23T08:01:00Z"), Code: 22},
				{Name: "Doe", Code: 33},
			},
		}, {
			name:           "parse_csv_without_headers",
			withoutHeaders: true,
			input: bufio.NewReaderSize(strings.NewReader(`John Doe,2006-01-02T15:04:05+07:00,11
Jane Doe,1992-05-23T08:01:00Z,22
Doe,,33`), 16),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Jane Doe", Dob: mustParseTime("1992-05-23T08:01:00Z"), Code: 22},
				{Name: "Doe", Code: 33},
			},
		}, {
			name:        "reader_error",
			input:       failReader{errors.New("reader error")},
			expected:    []testPerson{},
			expectedErr: "reader error",
		}, {
			name: "parse_error",
			input: strings.NewReader(`
					id,name,dob,code
					1,John Doe,2006-01-02T15:04:05+07:00,11
					xxxx,
					3,Doe,,33`),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Doe", Code: 33},
			},
			expectedErr: "record on line 4: wrong number of fields",
		}, {
			name: "canceled",
			input: strings.NewReader(`
					id,name,dob,code
					1,John Doe,2006-01-02T15:04:05+07:00,11
					2,Jane Doe,1992-05-23T08:01:00Z,22
					3,Doe,,33`),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Jane Doe", Dob: mustParseTime("1992-05-23T08:01:00Z"), Code: 22},
			},
			cancelled: true,
		},
	} {
		ctx, cancel := context.WithCancel(context.Background())
		t.Run(sc.name, func(t *testing.T) {
			opts := TransformerOptions{
				Ctx: ctx,
				PanicHandler: func(err error) {
					assert.ErrorContains(t, err, sc.expectedErr)
				},
				ErrorHandler: func(err error) {
					assert.ErrorContains(t, err, "context canceled")
				},
			}

			actual := []testPerson{}
			for res := range FromStreamingCsv[testPerson](sc.input, sc.withoutHeaders)(opts) {
				if sc.cancelled {
					cancel()
				}
				actual = append(actual, res)
			}
			assert.Equal(t, sc.expected, actual)
		})
	}
}

func TestFromJson(t *testing.T) {
	for _, sc := range []struct {
		name        string
		input       io.Reader
		expected    []testPerson
		expectedErr string
	}{
		{
			name: "parse_json",
			input: strings.NewReader(`
				[
					{"id":1,"name":"John Doe","dob":"2006-01-02T15:04:05+07:00","code":11},
					{"id":2,"name":"Jane Doe","dob":"1992-05-23T08:01:00Z","code":22},
					{"id":3,"name":"Doe","code":33}
				]`),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Jane Doe", Dob: mustParseTime("1992-05-23T08:01:00Z"), Code: 22},
				{Name: "Doe", Code: 33},
			},
		}, {
			name:        "reader_error",
			input:       &failReader{errors.New("reader error")},
			expectedErr: "reader error",
		}, {
			name:        "parse_error",
			input:       strings.NewReader(`[{"xxxx"}]`),
			expectedErr: "invalid character '}' after object key",
		},
	} {
		t.Run(sc.name, func(t *testing.T) {
			opts := TransformerOptions{
				PanicHandler: func(err error) {
					assert.ErrorContains(t, err, sc.expectedErr)
				},
			}
			actual := FromJson[testPerson](sc.input)(opts)
			assert.Equal(t, sc.expected, actual)
		})
	}
}

func TestStreamingJson(t *testing.T) {
	for _, sc := range []struct {
		name        string
		input       io.Reader
		expected    []testPerson
		expectedErr string
		cancelled   bool
	}{
		{
			name: "parse_json",
			input: bufio.NewReaderSize(strings.NewReader(`
					{"id":1,"name":"John Doe","dob":"2006-01-02T15:04:05+07:00","code":11}
					{"id":2,"name":"Jane Doe","dob":"1992-05-23T08:01:00Z","code":22}
					{"id":3,"name":"Doe","code":33}
				`), 16),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Jane Doe", Dob: mustParseTime("1992-05-23T08:01:00Z"), Code: 22},
				{Name: "Doe", Code: 33},
			},
		}, {
			name: "parse_json_in_brackets",
			input: bufio.NewReaderSize(strings.NewReader(`[
					{"id":1,"name":"John Doe","dob":"2006-01-02T15:04:05+07:00","code":11}
					{"id":2,"name":"Jane Doe","dob":"1992-05-23T08:01:00Z","code":22}
					{"id":3,"name":"Doe","code":33}
				]`), 16),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Jane Doe", Dob: mustParseTime("1992-05-23T08:01:00Z"), Code: 22},
				{Name: "Doe", Code: 33},
			},
		}, {
			name: "parse_error",
			input: strings.NewReader(`
					{"id":1,"name":"John Doe","dob":"2006-01-02T15:04:05+07:00","code":11}
					{"xxxx"}
					{"id":3,"name":"Doe","code":33}`),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Doe", Code: 33},
			},
			expectedErr: "invalid character '}' after object key",
		}, {
			name: "canceled",
			input: strings.NewReader(`
					{"id":1,"name":"John Doe","dob":"2006-01-02T15:04:05+07:00","code":11}
					{"id":2,"name":"Jane Doe","dob":"1992-05-23T08:01:00Z","code":22}
					{"id":3,"name":"Doe","code":33}`),
			expected: []testPerson{
				{Name: "John Doe", Dob: mustParseTime("2006-01-02T15:04:05+07:00"), Code: 11},
				{Name: "Jane Doe", Dob: mustParseTime("1992-05-23T08:01:00Z"), Code: 22},
			},
			cancelled: true,
		},
	} {
		ctx, cancel := context.WithCancel(context.Background())
		t.Run(sc.name, func(t *testing.T) {
			opts := TransformerOptions{
				Ctx: ctx,
				PanicHandler: func(err error) {
					assert.ErrorContains(t, err, sc.expectedErr)
				},
				ErrorHandler: func(err error) {
					assert.ErrorContains(t, err, "context canceled")
				},
			}

			actual := []testPerson{}
			for res := range FromStreamingJson[testPerson](sc.input)(opts) {
				if sc.cancelled {
					cancel()
				}
				actual = append(actual, res)
			}
			assert.Equal(t, sc.expected, actual)
		})
	}
}

func mustParseTime(s string) *time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &t
}

func ExampleFromCsv() {
	type person struct {
		ID   int    `csv:"id"`
		Name string `csv:"name"`
	}
	reader := strings.NewReader(`
		id,name
		1,John Doe
		2,Jane Doe`)

	res := TransformFn[person](FromCsv[person](reader)).
		WithSteps(
			Map(func(in person) (string, error) {
				return in.Name, nil
			}),
		).
		AsSlice()

	fmt.Println(res)
	// Output: [John Doe Jane Doe]
}

func ExampleFromStreamingCsv() {
	type person struct {
		ID   int    `csv:"id"`
		Name string `csv:"name"`
	}
	reader := strings.NewReader(`1,John Doe
2,Jane Doe`)

	res := TransformFn[person](FromStreamingCsv[person](reader, true)).
		WithSteps(
			Map(func(in person) (string, error) {
				return in.Name, nil
			}),
		).
		AsSlice()

	fmt.Println(res)
	// Output: [John Doe Jane Doe]
}

func ExampleFromJson() {
	type person struct {
		ID   int    `csv:"id"`
		Name string `csv:"name"`
	}
	reader := strings.NewReader(`[
		{"id":1,"name":"John Doe"},
		{"id":2,"name":"Jane Doe"}
		]`)

	res := TransformFn[person](FromJson[person](reader)).
		WithSteps(
			Map(func(in person) (string, error) {
				return in.Name, nil
			}),
		).
		AsSlice()

	fmt.Println(res)
	// Output: [John Doe Jane Doe]
}

func ExampleFromStreamingJson() {
	type person struct {
		ID   int    `csv:"id"`
		Name string `csv:"name"`
	}
	reader := strings.NewReader(`
		{"id":1,"name":"John Doe"}
		{"id":2,"name":"Jane Doe"}`)

	res := TransformFn[person](FromStreamingJson[person](reader)).
		WithSteps(
			Map(func(in person) (string, error) {
				return in.Name, nil
			}),
		).
		AsSlice()

	fmt.Println(res)
	// Output: [John Doe Jane Doe]
}
