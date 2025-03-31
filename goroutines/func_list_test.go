package goroutines_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestAsyncExecuteDataList(t *testing.T) {
	arr := make([]int, 0)
	for i := 0; i < 100; i++ {
		arr = append(arr, i+1)
	}
	num := 0
	ret, err := goroutines.AsyncExecuteDataList(1*time.Second, arr, func(key int, value int) (breakFlag bool, err error) {
		fmt.Println("key=", key, "; value=", value)
		num++
		time.Sleep(1 * time.Second)
		return true, fmt.Errorf("3333")
	})
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
