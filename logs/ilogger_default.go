package logs

import (
	"context"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/goroutines"
)

type Config struct {
	DefaultLogger ILogger
	LoggerCtxName string //logger在context里的名字
	LogLevel      LogLevel
}

// GetConfig 获取默认配置
func GetConfig() *Config {
	return defaultConfig
}

// SetConfig 设置默认日志，不能包含ctx，不然全局唯一会有问题
func SetConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	configTemp := GetConfig()
	if !cond.IsNil(cfg.DefaultLogger) {
		configTemp.DefaultLogger = cfg.DefaultLogger
	}
	if cfg.LogLevel > 0 {
		configTemp.LogLevel = cfg.LogLevel
	}

	// 初始化设置
	if configTemp.LogLevel > 0 {
		DefaultLogger().SetLevel(configTemp.LogLevel)
	}
}

// DefaultLogger 获取系统默认的日志
func DefaultLogger() ILogger {
	configTemp := GetConfig()
	if cond.IsNil(configTemp.DefaultLogger) {
		logger := NewPrintLogger(configTemp.LogLevel)
		logger.SetCallerSkip(3)
		configTemp.DefaultLogger = logger
	}
	return configTemp.DefaultLogger
}

// CtxLogger 获取系统默认的日志
func CtxLogger(ctx context.Context) ILogger {
	// 先从ctx里获取是否存在logger
	if ctx != nil {
		cLogger := getCtxLoggerFromContext(ctx)
		if cLogger != nil {
			goroutines.SetContext(&ctx)
			return cLogger
		}
	}

	// 再从全局保存的context里获取
	oneCtxPtr, _, _ := goroutines.GetContext()
	if oneCtxPtr != nil {
		cLogger := getCtxLoggerFromContext(*oneCtxPtr)
		if cLogger != nil {
			return cLogger
		}
	}
	// 再新建一个logger
	if ctx != nil {
		cLoggerNew, _ := NewCtxLogger(ctx, GetConfig().LogLevel, nil)
		return cLoggerNew
	}
	return DefaultLogger()
}
