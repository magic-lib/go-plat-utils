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
	LogId      string       `json:"log_id"`            //logId
	LogKey     string       `json:"log_key,omitempty"` //LogKey
	Env        conf.EnvCode `json:"env"`               //env
	Path       string       `json:"path,omitempty"`    //当前请求的地址
	Method     string       `json:"method,omitempty"`  //当前请求的方法
	LogTime    time.Time    `json:"log_time"`          //日志的日志时间
	CreateTime time.Time    `json:"create_time"`       //第一条日志的创建时间
	Extends    any          `json:"extends,omitempty"` //额外的业务参数
}

// LogData 每条单独日志的数据
type LogData struct {
	LogCommData
	Ip           string    `json:"ip"`
	Now          time.Time `json:"now"` //初始化时间
	LogLevel     LogLevel  `json:"log_level"`
	FileName     string    `json:"file_name"` //文件名
	Line         int       `json:"line"`      //行号
	Message      []any     `json:"message"`
	CostDuration int64     `json:"cost_duration"`
}

// Init 初始化
func (l *LogCommData) Init() {
	if cond.IsTimeEmpty(l.CreateTime) {
		l.CreateTime = time.Now()
	}
	if cond.IsTimeEmpty(l.LogTime) {
		l.LogTime = time.Now()
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
	l.Init()
	return l
}

func (l *LogData) Init() {
	l.LogCommData.Init()
	if l.CostDuration == 0 {
		l.CostDuration = time.Now().Sub(l.LogTime).Milliseconds()
	}
	if l.Ip == "" {
		l.Ip = param.MachineCode()
	}
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

	//if len(l.Extends) > 0 {
	//	logList = append(logList, fmt.Sprintf("[%s]", param.HttpBuildQuery(l.Extends)))
	//}

	if len(l.LogKey) > 0 {
		logList = append(logList, fmt.Sprintf("[%s]", l.LogKey))
	}

	logList = append(logList, fmt.Sprintf("%s", utils.Join(l.Message, " ")))

	minTime := l.Now.Sub(l.LogTime).Milliseconds()
	if minTime > 0 {
		logList = append(logList, fmt.Sprintf("[%dms]", minTime))
	}

	return utils.Join(logList, " ")
}
