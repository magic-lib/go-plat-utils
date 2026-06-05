package logs_test

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/magic-lib/go-plat-utils/logs"
	"testing"
	"time"
)

func TestLoggerData(t *testing.T) {
	logData := logs.NewLogData(&logs.LogCommData{
		CreateTime: time.Now(),
		LogId:      "logid12345",
		//UserId:     "userid44444",
		Env:    "dev",
		Path:   "/name/pri",
		Method: "get",
		//Extend: map[string]any{
		//	"service": "dnf",
		//},
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
		//UserId:     "userid44444",
		Env:    "dev",
		Path:   "/name/pri",
		Method: "get",
		//Extend: map[string]any{
		//	"service": "dnf",
		//},
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

type TrafficLog struct {
	LogId           int64     `db:"log_id" json:"log_id"`
	LogType         string    `db:"log_type" json:"log_type"`                 // 日志类型
	RequestId       string    `db:"request_id" json:"request_id"`             // 请求唯一标识
	UserId          int64     `db:"user_id" json:"user_id"`                   // 关联user表id
	AccountId       int64     `db:"account_id" json:"account_id"`             // 关联account表id
	Nid             string    `db:"nid" json:"nid"`                           // 客户Nrc
	Mobile          string    `db:"mobile" json:"mobile"`                     // 客户端mobile
	Ip              string    `db:"ip" json:"ip"`                             // 客户端IP地址
	Method          string    `db:"method" json:"method"`                     // HTTP方法(GET/POST/PUT/DELETE等)
	Url             string    `db:"url" json:"url"`                           // 请求URL
	Path            string    `db:"path" json:"path"`                         // 请求路径
	StatusCode      int       `db:"status_code" json:"status_code"`           // HTTP状态码
	ResponseTime    int       `db:"response_time" json:"response_time"`       // 响应时间(毫秒)
	RequestSize     int       `db:"request_size" json:"request_size"`         // 请求大小(字节)
	ResponseSize    int       `db:"response_size" json:"response_size"`       // 响应大小(字节)
	UserAgent       string    `db:"user_agent" json:"user_agent"`             // 用户代理
	Referer         string    `db:"referer" json:"referer"`                   // 来源页面
	RequestHeaders  string    `db:"request_headers" json:"request_headers"`   // 请求头信息(JSON格式)
	ResponseHeaders string    `db:"response_headers" json:"response_headers"` // 响应头信息(JSON格式)
	RequestBody     string    `db:"request_body" json:"request_body"`         // 请求体内容
	ResponseBody    string    `db:"response_body" json:"response_body"`       // 响应体内容
	ErrorMessage    string    `db:"error_message" json:"error_message"`       // 错误信息
	Extra1          string    `db:"extra1" json:"extra1"`                     // 扩展1
	Extra2          string    `db:"extra2" json:"extra2"`                     // 扩展2
	Extra3          string    `db:"extra3" json:"extra3"`                     // 扩展3
	CreateTime      time.Time `db:"create_time" json:"create_time"`           // 创建时间
	UpdateTime      time.Time `db:"update_time" json:"update_time"`           // 更新时间
}

func TestAAAA(t *testing.T) {
	msg := new(TrafficLog)
	msg.CreateTime = time.Now()
	logData := new(logs.LogData)

	//conv.OpenUnmarshalLog()
	_ = conv.Unmarshal(msg, logData)

	fmt.Println(conv.String(logData))

}
