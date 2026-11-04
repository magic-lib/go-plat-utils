package conv_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

type derefErr struct{ msg string }

func (e *derefErr) Error() string { return e.msg } // 解引用接收者

type derefErrExtra struct{ msg string }

func (e *derefErrExtra) Error() string  { return e.msg }
func (e *derefErrExtra) GetMsg() string { return e.msg }

// typed-nil error 的 Error() 解引用接收者时不能 panic
func TestStringTypedNilErrorDeref(t *testing.T) {
	var e1 *derefErr
	var e2 *derefErrExtra
	if got := conv.String(error(e1)); got != "" {
		t.Errorf("String(typed-nil error) = %q", got)
	}
	if got := conv.String(e2); got != "" && got != "null" {
		t.Errorf("String(typed-nil error with extra method) = %q", got)
	}
}

type unmarshalDst struct{ Name string }

// 目标不是指针时返回错误，不 panic
func TestUnmarshalNonPointerDst(t *testing.T) {
	if err := conv.Unmarshal(map[string]any{"Name": "a"}, unmarshalDst{}); err == nil {
		t.Fatal("目标不是指针时应返回错误")
	}
}
