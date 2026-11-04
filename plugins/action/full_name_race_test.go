package action

import (
	"sync"
	"testing"
)

// FullName 并发调用不能有数据竞争（配合 -race）
func TestFullNameConcurrent(t *testing.T) {
	am := &ActMetaData{Namespace: "ns", Action: "act"}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if am.FullName() != "ns"+namespaceLinkNameChar+"act" {
				t.Error("unexpected full name")
			}
		}()
	}
	wg.Wait()
}
