package goroutines

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	gocache "github.com/patrickmn/go-cache"
	"github.com/timandy/routine"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	expiration      = 20 * time.Minute //一个ctx的最长时间，避免长时间占用内存
	cleanupInterval = 30 * time.Minute
	baseInt         = 10
	onceCache       = sync.Once{}
	ctxCache        *gocache.Cache
	localStorage    routine.ThreadLocal[string]
)

// getInitCache 初始化并获得缓存
func getInitCache() *gocache.Cache {
	if ctxCache == nil {
		onceCache.Do(func() {
			ctxCache = gocache.New(expiration, cleanupInterval)
			localStorage = routine.NewInheritableThreadLocal[string]()
			localStorage.Set(getCurrentGoId()) //初始化
		})
	}
	return ctxCache
}

func getCurrentGoIdFromRuntime() int64 {
	var buf [64]byte
	// 获取当前 goroutine 的栈信息
	_ = runtime.Stack(buf[:], false)
	firstLine := strings.Split(string(buf[:]), "\n")[0]
	re := regexp.MustCompile(`\d+`)
	numbers := re.FindAllString(firstLine, -1)
	if len(numbers) > 0 {
		// 不知道是哪个数字，就进行组合起来，必然会一致的，并且会不同
		if goIdTemp, ok := conv.Int64(strings.Join(numbers, "")); ok {
			return goIdTemp
		}
	}
	return -1
}

// getCurrentGoId 取得当前的协程ID
func getCurrentGoId() string {
	goId := getCurrentGoIdFromRuntime()
	if goId > 0 {
		return strconv.FormatInt(goId, baseInt)
	}
	return strconv.FormatInt(routine.Goid(), baseInt)
}

// InitContext 设置上下文，需要在入口协程上执行，会返回当前的IdKey
func InitContext(ctx ...*context.Context) {
	ctxFactory := getInitCache()
	ctxKey := getCurrentGoId()
	localStorage.Set(ctxKey)
	if len(ctx) > 0 {
		ctxFactory.Set(ctxKey, ctx[0], expiration)
	}
}

// SetContext 设置上下文
func SetContext(ctx *context.Context) {
	_, ctxKey, _ := GetContext()
	if ctxKey == "" {
		ctxKey = getCurrentGoId() //会取到子协程的ID，所以这里会有问题
	}
	ctxFactory := getInitCache()
	ctxFactory.Set(ctxKey, ctx, expiration)
	if localStorage.Get() == "" {
		localStorage.Set(ctxKey)
	}
}

// GetContext 获取上下文
func GetContext() (ctx *context.Context, ctxKey string, err error) {
	ctxFactory := getInitCache()

	ctxKey = localStorage.Get()
	if ctxKey == "" {
		return nil, "", fmt.Errorf("需要入口处调用 InitContext 进行设置。")
	}

	val, ok := ctxFactory.Get(ctxKey)
	if !ok {
		return nil, ctxKey, fmt.Errorf("需要入口处调用 InitContext 进行设置, ctxKey: %s", ctxKey)
	}
	ctx, ok = val.(*context.Context)
	if ok {
		return ctx, ctxKey, nil
	}
	return nil, ctxKey, fmt.Errorf("需要入口处调用 SetContext 进行设置, 存储格式错误。 ctxKey: %s", ctxKey)
}

// DelContext 删除上下文
func DelContext() {
	ctxFactory := getInitCache()
	ctxKey := localStorage.Get()
	if ctxKey == "" {
		return
	}
	ctxFactory.Delete(ctxKey)
	localStorage.Remove()
}
