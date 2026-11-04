package commerror

import "testing"

// Wrap 添加 detail 不能改到原错误
func TestWrapDoesNotShareDetails(t *testing.T) {
	var origin CommError = &commErr{code: 1, message: "m", details: map[string]any{"a": 1}}
	wrapped := Wrap(origin, WithDetail("b", 2))
	if _, ok := origin.Details()["b"]; ok {
		t.Fatal("origin details mutated by Wrap")
	}
	if wrapped.Details()["a"] != 1 || wrapped.Details()["b"] != 2 {
		t.Fatalf("wrapped details = %v", wrapped.Details())
	}
}
