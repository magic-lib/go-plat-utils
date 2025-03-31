package snowflake_test

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/id-generator/snowflake"
	"testing"
)

func TestGetId(t *testing.T) {
	const cnt int = 10000
	worker, err := snowflake.New(&snowflake.Worker{
		WorkerId: 1,
		CenterId: 1,
	})
	if err != nil {
		t.Fatalf("Init Worker Error")
		return
	}
	// 启动1万个协程同时访问
	hm := make(map[int64]struct{})
	ch := make(chan int64, cnt)
	defer close(ch)
	done := make(chan struct{})
	defer close(done)
	t.Logf("开始生成\n")
	go func() {
		i := 0
		for {
			select {
			case num := <-ch:
				i++
				if _, ok := hm[num]; ok {
					t.Errorf("SnowFlake Id is not unique !\n")
					done <- struct{}{}
					return
				}
				hm[num] = struct{}{}
				if i == cnt {
					t.Logf("取完%d次数据了，监控协程退出\n", cnt)
					done <- struct{}{}
					return
				}
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < cnt; i++ {
		go func(ctx context.Context) {
			select {
			case <-ctx.Done():
				t.Logf("有重复元素, 退出生成\n")
				return
			default:
				id := worker.NextId()
				ch <- id
			}
		}(ctx)
	}

	<-done
	cancel()
	t.Logf("完成%d次请求, 没有重复ID\n", cnt)
}

func TestGetIdList(t *testing.T) {
	snowflake.DefaultConfig(&snowflake.Config{
		CenterBits: 23,
		WorkerBits: 20,
		Epoch:      1711442064000, //北京时间：2024-03-26 16:34:24
		SeqBits:    20,
	})

	worker, err := snowflake.New(&snowflake.Worker{
		WorkerId: 1,
		CenterId: 1,
	})
	if err != nil {
		t.Fatalf("Init Worker Error")
		return
	}
	list := worker.NextIdList(1)
	fmt.Printf("%b\n", list[0])
}
