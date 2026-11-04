package templates

import "testing"

// 替换值中的 $ 按字面量保留
func TestTemplateReplaceDollar(t *testing.T) {
	got := NewTemplate("price: {{a}}, all: {{.}}", DefaultPrefix, DefaultSuffix).Replace(map[string]any{"a": "$100"})
	want := `price: $100, all: {"a":"$100"}`
	if got != want {
		t.Fatalf("Replace = %q, want %q", got, want)
	}
}

// 同一 format 用不同分隔符渲染，不能命中按其他分隔符解析的缓存
func TestTemplateCacheByDelims(t *testing.T) {
	format := "[[ .a ]]-{{ .a }}"
	data := map[string]any{"a": "x"}
	got1, err1 := Template(format, data)
	got2, err2 := Template(format, data, "[[", "]]")
	if err1 != nil || err2 != nil || got1 != "[[ .a ]]-x" || got2 != "x-{{ .a }}" {
		t.Fatalf("got1=%q(%v) got2=%q(%v)", got1, err1, got2, err2)
	}
}

// && / || 两侧取到的是 "true" 字符串时不能 panic
func TestRuleExprStrBoolString(t *testing.T) {
	ok, err := ruleExprStr("a && b", `{"a":true,"b":true}`)
	if err != nil || !ok {
		t.Fatalf("ruleExprStr = %v, %v", ok, err)
	}
	ok, err = ruleExprStr("a || b", `{"a":false,"b":true}`)
	if err != nil || !ok {
		t.Fatalf("ruleExprStr(||) = %v, %v", ok, err)
	}
}
