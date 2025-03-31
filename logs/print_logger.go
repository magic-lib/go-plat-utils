package logs

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/magic-lib/go-plat-utils/utils"
)

// printLogger 自定义日志的使用方式
type printLogger struct {
	logCommData *LogCommData
	logExecute  LogExecute
	logLevel    LogLevel
	callerSkip  int
}

// NewPrintLogger 例子，一个完整的日志，需要实现如下方法
func NewPrintLogger(level LogLevel, logCommData ...*LogCommData) *printLogger {
	printLoggerInstance := &printLogger{
		callerSkip: 2,
	}

	if level >= DEBUG {
		printLoggerInstance.SetLevel(level)
	}

	if logCommData != nil && len(logCommData) > 0 {
		printLoggerInstance.logCommData = logCommData[0]
	}

	printLoggerInstance.logExecute = defaultPrintLogExecute

	return printLoggerInstance
}

// SetCallerSkip 设置文件忽略
func (x *printLogger) SetCallerSkip(skip int) *printLogger {
	x.callerSkip = skip
	return x
}

// WithLogExecute 自定义方法
func (x *printLogger) WithLogExecute(logExec LogExecute) *printLogger {
	if logExec != nil {
		x.logExecute = logExec
	}
	return x
}

// defaultPrintLogExecute 默认的处理方法，自定义的话可仿照着写
func defaultPrintLogExecute(ctx context.Context, logInfo *LogData) {
	if logInfo == nil || logInfo.Message == nil || len(logInfo.Message) == 0 {
		return
	}

	msgStr := logInfo.String()
	if msgStr == "" {
		return
	}

	if logInfo.LogLevel < ERROR {
		fmt.Println(msgStr)
		return
	}

	//如果是错误，则直接红色打印
	slantedRed := color.New(color.FgRed, color.Bold)
	_, err := slantedRed.Println(msgStr)
	if err == nil {
		return
	}
}

func (x *printLogger) printlnComm(level LogLevel, msg ...interface{}) {
	if len(msg) == 0 {
		return
	}

	logNewInfo := NewLogData(x.logCommData)

	fileName := ""
	line := 0
	file, _ := utils.SpecifyContext(x.callerSkip)
	if file != nil {
		fileName = file.FileName
		line = file.Line
	}
	logNewInfo.AddMessage(level, fileName, line, msg...)

	ctxPtr, _, _ := goroutines.GetContext()
	if ctxPtr == nil {
		ctx := context.Background()
		ctxPtr = &ctx
	}
	x.logExecute(*ctxPtr, logNewInfo)
}

// Debug Debug
func (x *printLogger) Debug(v ...interface{}) {
	if x.Level() > DEBUG {
		return
	}
	x.printlnComm(DEBUG, v...)
}

// Error Error
func (x *printLogger) Error(v ...interface{}) {
	if x.Level() > ERROR {
		return
	}
	x.printlnComm(ERROR, v...)
}

// Info Info
func (x *printLogger) Info(v ...interface{}) {
	if x.Level() > INFO {
		return
	}
	x.printlnComm(INFO, v...)
}

// Warn xx
func (x *printLogger) Warn(v ...interface{}) {
	if x.Level() > WARNING {
		return
	}
	x.printlnComm(WARNING, v...)
}

// Level xx
func (x *printLogger) Level() LogLevel { return x.logLevel }

// SetLevel SetLevel
func (x *printLogger) SetLevel(l LogLevel) {
	x.logLevel = l
}
