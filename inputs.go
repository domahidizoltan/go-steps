package steps

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"

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

func FromStreamingCsv[T any](reader io.Reader, withoutHeaders bool) func(TransformerOptions) chan T {
	return func(opts TransformerOptions) chan T {
		resCh := make(chan T, opts.ChanSize)

		header := []string{}
		if withoutHeaders {
			var t T
			var err error
			header, err = csvutil.Header(t, "csv")
			if err != nil {
				close(resCh)
				opts.PanicHandler(err)
				return resCh
			}
		}

		dec, err := csvutil.NewDecoder(csv.NewReader(reader), header...)
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
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanLines)

		go func(scanner *bufio.Scanner, resCh chan T) {
			defer close(resCh)
			for {
				select {
				case <-opts.Ctx.Done():
					opts.ErrorHandler(opts.Ctx.Err())
					return
				default:
					if !scanner.Scan() {
						return
					}
					if err := scanner.Err(); err != nil {
						opts.PanicHandler(err)
						continue
					}

					line := scanner.Text()
					switch strings.TrimSpace(line) {
					case "", "[", "]":
						continue
					default:
						var data T
						if err := json.Unmarshal([]byte(line), &data); err != nil {
							opts.PanicHandler(err)
							continue
						}
						resCh <- data
					}

				}
			}
		}(scanner, resCh)

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
