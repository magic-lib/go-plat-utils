package logs

import (
	"context"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/magic-lib/go-plat-utils/utils"
)

func (x *ctxLogger) printlnComm(level LogLevel, msg ...any) {
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
		ctxPtr := x.ctx
		if ctxPtr == nil {
			ctx := context.Background()
			ctxPtr = &ctx
		}
	}
	x.logExecute(*ctxPtr, logNewInfo)
}

// 获取一个logger
func getOneLogger(ctx context.Context) *ctxLogger {
	logInstance := getCtxLoggerFromContext(ctx) //获取原始的
	if logInstance != nil {
		logInstance.ctx = &ctx
		return logInstance
	}
	cLogger := new(ctxLogger)
	cLogger.buildLogger(ctx, &ctxLogger{
		logLevel: GetConfig().LogLevel,
	})
	return cLogger //默认使用新的
}

// 从ctx获取logger
func getCtxLoggerFromContext(ctx context.Context) *ctxLogger {
	if ctx == nil {
		return nil
	}
	if logInfoTemp := ctx.Value(GetConfig().LoggerCtxName); logInfoTemp != nil {
		if logInstance, ok := logInfoTemp.(*ctxLogger); ok {
			return logInstance
		}
	}
	return nil
}

func setLoggerToContext(ctx context.Context, logger *ctxLogger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		//直接拿原始的
		logger = getOneLogger(ctx)
	}
	newCtx := context.WithValue(ctx, GetConfig().LoggerCtxName, logger)
	logger.ctx = &newCtx
	goroutines.SetContext(&newCtx)
	return newCtx
}

func (x *ctxLogger) buildLogger(ctx context.Context, logParam *ctxLogger) {
	if logParam == nil {
		return
	}

	if ctx != nil {
		x.ctx = &ctx
	}
	if logParam.logLevel >= DEBUG {
		x.logLevel = logParam.logLevel
	}
	if logParam.logExecute != nil {
		x.logExecute = logParam.logExecute
	} else {
		if x.logExecute == nil {
			x.logExecute = func(ctx context.Context, logInfo *LogData) {
				if logInfo == nil || len(logInfo.Message) == 0 {
					return
				}
				//默认是控制台打印
				logger := NewPrintLogger(logInfo.LogLevel, &logInfo.LogCommData)
				logger.SetCallerSkip(6)
				Logger(logger, logInfo.LogLevel, logInfo.Message...)
			}
		}
	}
	if logParam.callerSkip > 0 {
		x.callerSkip = logParam.callerSkip
	}

	if logParam.logCommData != nil {
		x.logCommData = logParam.logCommData
	}

	if x.logCommData == nil {
		x.logCommData = new(LogCommData)
	}
	x.logCommData.Init()
}
