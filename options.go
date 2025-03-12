package steps

import (
	"context"
	"fmt"
	"io"
	"os"
)

// WithName adds a name to the transformer
func WithName(name string) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.Name = name
	}
}

// WithLogWriter sets the log writer
func WithLogWriter(writer io.Writer) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.LogWriter = writer
	}
}

// WithErrorHandler sets the error handler or the transformer
func WithErrorHandler(handler func(error)) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.ErrorHandler = handler
	}
}

// WithPanicHandler sets the panic handler used by the transformer input
func WithPanicHandler(handler func(error)) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.PanicHandler = handler
	}
}

// WithContext sets the context of the transformer (used to cancel operations)
func WithContext(ctx context.Context) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.Ctx = ctx
	}
}

// WithChanSize sets the channel size used by streaming inputs
func WithChanSize(size uint) func(*TransformerOptions) {
	return func(opts *TransformerOptions) {
		opts.ChanSize = size
	}
}

func buildOpts(options ...func(*TransformerOptions)) TransformerOptions {
	opts := TransformerOptions{
		Ctx:       context.Background(),
		LogWriter: os.Stdout,
	}
	for _, withOption := range options {
		withOption(&opts)
	}
	if opts.ErrorHandler == nil {
		opts.ErrorHandler = func(err error) {
			fmt.Fprintln(opts.LogWriter, "error occured:", err)
		}
	}
	if opts.PanicHandler == nil {
		opts.PanicHandler = func(err error) {
			panic(err)
		}
	}
	return opts
}
