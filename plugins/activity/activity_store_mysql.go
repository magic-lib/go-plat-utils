package activity

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SQLActivityStore 基于标准库 database/sql 的存储实现。
// DB 由调用方通过具体驱动（sqlite3/mysql 等）打开后注入，本包不引入驱动。
type SQLActivityStore struct {
	DB                *sql.DB
	activityTableName string
}

// Migrate 建表（SQLite 语法；MySQL 见下方说明）。
func (s *SQLActivityStore) Migrate(ctx context.Context, tableName string) error {
	s.activityTableName = tableName
	_, err := s.DB.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id      INTEGER PRIMARY KEY AUTOINCREMENT,
	type         VARCHAR(191) NOT NULL UNIQUE,
	activity_type VARCHAR(191),
	act_namespace VARCHAR(191),
	act_name     VARCHAR(191),
	arguments    JSON,
	arg_template VARCHAR(191),
	responses    TEXT,
	depends_on   JSON,
	hooks        JSON,
	control      JSON,
	create_time   DATETIME,
	update_time   DATETIME
)`, tableName))
	return err
}

// Save 写入单个 activity（按 type 先查后插/更新，跨库兼容，无方言依赖）。
func (s *SQLActivityStore) Save(ctx context.Context, act *Activity) error {
	if act == nil {
		return fmt.Errorf("activity is nil")
	}
	typeName := act.Type()
	if typeName == "" {
		return fmt.Errorf("activity type is empty, please set ActNamespace/ActName or ActivityType")
	}
	cfg := activityToConfig(act)
	cfg.Type = typeName
	now := time.Now()

	var cnt int
	if err := s.DB.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(1) FROM `%s` WHERE type = ?", s.activityTableName), typeName).Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		_, err := s.DB.ExecContext(ctx,
			fmt.Sprintf(`UPDATE %s SET activity_id=?, activity_type=?, act_namespace=?,
			 act_name=?, arguments=?, arg_template=?, responses=?, depends_on=?, hooks=?,
			 control=?, updated_at=? WHERE type=?`, s.activityTableName),
			cfg.Id, cfg.ActivityType, cfg.ActNamespace, cfg.ActName,
			cfg.Arguments, cfg.ArgTemplate, cfg.Responses, cfg.DependsOn, cfg.Hooks,
			cfg.Control, now, typeName)
		return err
	}
	_, err := s.DB.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (`type`, activity_id, activity_type, act_namespace, act_name, arguments, arg_template, responses, depends_on, hooks, control, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)", s.activityTableName),
		typeName, cfg.Id, cfg.ActivityType, cfg.ActNamespace, cfg.ActName,
		cfg.Arguments, cfg.ArgTemplate, cfg.Responses, cfg.DependsOn, cfg.Hooks,
		cfg.Control, now, now)
	return err
}

// List 拉取全部 activity 并还原。
func (s *SQLActivityStore) List(ctx context.Context) ([]*Activity, error) {
	rows, err := s.DB.QueryContext(ctx,
		fmt.Sprintf(`SELECT type, activity_id, activity_type, act_namespace, act_name,
		        arguments, arg_template, responses, depends_on, hooks, control
		 FROM %s`, s.activityTableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	acts := make([]*Activity, 0)
	for rows.Next() {
		var c ActivityConfig
		if err := rows.Scan(&c.Type, &c.Id, &c.ActivityType, &c.ActNamespace, &c.ActName,
			&c.Arguments, &c.ArgTemplate, &c.Responses, &c.DependsOn, &c.Hooks, &c.Control); err != nil {
			return nil, err
		}
		act, err := c.toActivity()
		if err != nil {
			return nil, fmt.Errorf("restore activity %q failed: %w", c.Type, err)
		}
		acts = append(acts, act)
	}
	return acts, rows.Err()
}
