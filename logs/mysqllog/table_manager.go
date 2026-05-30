package mysqllog

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	"strings"
	"sync"
	"time"

	"github.com/magic-lib/go-plat-utils/logs"
	"github.com/robfig/cron/v3"
)

// tableManager 表管理器，负责表的创建、清理、定时调度
type tableManager struct {
	cfg    *Config
	cron   *cron.Cron
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// 已创建表的缓存，避免重复创建
	createdTables map[string][]string
}

// newTableManager 创建表管理器
func newTableManager(cfg *Config) *tableManager {
	ctx, cancel := context.WithCancel(context.Background())
	tm := &tableManager{
		cfg:           cfg,
		cron:          cron.New(cron.WithSeconds()),
		ctx:           ctx,
		cancel:        cancel,
		createdTables: make(map[string][]string),
	}
	return tm
}

// Init 初始化：创建今天+未来N天的表，清理过期表，启动定时任务
func (tm *tableManager) Init() error {
	// 1. 创建今天和未来 PreCreateDays 天的表
	for i := 0; i <= tm.cfg.PreCreateDays; i++ {
		tableName := tm.cfg.getTableNameByOffset(i)
		if err := tm.createTableIfNotExists(tableName); err != nil {
			return fmt.Errorf("mysqllog: create table %s failed: %w", tableName, err)
		}
	}

	// 2. 清理过期表
	if err := tm.cleanOldTables(); err != nil {
		// 清理失败不阻塞初始化，只打印日志
		logs.DefaultLogger().Warn("mysqllog: clean old tables failed: ", err.Error())
	}

	// 3. 启动定时任务
	if err := tm.startCronJobs(); err != nil {
		return fmt.Errorf("mysqllog: start cron jobs failed: %w", err)
	}

	return nil
}

// createTableIfNotExists 创建表（如果不存在）
func (tm *tableManager) createTableIfNotExists(tableName string) error {
	tm.mu.RLock()
	if _, ok := tm.createdTables[tableName]; ok {
		tm.mu.RUnlock()
		return nil
	}
	tm.mu.RUnlock()

	sql := tm.cfg.buildCreateTableSQL(tableName)
	_, err := tm.cfg.DB.ExecContext(tm.ctx, sql)
	if err != nil {
		return fmt.Errorf("create table %s: %w", tableName, err)
	}

	// 查询表的字段列表
	fieldList, err := tm.getTableColumns(tableName)
	if err != nil {
		logs.DefaultLogger().Warn("mysqllog: get table columns failed: ", tableName, ", err: ", err.Error())
		// 查询失败不影响建表，使用空列表
		fieldList = tm.cfg.getTableFieldList()
	}

	tm.mu.Lock()
	tm.createdTables[tableName] = fieldList
	tm.mu.Unlock()

	logs.DefaultLogger().Info("mysqllog: table created: ", tableName)
	return nil
}

