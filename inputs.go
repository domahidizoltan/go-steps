package steps

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"

	"github.com/jszwec/csvutil"
)

func FromCsv[T any](reader io.Reader) func(TransformerOptions) []T {
	return func(opts TransformerOptions) []T {
		data, err := io.ReadAll(reader)
		if err != nil && err != io.EOF {
			opts.PanicHandler(err)
			return nil
		}
		var res []T
		if err := csvutil.Unmarshal(data, &res); err != nil {
			opts.PanicHandler(err)
			return nil
		}
		return res
	}
}

func FromStreamingCsv[T any](reader io.Reader) func(TransformerOptions) chan T {
	return func(opts TransformerOptions) chan T {
		resCh := make(chan T, opts.ChanSize)
		dec, err := csvutil.NewDecoder(csv.NewReader(reader))
		if err != nil && err != io.EOF {
			close(resCh)
			opts.PanicHandler(err)
			return resCh
		}

		go func(dec *csvutil.Decoder, resCh chan T) {
			defer close(resCh)
			for {
				select {
				case <-opts.Ctx.Done():
					opts.ErrorHandler(opts.Ctx.Err())
					return
				default:
					var data T
					if err := dec.Decode(&data); err != nil {
						if err == io.EOF {
							return
						}
						opts.PanicHandler(err)
						continue
					}
					resCh <- data
				}
			}
		}(dec, resCh)

		return resCh
	}
}

func FromJson[T any](reader io.Reader) func(TransformerOptions) []T {
	return func(opts TransformerOptions) []T {
		data, err := io.ReadAll(reader)
		if err != nil && err != io.EOF {
			opts.PanicHandler(err)
			return nil
		}
		var res []T
		if err := json.Unmarshal(data, &res); err != nil {
			opts.PanicHandler(err)
			return nil
		}
		return res
	}
}

func FromStreamingJson[T any](reader io.Reader) func(TransformerOptions) chan T {
	return func(opts TransformerOptions) chan T {
		resCh := make(chan T, opts.ChanSize)
		dec := json.NewDecoder(reader)

		go func(dec *json.Decoder, resCh chan T) {
			defer close(resCh)
			for {
				select {
				case <-opts.Ctx.Done():
					opts.ErrorHandler(opts.Ctx.Err())
					return
				default:
					var data T
					if err := dec.Decode(&data); err != nil {
						if err == io.EOF {
							return
						}
						opts.PanicHandler(err)
						continue
					}
					resCh <- data
				}
			}
		}(dec, resCh)

		return resCh
	}
}

func File(filePath string) io.Reader {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	return f
}
