package mysqllog

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 写日志读取扩展字段与建表写入 createdTables 并发执行，配合 -race 检查
func TestGetTableExtendFieldListConcurrent(t *testing.T) {
	tm := &tableManager{cfg: &Config{TablePrefix: "app_log"}, createdTables: map[string][]string{}}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			tm.mu.Lock()
			tm.createdTables[fmt.Sprintf("app_log_%d", i)] = []string{"id"}
			tm.createdTables[tm.cfg.getTableName(time.Now())] = []string{"id"}
			tm.mu.Unlock()
		}(i)
		go func() {
			defer wg.Done()
			_ = tm.getTableExtendFieldList()
		}()
	}
	wg.Wait()
}
