package goroutines_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAsyncExecuteDataList(t *testing.T) {
	arr := make([]int, 0)
	for i := 0; i < 100; i++ {
		arr = append(arr, i+1)
	}
	num := 0
	var lock sync.Mutex
	ret, err := goroutines.AsyncExecuteDataList(12*time.Second, arr, func(value int, key int) (breakFlag bool, err error) {
		fmt.Println("key=", key, "; value=", value)
		time.Sleep(1 * time.Second) //模拟执行耗时
		lock.Lock()
		defer lock.Unlock()
		num++ //可以将执行结果放入数组中
		return false, fmt.Errorf("3333")
	})
	fmt.Println(num, ret, err)
}
func TestAsyncForEachWhile(t *testing.T) {
	arr := make([]int, 0)
	for i := 0; i < 10; i++ {
		arr = append(arr, i+1)
	}
	num := 0
	ret, err := goroutines.AsyncForEachWhile(arr, func(value int, key int) (bool, error) {
		fmt.Println("value=", value)
		num++
		//time.Sleep(1 * time.Second)
		if value > 4 {
			return true, fmt.Errorf("eeee")
		}
		return true, nil
	}, goroutines.AsyncForEachWhileOptions{ChunkSize: 2, MaxConcurrency: 2})
	fmt.Println(num, ret, err)
}
func TestGoroutineId(t *testing.T) {

	// 假设这是第一行文本
	firstLine := "goroutine 75 [running]:\ngithub.com/magic-lib/go-plat-utils/gorout"
	// 正则表达式匹配一个或多个数字
	firstLine = strings.Split(firstLine, "\n")[0]

	re := regexp.MustCompile(`\d+`)
	numbers := re.FindAllString(firstLine, -1)
	fmt.Println(numbers)
}
