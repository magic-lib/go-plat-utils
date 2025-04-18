package logs

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conf"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/magic-lib/go-plat-utils/utils/httputil"
	"github.com/magic-lib/go-plat-utils/utils/httputil/param"
	"path/filepath"
	"time"
)

type (
	LogExecute func(ctx context.Context, logInfo *LogData) //日志的处理函数
)

// LogCommData 不会改变的数据
type LogCommData struct {
	CreateTime time.Time      `json:"createTime"`       //第一条日志的创建时间
	LogId      string         `json:"id"`               //logId
	UserId     string         `json:"userid,omitempty"` //userID
	Env        conf.EnvCode   `json:"env"`              //env
	Path       string         `json:"path,omitempty"`   //当前请求的地址
	Method     string         `json:"method,omitempty"` //当前请求的方法
	Extend     map[string]any `json:"extend,omitempty"` //额外的业务参数
}

// LogData 每条单独日志的数据
type LogData struct {
	LogCommData
	Now      time.Time `json:"now"` //初始化时间
	FileName string    //文件名
	Line     int       //行号
	LogLevel LogLevel  `json:"logLevel"`
	Message  []any     `json:"message"`
}

// Init 初始化
func (l *LogCommData) Init() {
	if cond.IsTimeEmpty(l.CreateTime) {
		l.CreateTime = time.Now()
	}
	if l.LogId == "" {
		l.LogId = httputil.GetLogId()
	}
}

// NewLogData 初始化一个日志变量
func NewLogData(logCommData ...*LogCommData) *LogData {
	l := new(LogData)
	if logCommData != nil && len(logCommData) > 0 {
		if logCommData[0] != nil {
			l.LogCommData = *(logCommData[0])
		}
	}

	logData := &l.LogCommData
	logData.Init()

	return l
}

// AddMessage 将日志添加
func (l *LogData) AddMessage(level LogLevel, fileName string, line int, msg ...any) {
	if len(msg) == 0 {
		return
	}
	l.Now = time.Now()
	l.FileName = fileName
	l.Line = line
	l.LogLevel = level
	l.Message = append([]any{}, msg...)
}

// String 生成字符串
func (l *LogData) String() string {
	if l.Message == nil || len(l.Message) == 0 {
		return ""
	}

	logList := make([]string, 0)

	if !cond.IsTimeEmpty(l.Now) {
		logList = append(logList, l.Now.Format("2006/01/02 15:04:05"))
	}

	if l.LogLevel > 0 {
		logList = append(logList, l.LogLevel.GetName())
	}

	if l.LogCommData.LogId != "" {
		logList = append(logList, l.LogCommData.LogId)
	}

	if l.Env != "" {
		logList = append(logList, l.Env.String())
	}

	if l.FileName != "" {
		fileNameTemp := filepath.Base(l.FileName)
		if fileNameTemp != "" {
			logList = append(logList, fmt.Sprintf("[%s:%d]", fileNameTemp, l.Line))
		}
	}

	if l.Path != "" || l.Method != "" {
		if l.Path != "" && l.Method != "" {
			logList = append(logList, fmt.Sprintf("[%s %s]", l.Path, l.Method))
		} else if l.Path != "" {
			logList = append(logList, fmt.Sprintf("[%s]", l.Path))
		} else if l.Path != "" && l.Method != "" {
			logList = append(logList, fmt.Sprintf("[%s]", l.Method))
		}
	}

	if len(l.Extend) > 0 {
		logList = append(logList, fmt.Sprintf("[%s]", param.HttpBuildQuery(l.Extend)))
	}

	if len(l.UserId) > 0 {
		logList = append(logList, fmt.Sprintf("[%s]", l.UserId))
	}

	logList = append(logList, fmt.Sprintf("%s", utils.Join(l.Message, " ")))

	minTime := l.Now.Sub(l.CreateTime).Milliseconds()
	if minTime > 0 {
		logList = append(logList, fmt.Sprintf("[%dms]", minTime))
	}

	return utils.Join(logList, " ")
}
