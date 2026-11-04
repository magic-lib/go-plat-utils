package deque_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/deque"
)

func TestSafeDequeAtOutOfRange(t *testing.T) {
	d := deque.New[int]()
	d.PushBack(1)
	if v, ok := d.At(0); !ok || v != 1 {
		t.Fatalf("At(0) = %v, %v", v, ok)
	}
	for _, i := range []int{-1, 1, 5} {
		if _, ok := d.At(i); ok {
			t.Fatalf("At(%d) 越界应返回 false", i)
		}
	}
}
