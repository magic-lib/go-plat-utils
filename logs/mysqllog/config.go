package mysqllog

import (
	"database/sql"
	"fmt"
	"github.com/magic-lib/go-plat-utils/logs"
	"github.com/magic-lib/go-plat-utils/utils"
	"strings"
	"time"
)

// ExtendField 扩展字段定义
type ExtendField struct {
	Name    string // 字段名
	DBType  string // MySQL 字段类型，如 VARCHAR(64)、INT、TEXT 等
	Comment string // 字段注释
}

// Config MySQL日志配置
type Config struct {
	// DB 数据库连接（必填）
	DB *sql.DB

	// TablePrefix 表名前缀，默认为 "log"
	TablePrefix string

	// RetentionDays 日志保留天数，自动清除多少天以前的表，默认30天
	RetentionDays int

	// PreCreateDays 提前创建未来多少天的表，默认为2天
	PreCreateDays int

	// ExtendFields 扩展字段配置，会在每个日志表中新增这些字段
	ExtendFields []ExtendField

	// CleanCronSpec 自动清理cron表达式，默认每天凌晨2点 "0 2 * * *"
	CleanCronSpec string

	// CreateCronSpec 自动创建表的cron表达式，默认每天凌晨0点 "0 0 * * *"
	CreateCronSpec string

	// LogLevel 最低日志级别，低于此级别的日志不会写入数据库
	LogLevel logs.LogLevel

	// MaxRetry 写入失败最大重试次数，默认0（不重试）
	MaxRetry int

	// BatchSize 批量写入大小，默认1（逐条写入）
	BatchSize int

	// SubBatchSize 子批次插入大小，flushBatch 时按此大小拆分为多条 INSERT，每条包含多行数据。
	// 默认5，设为1表示逐条插入（不合并多行），设为0使用默认值5。
	SubBatchSize int
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		TablePrefix:    "log",
		RetentionDays:  10,
		PreCreateDays:  2,
		CleanCronSpec:  "0 0 2 * * *",
		CreateCronSpec: "0 0 0 * * *",
		LogLevel:       logs.INFO, // INFO级别
		MaxRetry:       0,
		BatchSize:      1,
		SubBatchSize:   5,
	}
}

// Validate 校验配置
func (c *Config) Validate() error {
	if c.DB == nil {
		return fmt.Errorf("mysqllog: DB connection is required")
	}
	newConfig := DefaultConfig()
	if c.TablePrefix == "" {
		c.TablePrefix = newConfig.TablePrefix
	}
	if c.RetentionDays <= 0 {
		c.RetentionDays = newConfig.RetentionDays
	}
	if c.PreCreateDays < 0 {
		c.PreCreateDays = newConfig.PreCreateDays
	}
	if c.CleanCronSpec == "" {
		c.CleanCronSpec = newConfig.CleanCronSpec
	}
	if c.CreateCronSpec == "" {
		c.CreateCronSpec = newConfig.CreateCronSpec
	}
	if c.LogLevel <= 0 {
		c.LogLevel = newConfig.LogLevel
	}
	if c.SubBatchSize <= 0 {
		c.SubBatchSize = newConfig.SubBatchSize
	}
	return nil
}

// getTableName 根据日期获取表名，格式: {prefix}_20260529
func (c *Config) getTableName(t time.Time) string {
	return fmt.Sprintf("%s_%s", c.TablePrefix, t.Format("20060102"))
}

// getTableNameByOffset 根据天数偏移获取表名，0=今天，1=明天，-1=昨天
func (c *Config) getTableNameByOffset(offset int) string {
	return c.getTableName(time.Now().AddDate(0, 0, offset))
}

