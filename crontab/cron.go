package crontab

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/robfig/cron/v3"
	"log"
	"sync"
)

type cronInstance struct {
	isStart bool
	c       *cron.Cron
}

var (
	oneCrontab *cronInstance
	runningMu  sync.Mutex
	cOnce      sync.Once
)

// getCron 初始化
func getCron() *cronInstance {
	cOnce.Do(func() {
		oneCrontab = &cronInstance{
			isStart: false,
			c:       cron.New(),
		}
	})
	return oneCrontab
}

/*
//按分钟开始定时
crontab.Start(map[string]func(){
		"* * * * *" : func(){
},
})
minute     = field(fields[1], minutes)
hour       = field(fields[2], hours)
dayOfMonth = field(fields[3], dom)
month      = field(fields[4], months)
dayOfWeek  = field(fields[5], dow)
*/

// StartJobs 启动定时任务，格式：分钟 小时 天 月 星期
func StartJobs(immediateRun bool, jobs ...map[string]func()) error {
	runningMu.Lock()
	defer runningMu.Unlock()

	if len(jobs) == 0 {
		return nil
	}
	oneCron := getCron()

	allKey := make([]string, 0)
	for _, jobMap := range jobs {
		for key, _ := range jobMap {
			allKey = append(allKey, key)
		}
	}

	log.Println("[crontab] StartJobs start:", allKey)

	runList := make([]func(), 0)

	allSuccess := true //如果全部出错的，则不用启动
	for _, jobMap := range jobs {
		for times := range jobMap {
			var err error
			if immediateRun {
				runList = append(runList, jobMap[times])
			}
			//列表里需要将所有的内容保存一份，这样就可以到时候进行删除了
			_, err = oneCron.c.AddFunc(times, jobMap[times])
			if err != nil {
				log.Println("[crontab] StartJobs error:", times, err.Error())
				allSuccess = false
			}
		}
	}

	//没有添加成功，则不用启动
	if !allSuccess {
		return fmt.Errorf("[crontab] StartJobs allSuccess Fail")
	}

	//立即执行：每次调用都要执行本次传入的任务（不能放在 isStart 判断之后，否则第二次调用会被跳过）
	if immediateRun {
		goroutines.GoAsync(func(params ...any) {
			for _, runFunc := range runList {
				runFunc()
			}
		})
	}

	//cron 的调度循环只会启动一次，后续调用只是向已运行的调度器追加任务
	if oneCron.isStart {
		return nil
	}
	oneCron.isStart = true

	//异步启动
	goroutines.GoAsync(func(params ...any) {
		oneCron.c.Run()
		select {}
	})
	return nil
}
