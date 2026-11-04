package plugins_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/magic-lib/go-plat-utils/plugins"
)

type testPlugin struct{ name string }

func (p *testPlugin) Name() string { return p.name }
func (p *testPlugin) Execute(_ context.Context, args any) (any, error) {
	return args, nil
}

// 空 namespace 注册后应能按空 namespace 或 default 加载，多次注册不会互相覆盖
func TestRegisterEmptyNamespace(t *testing.T) {
	if err := plugins.Register("", &testPlugin{name: "empty-ns-p1"}); err != nil {
		t.Fatalf("Register err: %v", err)
	}
	if err := plugins.Register("", &testPlugin{name: "empty-ns-p2"}); err != nil {
		t.Fatalf("Register err: %v", err)
	}
	for _, ns := range []string{"", "default"} {
		for _, name := range []string{"empty-ns-p1", "empty-ns-p2"} {
			if _, err := plugins.Load(ns, name); err != nil {
				t.Errorf("Load(%q, %q) err: %v", ns, name, err)
			}
		}
	}
}

// 并发向同一个新 namespace 注册，所有插件都应保留
func TestRegisterConcurrentNewNamespace(t *testing.T) {
	const ns, n = "concurrent-ns", 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = plugins.Register(ns, &testPlugin{name: fmt.Sprintf("p%d", i)})
		}(i)
	}
	wg.Wait()
	all, err := plugins.LoadAll(ns)
	if err != nil || len(all) != n {
		t.Fatalf("LoadAll = %d 个, err=%v; want %d", len(all), err, n)
	}
}