// getTableColumns 通过 SQL 查询表的字段列表
func (tm *tableManager) getTableColumns(tableName string) ([]string, error) {
	query := "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION"

	rows, err := tm.cfg.DB.QueryContext(tm.ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("query columns failed: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var columns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			continue
		}
		columns = append(columns, columnName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan rows failed: %w", err)
	}

	return columns, nil
}

// cleanOldTables 清理过期表
func (tm *tableManager) cleanOldTables() error {
	cutoffTime := time.Now().AddDate(0, 0, -tm.cfg.RetentionDays)
	cutoffTableName := tm.cfg.getTableName(cutoffTime)

	// 查询所有匹配前缀的表
	rows, err := tm.cfg.DB.QueryContext(tm.ctx,
		fmt.Sprintf("SHOW TABLES LIKE '%s_%%'", tm.cfg.TablePrefix))
	if err != nil {
		return fmt.Errorf("show tables failed: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var dropTables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		// 表名小于截止日期表名，说明已过期
		if tableName < cutoffTableName {
			dropTables = append(dropTables, tableName)
		}
	}

	if len(dropTables) == 0 {
		return nil
	}

	// 批量删除过期表
	for _, tableName := range dropTables {
		sql := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)
		if _, err := tm.cfg.DB.ExecContext(tm.ctx, sql); err != nil {
			logs.DefaultLogger().Warn("mysqllog: drop table failed: ", tableName, ", err: ", err.Error())
			continue
		}
		// 从缓存中移除
		tm.mu.Lock()
		delete(tm.createdTables, tableName)
		tm.mu.Unlock()

		logs.DefaultLogger().Info("mysqllog: table dropped: ", tableName)
	}

	return nil
}

// createFutureTables 创建未来几天的表
func (tm *tableManager) createFutureTables() error {
	for i := 1; i <= tm.cfg.PreCreateDays; i++ {
		tableName := tm.cfg.getTableNameByOffset(i)
		if err := tm.createTableIfNotExists(tableName); err != nil {
			logs.DefaultLogger().Warn("mysqllog: pre-create table failed: ", tableName, ", err: ", err.Error())
		}
	}
	return nil
}

// ensureTodayTable 确保今天的表存在
func (tm *tableManager) ensureTodayTable() error {
	tableName := tm.cfg.getTableName(time.Now())
	return tm.createTableIfNotExists(tableName)
}

// startCronJobs 启动定时任务
func (tm *tableManager) startCronJobs() error {
	// 定时创建未来表
	_, err := tm.cron.AddFunc(tm.cfg.CreateCronSpec, func() {
		_ = tm.createFutureTables()
		_ = tm.ensureTodayTable()
	})
	if err != nil {
		return fmt.Errorf("add create cron job failed: %w", err)
	}

	// 定时清理过期表
	_, err = tm.cron.AddFunc(tm.cfg.CleanCronSpec, func() {
		if err := tm.cleanOldTables(); err != nil {
			logs.DefaultLogger().Warn("mysqllog: cron clean old tables failed: ", err.Error())
		}
	})
	if err != nil {
		return fmt.Errorf("add clean cron job failed: %w", err)
	}

	tm.cron.Start()
	logs.DefaultLogger().Info("mysqllog: cron jobs started, create: ", tm.cfg.CreateCronSpec, ", clean: ", tm.cfg.CleanCronSpec)
	return nil
}

// Stop 停止定时任务
func (tm *tableManager) Stop() {
	tm.cron.Stop()
	tm.cancel()
}

// getInsertSQL 获取带表名的插入SQL
func (tm *tableManager) getInsertSQL() string {
	tableName := tm.cfg.getTableName(time.Now())
	if tableFields, ok := tm.createdTables[tableName]; ok {
		return tm.cfg.buildInsertSQLByFields(tableName, tableFields)
	}

	return tm.cfg.buildInsertSQL(tableName)
}

func (tm *tableManager) getTableExtendFieldList() []string {
	tableName := tm.cfg.getTableName(time.Now())
	allFields, ok := tm.createdTables[tableName]
	if !ok {
		extendFields := tm.cfg.getTableExtendFieldList()
		extendFieldList := lo.Map(extendFields, func(field ExtendField, _ int) string {
			return field.Name
		})
		return extendFieldList
	}
	basicFields := tm.cfg.getTableBasicFieldList()
	extendFieldList, _ := lo.Difference(allFields, basicFields)
	return extendFieldList
}

// buildInsertArgs 构建插入参数
func (tm *tableManager) buildInsertArgs(logData *logs.LogData) []any {
	if logData == nil {
		return nil
	}
	logData.Init()

	basicFields := tm.cfg.getTableBasicFieldList()
	extendFields := tm.getTableExtendFieldList()

	allLogMap := make(map[string]any)
	_ = conv.Unmarshal(logData, &allLogMap)
	extendMap := make(map[string]any)
	_ = conv.Unmarshal(logData.Extends, &extendMap)
	if len(extendMap) > 0 {
		allLogMap = lo.Assign(allLogMap, extendMap)
	}
	if msgObj, ok := allLogMap["message"]; ok {
		if msgList, ok := msgObj.([]any); ok {
			if len(msgList) == 1 {
				allLogMap["message"] = msgList[0]
			} else {
				hasStr := false
				lo.ForEachWhile(msgList, func(item any, _ int) bool {
					if _, ok := item.(string); ok {
						hasStr = true
						return false
					}
					return true
				})
				if hasStr {
					allLogMap["message"] = strings.Join(lo.Map(msgList, func(item any, _ int) string {
						return conv.String(item)
					}), "")
				}
			}
		}
	}
	if _, ok := allLogMap["extends"]; !ok {
		allLogMap["extends"] = ""
	}

	allLogMap["id"] = 0 // 设置 id 为 0，让数据库自增

	allLogData := make([]any, 0)
	lo.ForEach(basicFields, func(field string, _ int) {
		if val, ok := allLogMap[field]; ok {
			allLogData = append(allLogData, conv.String(val))
		} else {
			allLogData = append(allLogData, "")
		}
	})
	lo.ForEach(extendFields, func(field string, _ int) {
		if val, ok := allLogMap[field]; ok {
			allLogData = append(allLogData, conv.String(val))
		} else {
			allLogData = append(allLogData, nil)
		}
	})

	return allLogData
}
