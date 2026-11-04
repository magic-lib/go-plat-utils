package ioc

import (
	"context"
	"testing"
)

type nilEmbedInner struct {
	Name string `inject:"x"`
}

type nilEmbedOuter struct {
	*nilEmbedInner
}

// 嵌入指针为 nil 时跳过，不能 panic
func TestInjectNilEmbeddedPointer(t *testing.T) {
	c := NewContainer()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Inject panic: %v", r)
		}
	}()
	if err := c.Inject(context.Background(), &nilEmbedOuter{}); err != nil {
		t.Fatal(err)
	}
}
