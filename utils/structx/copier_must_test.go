package structx

import "testing"

// must,nopanic 字段缺失时应返回错误（原来错误被循环内的同名变量遮蔽而丢失）
func TestCopierMustNoPanicError(t *testing.T) {
	type from struct{ B string }
	type to struct {
		A string `copier:"must,nopanic"`
	}
	var dst to
	if err := Copy(&dst, &from{B: "x"}); err == nil {
		t.Fatal("expected error for missing must field")
	}
}
