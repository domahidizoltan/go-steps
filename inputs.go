package steps

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"

	"github.com/jszwec/csvutil"
)

func FromCsv[T any](reader io.Reader, inputOptions ...func(*inputOption)) []T {
	opts := makeOptions(inputOptions...)

	data, err := io.ReadAll(reader)
	if err != nil {
		opts.panicHandler(err)
	}
	var res []T
	if err := csvutil.Unmarshal(data, &res); err != nil {
		opts.panicHandler(err)
	}
	return res
}

func FromStreamingCsv[T any](reader io.Reader, inputOptions ...func(*inputOption)) chan T {
	opts := makeOptions(inputOptions...)
	resCh := make(chan T)
	dec, err := csvutil.NewDecoder(csv.NewReader(reader))
	if err != nil {
		opts.panicHandler(err)
	}

	go func(dec *csvutil.Decoder, resCh chan T) {
		defer close(resCh)
		for {
			var data T
			if err := dec.Decode(&data); err != nil {
				if err == io.EOF {
					break
				}
				opts.panicHandler(err)
			}
			resCh <- data
		}
	}(dec, resCh)

	return resCh
}

func FromJson[T any](reader io.Reader, inputOptions ...func(*inputOption)) []T {
	opts := makeOptions(inputOptions...)

	data, err := io.ReadAll(reader)
	if err != nil {
		opts.panicHandler(err)
	}
	var res []T
	if err := json.Unmarshal(data, &res); err != nil {
		opts.panicHandler(err)
	}
	return res
}

func FromStreamingJson[T any](reader io.Reader, inputOptions ...func(*inputOption)) chan T {
	opts := makeOptions(inputOptions...)
	resCh := make(chan T)
	dec := json.NewDecoder(reader)

	go func(dec *json.Decoder, resCh chan T) {
		defer close(resCh)
		for {
			var data T
			if err := dec.Decode(&data); err != nil {
				if err == io.EOF {
					break
				}
				opts.panicHandler(err)
			}
			resCh <- data
		}
	}(dec, resCh)

	return resCh
}

func File(filePath string) io.Reader {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	return f
}

type inputOption struct {
	panicHandler func(err error)
}

func WithPanicHandler(panicHandler func(err error)) func(*inputOption) {
	return func(opts *inputOption) {
		opts.panicHandler = panicHandler
	}
}

func makeOptions(opts ...func(*inputOption)) inputOption {
	o := inputOption{
		panicHandler: func(err error) {
			panic(err)
		},
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
