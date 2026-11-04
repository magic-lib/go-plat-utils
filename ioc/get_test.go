package ioc_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/magic-lib/go-plat-utils/ioc"
)

type getTestService struct{ Name string }

// 未注册或类型不符时返回零值，不 panic
func TestGetMissingReturnsZero(t *testing.T) {
	if got := ioc.Get[*getTestService]("not-registered-component"); got != nil {
		t.Fatalf("未注册组件应返回 nil, got %v", got)
	}
	ioc.Register(&getTestService{Name: "a"}, ioc.WithComponentName("get-test-svc"))
	if got := ioc.Get[*getTestService]("get-test-svc"); got == nil || got.Name != "a" {
		t.Fatalf("Get = %v", got)
	}
	if got := ioc.Get[string]("get-test-svc"); got != "" {
		t.Fatalf("类型不符应返回零值, got %q", got)
	}
}

// 注册与注入并发执行，配合 -race 检查数据竞争
func TestInjectComponentsConcurrentRegister(t *testing.T) {
	c := ioc.NewContainer()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			c.Register(&getTestService{}, ioc.WithComponentName(fmt.Sprint("svc", i)))
		}(i)
		go func() {
			defer wg.Done()
			_ = c.InjectComponents()
		}()
	}
	wg.Wait()
}