func (c *Config) getTableFieldList() []string {
	fields := c.getTableBasicFieldList()
	extendFields := c.getTableExtendFieldList()
	for _, f := range extendFields {
		fields = utils.AppendUniq(fields, f.Name)
	}
	return fields
}
func (c *Config) getTableBasicFieldList() []string {
	fields := []string{
		"id", "log_id", "log_key", "log_level", "log_time", "create_time", "ip",
		"env", "path", "method", "file_name", "line",
		"message", "cost_duration", "extends",
	}
	return fields
}
func (c *Config) getTableExtendFieldList() []ExtendField {
	fields := make([]ExtendField, 0)
	for _, f := range c.ExtendFields {
		f.Name = strings.TrimSpace(f.Name)
		f.Name = strings.ReplaceAll(f.Name, "`", "")
		f.Name = strings.ToLower(f.Name)
		if f.Name == "" {
			continue
		}
		if f.DBType == "" {
			f.DBType = "VARCHAR(256)"
		}
		fields = append(fields, f)
	}
	return fields
}

// buildCreateTableSQL 构建建表SQL
func (c *Config) buildCreateTableSQL(tableName string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n", tableName))
	sb.WriteString("  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,\n")
	sb.WriteString("  `log_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '日志ID',\n")
	sb.WriteString("  `log_key` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '日志关键标识，比如用户Id等关键信息',\n")
	sb.WriteString("  `log_level` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '日志级别',\n")
	sb.WriteString("  `log_time` DATETIME(3) NOT NULL COMMENT '日志记录时间',\n")
	sb.WriteString("  `create_time` DATETIME NOT NULL COMMENT '请求创建时间',\n")
	sb.WriteString("  `ip` VARCHAR(50) NOT NULL DEFAULT '' COMMENT 'Ip地址',\n")
	sb.WriteString("  `env` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '环境',\n")
	sb.WriteString("  `path` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '请求路径',\n")
	sb.WriteString("  `method` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '请求方法',\n")
	sb.WriteString("  `file_name` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '文件名',\n")
	sb.WriteString("  `line` INT NOT NULL DEFAULT 0 COMMENT '行号',\n")
	sb.WriteString("  `message` TEXT COMMENT '日志内容',\n")
	sb.WriteString("  `cost_duration` BIGINT NOT NULL DEFAULT 0 COMMENT '耗时(毫秒)',\n")
	sb.WriteString("  `extends` TEXT COMMENT '扩展字段',\n")

	// 扩展字段
	extendFields := c.getTableExtendFieldList()
	for _, f := range extendFields {
		comment := ""
		if f.Comment != "" {
			comment = fmt.Sprintf(" COMMENT '%s'", f.Comment)
		}
		sb.WriteString(fmt.Sprintf("  `%s` %s NULL%s,\n", f.Name, f.DBType, comment))
	}

	sb.WriteString(fmt.Sprintf("  INDEX `idx_log_level` (`log_level`),\n"))
	sb.WriteString(fmt.Sprintf("  INDEX `idx_log_time` (`log_time`),\n"))
	sb.WriteString(fmt.Sprintf("  INDEX `idx_log_key` (`log_key`)\n"))
	sb.WriteString(") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日志表';")

	return sb.String()
}

// buildInsertSQL 构建插入SQL
func (c *Config) buildInsertSQL(tableName string) string {
	fields := c.getTableFieldList()
	return c.buildInsertSQLByFields(tableName, fields)
}
func (c *Config) buildInsertSQLByFields(tableName string, fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	sqlStr, err := c.getInsertSQLByFields(tableName, fields)
	if err != nil {
		return ""
	}
	return sqlStr
}
func (c *Config) getInsertSQLByFields(tableName string, fields []string) (string, error) {
	if len(fields) == 0 {
		return "", fmt.Errorf("fields cannot be empty")
	}
	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		tableName,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", ")), nil
}

// buildBatchInsertSQL 构建批量插入SQL，支持一次插入多行数据
// 生成格式: INSERT INTO `table` (col1, col2) VALUES (?, ?), (?, ?), (?, ?)
func (c *Config) buildBatchInsertSQL(tableName string, fields []string, rowCount int) string {
	if len(fields) == 0 || rowCount <= 0 {
		return ""
	}
	oneRow := "(" + strings.Repeat("?, ", len(fields)-1) + "?)"
	rows := make([]string, rowCount)
	for i := range rows {
		rows[i] = oneRow
	}
	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s",
		tableName,
		strings.Join(fields, ", "),
		strings.Join(rows, ", "))
}
