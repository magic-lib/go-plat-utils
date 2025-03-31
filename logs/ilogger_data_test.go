package logs_test

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/magic-lib/go-plat-utils/logs"
	"testing"
	"time"
)

func TestLoggerData(t *testing.T) {
	logData := logs.NewLogData(&logs.LogCommData{
		CreateTime: time.Now(),
		LogId:      "logid12345",
		UserId:     "userid44444",
		Env:        "dev",
		Path:       "/name/pri",
		Method:     "get",
		Extend: map[string]interface{}{
			"service": "dnf",
		},
	})

	time.Sleep(5 * time.Millisecond)

	logData.AddMessage(logs.DEBUG, "fileName.go", 112, "有了错误", "有了第二个错误")

	str := logData.String()

	fmt.Println(str)

}
func TestPrintLogger(t *testing.T) {
	prtLogger := logs.NewPrintLogger(logs.DEBUG, &logs.LogCommData{
		CreateTime: time.Now(),
		LogId:      "",
		UserId:     "userid44444",
		Env:        "dev",
		Path:       "/name/pri",
		Method:     "get",
		Extend: map[string]interface{}{
			"service": "dnf",
		},
	})

	time.Sleep(5 * time.Millisecond)

	ctx := context.Background()
	sonCtx := context.WithValue(ctx, "logid", "1234567")
	goroutines.SetContext(&sonCtx)

	prtLogger.Debug("有了错误", "有了第二个错误")
	prtLogger.Error("有了第二个错误dddd")

}
func TestDefaultLogger(t *testing.T) {
	prtLogger := logs.DefaultLogger()

	prtLogger.Debug("有了错误", "有了第二个错误")
	prtLogger.Error("有了第二个错误dddd")

	ctx := context.Background()
	sonCtx := context.WithValue(ctx, "logid", "1234567")

	ctxLogger := logs.CtxLogger(sonCtx)
	ctxLogger.Info("aaaaa")

}

func getOneInt(n int) bool {
	arrList := [][]int{{
		1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33, 35, 37, 39, 41, 43, 45, 47, 49, 51, 53, 55, 57, 59, 61, 63,
	}, {
		2, 3, 6, 7, 10, 11, 14, 15, 18, 19, 22, 23, 26, 27, 30, 31, 34, 35, 38, 39, 42, 43, 46, 47, 50, 51, 54, 55, 58, 59, 62, 63,
	}, {
		4, 5, 6, 7, 12, 13, 14, 15, 20, 21, 22, 23, 28, 29, 30, 31, 36, 37, 38, 39, 44, 45, 46, 47, 52, 53, 54, 55, 60, 61, 62, 63,
	}, {
		8, 9, 10, 11, 12, 13, 14, 15, 24, 25, 26, 27, 28, 29, 30, 31, 40, 41, 42, 43, 44, 45, 46, 47, 56, 57, 58, 59, 60, 61, 62, 63,
	}, {
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63,
	}, {
		32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63,
	}}

	num := 0
	for _, v := range arrList {
		for _, vv := range v {
			if vv == n {
				num = num + v[0]
				break
			}
		}
	}
	if num == n {
		return true
	}
	return false
}

func TestGuessNumber(t *testing.T) {
	for i := 1; i <= 63; i++ {
		if !getOneInt(i) {
			fmt.Println("error:", i)
		} else {
			fmt.Println("ok:", i)
		}
	}

}
