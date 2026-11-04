package id_test

import (
	"sync"
	"testing"

	"github.com/magic-lib/go-plat-utils/id-generator/id"
)

// 并发生成大量 id，校验没有重复
func checkUniqueConcurrent(t *testing.T, name string, gen func() string) {
	const goroutines, perGoroutine = 20, 500
	var mu sync.Mutex
	seen := make(map[string]struct{}, goroutines*perGoroutine)
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ids := make([]string, 0, perGoroutine)
			for j := 0; j < perGoroutine; j++ {
				ids = append(ids, gen())
			}
			mu.Lock()
			defer mu.Unlock()
			for _, one := range ids {
				if _, ok := seen[one]; ok {
					t.Errorf("%s 生成了重复 id: %s", name, one)
				}
				seen[one] = struct{}{}
			}
		}()
	}
	wg.Wait()
}

func TestSnowflakeUnique(t *testing.T) {
	checkUniqueConcurrent(t, "Snowflake", func() string { return id.Snowflake().String() })
	checkUniqueConcurrent(t, "Snowflake(5)", func() string { return id.Snowflake(5).String() })
}

func TestGeneratorUnique(t *testing.T) {
	checkUniqueConcurrent(t, "GeneratorBase32", id.GeneratorBase32)
}
