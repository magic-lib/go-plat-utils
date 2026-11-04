package seqid_test

import (
	"testing"
	"time"

	"github.com/magic-lib/go-plat-cache/cache"
	"github.com/magic-lib/go-plat-utils/id-generator/seqid"
)

// SetStartId 只能往大调，不能让已发出的 ID 再次发出
func TestSetStartIdOnlyIncrease(t *testing.T) {
	gen, err := seqid.NewIDGenerator[int64](&seqid.IDGenerator[int64]{
		Namespace: "test", IdStore: cache.NewMemGoCache[int64](time.Hour, time.Hour), BatchSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[int64]bool{}
	for i := 0; i < 5; i++ {
		id, _ := gen.Generate("k")
		seen[id] = true
	}
	gen.SetStartId("k", 1) // 比已发出的小，应忽略
	if id, _ := gen.Generate("k"); seen[id] {
		t.Fatalf("SetStartId 调小后发出了重复 id %d", id)
	}
	gen.SetStartId("k", 1000) // 往大调生效
	if id, _ := gen.Generate("k"); id != 1000 {
		t.Fatalf("SetStartId(1000) 后应发 1000, got %d", id)
	}
}
