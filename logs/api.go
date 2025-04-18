package logs

// ILogger is a logger interface
type ILogger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})

	Level() LogLevel
	SetLevel(l LogLevel)
}

// Logger 直接根据等级打印所有日志
func Logger(logger ILogger, l LogLevel, msg ...interface{}) {
	if l <= DEBUG {
		logger.Debug(msg...)
	} else if l <= INFO {
		logger.Info(msg...)
	} else if l <= WARNING {
		logger.Warn(msg...)
	} else if l <= ERROR {
		logger.Error(msg...)
	} else {
		logger.Debug(msg...)
	}
}

// CallSkip 设置文件跳过
type CallSkip interface {
	SetCallerSkip(skip int)
}
