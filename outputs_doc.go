package steps

import (
	"io"
	"iter"
)

// these functions are only here to hack the documentation
//

func _stepsTransformer_AsRange() iter.Seq[any]               { return nil }
func _stepsTransformer_AsKeyValueRange() iter.Seq2[any, any] { return nil }
func _stepsTransformer_AsIndexedRange() iter.Seq2[any, any]  { return nil }
func _stepsTransformer_AsMultiMap() map[any][]any            { return nil }
func _stepsTransformer_AsMap() map[any]any                   { return nil }
func _stepsTransformer_AsSlice() []any                       { return nil }
func _stepsTransformer_AsCsv() string                        { return "" }
func _stepsTransformer_ToStreamingCsv(writer io.Writer)      {}
func _stepsTransformer_AsJson() string                       { return "" }
func _stepsTransformer_ToStreamingJson(writer io.Writer)     {}
