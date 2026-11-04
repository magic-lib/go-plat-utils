package crontab_test

import (
	"github.com/magic-lib/go-plat-utils/crontab"
	"github.com/robfig/cron/v3"
	"os"
	"testing"
)
import "fmt"

func TestCrontab(t *testing.T) {
	if os.Getenv("GO_PLAT_INTEGRATION") == "" {
		t.Skip("集成测试，启动定时任务后永久阻塞；设置 GO_PLAT_INTEGRATION=1 后运行")
	}

	_ = crontab.StartJobs(true, map[string]cron.FuncJob{
		"*/2 * * * *": func() {
			fmt.Println("1分钟1")
		},
		"0 02 17 * *": func() {
			fmt.Println("定点1")
		},
	})

	select {}
}
func TestCrontabLockKey(t *testing.T) {
	if os.Getenv("GO_PLAT_INTEGRATION") == "" {
		t.Skip("集成测试，启动定时任务后永久阻塞；设置 GO_PLAT_INTEGRATION=1 后运行")
	}
	_ = crontab.StartJobs(true, map[string]cron.FuncJob{
		"* * * * *": func() {
			fmt.Println("1分钟1")
		},
	})
	_ = crontab.StartJobs(true, map[string]cron.FuncJob{
		"* * * * *": func() {
			fmt.Println("1分钟2")
		},
	})

	select {}
}
