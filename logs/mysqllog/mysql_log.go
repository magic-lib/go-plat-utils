package mysqllog

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"sync"
	"time"

	"github.com/magic-lib/go-plat-utils/logs"
)

// mysqlLogger 实现 logs.ILogger 接口的 MySQL 日志记录器
type mysqlLogger struct {
	cfg         *Config
	tm          *tableManager
	logLevel    logs.LogLevel
	mu          sync.Mutex
	batchBuffer []*logs.LogData
	batchMu     sync.Mutex
	flushCh     chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

// New 创建一个新的 MySQL 日志记录器
//
// 使用示例:
//
//	db, _ := sql.Open("mysql", "user:pass@tcp(127.0.0.1:3306)/mydb")
//	cfg := &Config{
//	    DB:            db,
//	    TablePrefix:   "app_log",
//	    ExtendFields: []ExtendField{
//	        {Name: "biz_type", DBType: "VARCHAR(32)", Comment: "业务类型"},
//	        {Name: "trace_id", DBType: "VARCHAR(64)", Comment: "链路追踪ID"},
//	    },
//	}
//	logger, err := New(cfg)
//	if err != nil {
//	    panic(err)
//	}
//	logger.Info("service started")
func New(cfg *Config) (logs.ILogger, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	ml := &mysqlLogger{
		cfg:         cfg,
		logLevel:    cfg.LogLevel,
		ctx:         ctx,
		cancel:      cancel,
		batchBuffer: make([]*logs.LogData, 0, cfg.BatchSize),
		flushCh:     make(chan struct{}, 1),
	}

	// 初始化表管理器
	tm := newTableManager(cfg)
	if err := tm.Init(); err != nil {
		cancel()
		return nil, fmt.Errorf("mysqllog: init table manager failed: %w", err)
	}
	ml.tm = tm

	// 如果批量写入大小大于1，启动批量刷新协程
	if cfg.BatchSize > 1 {
		goroutines.GoAsync(func(params ...any) {
			ml.batchFlushLoop()
		})
	}

	return ml, nil
}

// writeLog 写入单条日志到数据库
func (ml *mysqlLogger) writeLog(logData *logs.LogData) error {
	if logData == nil {
		return nil
	}

	// 确保表存在
	if err := ml.tm.ensureTodayTable(); err != nil {
		return fmt.Errorf("ensure table: %w", err)
	}

	sqlStr := ml.tm.getInsertSQL()
	args := ml.tm.buildInsertArgs(logData)
	if args == nil || sqlStr == "" {
		return nil
	}

	var lastErr error
	maxRetry := ml.cfg.MaxRetry + 1 // 至少执行一次
	for attempt := 0; attempt < maxRetry; attempt++ {
		if attempt > 0 {
			// 重试前短暂等待
			time.Sleep(time.Duration(attempt) * 50 * time.Millisecond)
		}
		_, err := ml.cfg.DB.ExecContext(ml.ctx, sqlStr, args...)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("mysqllog: write log failed after %d retries: %w", maxRetry, lastErr)
}

// batchFlushLoop 批量刷新协程
func (ml *mysqlLogger) batchFlushLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ml.ctx.Done():
			ml.flushBatch()
			return
		case <-ml.flushCh:
			ml.flushBatch()
		case <-ticker.C:
			ml.flushBatch()
		}
	}
}

// flushBatch 刷新批量缓冲区
func (ml *mysqlLogger) flushBatch() {
	ml.batchMu.Lock()
	if len(ml.batchBuffer) == 0 {
		ml.batchMu.Unlock()
		return
	}
	batch := ml.batchBuffer
	ml.batchBuffer = make([]*logs.LogData, 0, ml.cfg.BatchSize)
	ml.batchMu.Unlock()

	for _, logData := range batch {
		if err := ml.writeLog(logData); err != nil {
			logs.DefaultLogger().Warn("mysqllog: batch write log failed: ", err.Error())
		}
	}
}

// addToBatch 添加到批量缓冲区
func (ml *mysqlLogger) addToBatch(logData *logs.LogData) {
	ml.batchMu.Lock()
	ml.batchBuffer = append(ml.batchBuffer, logData)
	shouldFlush := len(ml.batchBuffer) >= ml.cfg.BatchSize
	ml.batchMu.Unlock()

	if shouldFlush {
		select {
		case ml.flushCh <- struct{}{}:
		default:
		}
	}
}

// log 统一日志写入入口
func (ml *mysqlLogger) log(level logs.LogLevel, msg ...any) {
	if len(msg) == 0 {
		return
	}

	var logData *logs.LogData

	// 如果传入的是 *logs.LogData 类型，直接使用
	isMultiLog := false
	if len(msg) == 1 {
		if cond.IsNil(msg[0]) {
			return
		}
		if data, ok := msg[0].(*logs.LogData); ok && data != nil {
			logData = data
		}
		if data, ok := msg[0].(logs.LogData); ok {
			logData = &data
		}
		if logData == nil {
			logData = logs.NewLogData(nil)
			if cond.IsJsonMap(conv.String(msg[0])) {
				_ = conv.Unmarshal(msg[0], logData)
				extends := make(map[string]any)
				_ = conv.Unmarshal(msg[0], &extends)
				logData.Extends = extends
			}
			logData.Message = []any{msg[0]}
		}
	} else {
		isMultiLog = true
		logData = logs.NewLogData(nil)
	}

	if logData != nil {
		if logData.LogLevel == 0 {
			logData.LogLevel = level
		}
		if logData.Now.IsZero() {
			logData.Now = time.Now()
		}
		if logData.LogTime.IsZero() {
			logData.LogTime = time.Now()
		}
		if isMultiLog {
			logData.Message = msg
		}
	}

	// 批量模式
	if ml.cfg.BatchSize > 1 {
		ml.addToBatch(logData)
		return
	}

	// 单条写入模式
	goroutines.GoAsync(func(param ...any) {
		oneLog := param[0].(*logs.LogData)
		if err := ml.writeLog(oneLog); err != nil {
			logs.DefaultLogger().Warn("mysqllog: write log error: ", err.Error())
		}
	}, logData)
}

// Debug 实现 ILogger 接口
func (ml *mysqlLogger) Debug(v ...any) {
	if ml.Level() > logs.DEBUG {
		return
	}
	ml.log(logs.DEBUG, v...)
}

// Info 实现 ILogger 接口
func (ml *mysqlLogger) Info(v ...any) {
	if ml.Level() > logs.INFO {
		return
	}
	ml.log(logs.INFO, v...)
}

// Warn 实现 ILogger 接口
func (ml *mysqlLogger) Warn(v ...any) {
	if ml.Level() > logs.WARNING {
		return
	}
	ml.log(logs.WARNING, v...)
}

// Error 实现 ILogger 接口
func (ml *mysqlLogger) Error(v ...any) {
	if ml.Level() > logs.ERROR {
		return
	}
	ml.log(logs.ERROR, v...)
}

// Level 实现 ILogger 接口
func (ml *mysqlLogger) Level() logs.LogLevel {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return ml.logLevel
}

// SetLevel 实现 ILogger 接口
func (ml *mysqlLogger) SetLevel(l logs.LogLevel) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	if l >= logs.DEBUG {
		ml.logLevel = l
	}
}

// Close 关闭日志记录器，停止定时任务，刷新缓冲区
func (ml *mysqlLogger) Close() {
	ml.tm.Stop()
	ml.cancel()

	if ml.cfg.BatchSize > 1 {
		ml.flushBatch()
	}

	if closer, ok := interface{}(ml.cfg.DB).(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}
