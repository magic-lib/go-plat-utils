package crontab

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/robfig/cron/v3"
	"log"
	"sync"
)

// JobFunc 定时任务函数
type JobFunc func()

type cronInstance struct {
	running  bool
	c        *cron.Cron
	entryMap map[string]cron.EntryID // expr -> entryID
	mu       sync.RWMutex
}

var (
	oneCrontab *cronInstance
	cOnce      sync.Once
)

// getCron 初始化
func getCron() *cronInstance {
	cOnce.Do(func() {
		oneCrontab = &cronInstance{
			running:  false,
			c:        cron.New(),
			entryMap: make(map[string]cron.EntryID),
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

// StartJobs 批量添加定时任务
// expr格式：分 时 日 月 周
// immediateRun：是否添加成功后立刻执行一遍
func StartJobs(immediateRun bool, jobs ...map[string]JobFunc) error {
	oneCron := getCron()
	oneCron.mu.Lock()
	defer oneCron.mu.Unlock()

	if len(jobs) == 0 {
		return nil
	}

	log.Println("[crontab] StartJobs start:")

	var (
		allKey   = make([]string, 0)
		runList  = make([]JobFunc, 0)
		rollback = make([]cron.EntryID, 0)
		failExpr string
	)

	for _, jobMap := range jobs {
		for times, fn := range jobMap {
			allKey = append(allKey, times)
			if immediateRun {
				runList = append(runList, fn)
			}
			//列表里需要将所有的内容保存一份，这样就可以到时候进行删除了
			eid, err := oneCron.c.AddFunc(times, fn)
			if err != nil {
				failExpr = times
				log.Printf("[crontab] StartJobs add fail, times=%s, err=%s", times, err.Error())
				// 回滚：删掉本次已经加上的任务
				for _, rid := range rollback {
					oneCron.c.Remove(rid)
				}
				return fmt.Errorf("crontab add job failed, expr=%s: %w", failExpr, err)
			}
			rollback = append(rollback, eid)
			oneCron.entryMap[times] = eid
		}
	}

	//没有添加成功，则不用启动
	log.Println("[crontab] StartJobs loaded times:", allKey)

	//立即执行：每次调用都要执行本次传入的任务（不能放在 isStart 判断之后，否则第二次调用会被跳过）
	if immediateRun {
		goroutines.GoAsync(func(params ...any) {
			for _, runFn := range runList {
				runFn()
			}
		})
	}

	//cron 的调度循环只会启动一次，后续调用只是向已运行的调度器追加任务
	if oneCron.running {
		return nil
	}
	oneCron.running = true

	//异步启动
	goroutines.GoAsync(func(params ...any) {
		oneCron.c.Run()
	})
	return nil
}

// Stop 优雅停止，等待正在运行任务全部结束
func Stop() {
	oneCron := getCron()
	oneCron.mu.Lock()
	defer oneCron.mu.Unlock()

	ctx := oneCron.c.Stop()
	<-ctx.Done()
	oneCron.running = false
	log.Println("[crontab] cron stopped, all running jobs finished")
}
