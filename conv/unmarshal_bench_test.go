package conv_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

type benchUserT struct {
	Id      int64             `json:"id"`
	Name    string            `json:"name"`
	Tags    []string          `json:"tags"`
	Extra   map[string]string `json:"extra"`
	Enabled bool              `json:"enabled"`
}

func BenchmarkUnmarshalMapToStruct(b *testing.B) {
	src := map[string]any{"id": 1, "name": "a", "tags": []any{"x", "y"}, "extra": map[string]any{"k": "v"}, "enabled": true}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var u benchUserT
		_ = conv.Unmarshal(src, &u)
	}
}
