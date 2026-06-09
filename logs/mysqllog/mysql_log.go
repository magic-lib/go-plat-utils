package mysqllog

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/samber/lo"
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

var (
	mysqlLoggerTemp logs.ILogger
)

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

	if mysqlLoggerTemp == nil {
		mysqlLoggerTemp = ml
	}

	return ml, nil
}

// DefaultLogger 获取mysql默认的日志
func DefaultLogger() logs.ILogger {
	if mysqlLoggerTemp == nil {
		panic("mysqlLogger is nil, please init first")
	}
	return mysqlLoggerTemp
}

// batchFlushLoop 批量刷新协程
func (ml *mysqlLogger) batchFlushLoop() {
	ticker := time.NewTicker(10 * time.Second)
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

// flushBatch 刷新批量缓冲区，按 SubBatchSize 分组进行多行批量插入
func (ml *mysqlLogger) flushBatch() {
	ml.batchMu.Lock()
	if len(ml.batchBuffer) == 0 {
		ml.batchMu.Unlock()
		return
	}
	batch := ml.batchBuffer
	ml.batchBuffer = make([]*logs.LogData, 0, ml.cfg.BatchSize)
	ml.batchMu.Unlock()

	subSize := ml.cfg.SubBatchSize
	if subSize <= 1 {
		// 逐条插入
		for _, logData := range batch {
			if err := ml.writeLog(logData); err != nil {
				logs.DefaultLogger().Warn("mysqllog: batch write log failed: ", err.Error())
			}
		}
		return
	}

	// 按 subSize 拆分为子批次，每个子批次用一条多行 INSERT 写入
	for i := 0; i < len(batch); i += subSize {
		end := i + subSize
		if end > len(batch) {
			end = len(batch)
		}
		subBatch := batch[i:end]
		if err := ml.writeLogBatch(subBatch); err != nil {
			logs.DefaultLogger().Warn("mysqllog: batch write log failed: ", err.Error())
		}
	}
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
	logData.LogCommData.CreateTime = time.Now() //插入mysql时间
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

// writeLogBatch 批量写入多条日志（单条 INSERT 多行 VALUES）
func (ml *mysqlLogger) writeLogBatch(batch []*logs.LogData) error {
	if len(batch) == 0 {
		return nil
	}

	// 确保表存在
	if err := ml.tm.ensureTodayTable(); err != nil {
		return fmt.Errorf("ensure table: %w", err)
	}

	sqlStr := ml.tm.getBatchInsertSQL(len(batch))
	lo.ForEach(batch, func(logData *logs.LogData, i int) {
		if logData != nil {
			logData.LogCommData.CreateTime = time.Now()
		}
	})
	args := ml.tm.buildBatchInsertArgs(batch)
	if args == nil || sqlStr == "" {
		return nil
	}

	var lastErr error
	maxRetry := ml.cfg.MaxRetry + 1
	for attempt := 0; attempt < maxRetry; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 50 * time.Millisecond)
		}
		_, err := ml.cfg.DB.ExecContext(ml.ctx, sqlStr, args...)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("mysqllog: batch write log failed after %d retries: %w", maxRetry, lastErr)
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

// buildLogData 从消息参数构建 LogData 对象
func (ml *mysqlLogger) buildLogData(level logs.LogLevel, msg ...any) *logs.LogData {
	if len(msg) == 0 {
		return nil
	}

	var logData *logs.LogData

	// 单参数处理
	if len(msg) == 1 {
		logData = ml.parseSingleMessage(msg[0])
	} else {
		logData = ml.parseMultipleMessages(msg)
	}

	// 设置默认值
	if logData != nil {
		ml.setLogDataDefaults(logData, level)
	}

	return logData
}

// parseSingleMessage 解析单个消息参数
func (ml *mysqlLogger) parseSingleMessage(msg any) *logs.LogData {
	if cond.IsNil(msg) {
		return nil
	}

	// 直接是 LogData 类型
	if data, ok := msg.(*logs.LogData); ok && data != nil {
		return ml.ensureLogDataFields(data)
	}
	// LogData 值类型
	if data, ok := msg.(logs.LogData); ok {
		return ml.ensureLogDataFields(&data)
	}

	// 其他类型，尝试解析
	logData := logs.NewLogData(nil)
	if cond.IsJsonMap(conv.String(msg)) {
		_ = conv.Unmarshal(msg, logData)
		extends := make(map[string]any)
		_ = conv.Unmarshal(msg, &extends)
		logData.Extends = extends
	} else {
		logData.Message = []any{msg}
	}

	return logData
}

// parseMultipleMessages 解析多个消息参数
func (ml *mysqlLogger) parseMultipleMessages(msg []any) *logs.LogData {
	allMap := make(map[string]any)
	textMsg := make([]any, 0)

	for _, v := range msg {
		// 尝试作为 JSON map 解析
		if cond.IsJsonMap(conv.String(v)) {
			tempMap := make(map[string]any)
			_ = conv.Unmarshal(v, &tempMap)
			if len(tempMap) > 0 {
				allMap = lo.Assign(allMap, tempMap)
			}
		} else {
			textMsg = append(textMsg, v)
		}
	}
	if len(allMap) == 0 && len(textMsg) == 0 {
		return nil
	}

	logData := logs.NewLogData(nil)

	if len(allMap) > 0 {
		_ = conv.Unmarshal(allMap, logData)
		// allMap 需要去掉包含的所有logData包含的所有字段
		// 获取 LogData 结构体的所有字段名，用于从 allMap 中排除
		logDataMap := make(map[string]any)
		_ = conv.Unmarshal(logData, &logDataMap)
		if len(logDataMap) > 0 {
			for k, field := range logDataMap {
				if oneData, ok := allMap[k]; ok {
					if oneData == field {
						delete(allMap, k)
					}
				}
			}
		}
		logData.Extends = allMap
	}

	if len(textMsg) > 0 {
		if len(logData.Message) > 0 {
			logData.Message = append(logData.Message, textMsg...)
		} else {
			logData.Message = textMsg
		}
	}

	return logData
}

// ensureLogDataFields 确保 LogData 包含必要的字段
func (ml *mysqlLogger) ensureLogDataFields(logData *logs.LogData) *logs.LogData {
	if logData == nil {
		return nil
	}

	// 确保 Message 字段已初始化
	if logData.Message == nil {
		logData.Message = []any{}
	}

	return logData
}

// setLogDataDefaults 设置 LogData 的默认值
func (ml *mysqlLogger) setLogDataDefaults(logData *logs.LogData, level logs.LogLevel) {
	if logData == nil {
		return
	}
	// 设置日志级别
	if logData.LogLevel == 0 {
		logData.LogLevel = level
	}
	now := time.Now()
	// 设置 Now 时间
	if logData.Now.IsZero() {
		logData.Now = now
	}
	// 设置 LogTime
	if logData.LogTime.IsZero() {
		logData.LogTime = now
	}
}

// log 统一日志写入入口
func (ml *mysqlLogger) log(level logs.LogLevel, msg ...any) {
	logData := ml.buildLogData(level, msg...)
	if logData == nil {
		return
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
