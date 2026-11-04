package templates_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/templates"
)

func BenchmarkRunString(b *testing.B) {
	e := templates.NewRuleExprEngine()
	args := map[string]any{"a": 3, "b": 5}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = e.RunString("{{a}} + {{b}} > 6", args)
	}
}
