package logs

import (
	"context"
)

// ctxLogger 自定义日志的使用方式
type ctxLogger struct {
	ctx         *context.Context //存储当前上下文
	logExecute  LogExecute
	logCommData *LogCommData
	logLevel    LogLevel
	callerSkip  int
}

// NewCtxLogger 例子，一个完整的日志，需要实现如下方法
func NewCtxLogger(ctx context.Context, level LogLevel, logFunc LogExecute, logCommData ...*LogCommData) (ILogger, context.Context) {
	cLogger := getOneLogger(ctx)
	paramLogger := &ctxLogger{
		ctx:        &ctx,
		logExecute: logFunc,
		logLevel:   level,
	}
	if logCommData != nil && len(logCommData) > 0 {
		paramLogger.logCommData = logCommData[0]
	}

	cLogger.buildLogger(ctx, paramLogger)
	newCtx := setLoggerToContext(ctx, cLogger)

	return cLogger, newCtx
}

// SetCallerSkip 设置文件忽略
func (x *ctxLogger) SetCallerSkip(skip int) *ctxLogger {
	x.callerSkip = skip
	return x
}

// WithLogExecute 自定义方法
func (x *ctxLogger) WithLogExecute(logExec LogExecute) *ctxLogger {
	if logExec != nil {
		x.logExecute = logExec
	}
	return x
}

// Debug Debug
func (x *ctxLogger) Debug(v ...interface{}) {
	if x.Level() > DEBUG {
		return
	}
	x.printlnComm(DEBUG, v...)
}

// Error Error
func (x *ctxLogger) Error(v ...interface{}) {
	if x.Level() > ERROR {
		return
	}
	x.printlnComm(ERROR, v...)
}

// Info Info
func (x *ctxLogger) Info(v ...interface{}) {
	if x.Level() > INFO {
		return
	}
	x.printlnComm(INFO, v...)
}

// Warn xx
func (x *ctxLogger) Warn(v ...interface{}) {
	if x.Level() > WARNING {
		return
	}
	x.printlnComm(WARNING, v...)
}

// Level xx
func (x *ctxLogger) Level() LogLevel { return x.logLevel }

// SetLevel SetLevel
func (x *ctxLogger) SetLevel(l LogLevel) {
	x.logLevel = l
}
