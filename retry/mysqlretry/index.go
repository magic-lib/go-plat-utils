// Package mysqlretry 重试器
package mysqlretry

import (
	"database/sql"
	"fmt"
	"github.com/magic-lib/go-plat-mysql/sqlstatement"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/samber/lo"
	"time"
)

const (
	retryStatusPending = "pending"
	retryStatusFailure = "failure"
	retryStatusSuccess = "success"

	defaultTableName = "retry_records"
	maxRetries       = 3
	defaultInterval  = 60
)

var (
	registorCallback = make(map[string]func(param []any) (any, error))
)

// retryRecord 定义请求记录结构体
type retryRecord struct {
	Id             int64     `json:"id"`
	Namespace      string    `json:"namespace"`  // 命名空间，用于区分不同的业务
	RetryType      string    `json:"retry_type"` // 重试类型，函数里定义特殊的执行方法
	Param          string    `json:"param"`
	Errors         string    `json:"errors"`
	Response       string    `json:"response"`
	Retries        int       `json:"retries"`
	MaxRetries     int       `json:"max_retries"`
	IntervalSecond int64     `json:"interval_second"` // 重试间隔时间，单位秒
	NextRetry      time.Time `json:"next_retry"`
	Status         string    `json:"status"`
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
}

type RetryConfig struct {
	SqlExec   func(query string, args ...any) error
	SqlQuery  func(query string, args ...any) (*sql.Rows, error)
	TableName string `json:"table_name"`
	Namespace string `json:"namespace"`
}
type RetryRecord struct {
	RetryType  string        `json:"retry_type"`
	Param      []any         `json:"param"`
	MaxRetries int           `json:"max_retries"`
	Interval   time.Duration `json:"interval"`
	Execute    func(param []any) (any, error)
}

func NewMysqlRetry(rc *RetryConfig) (*RetryConfig, error) {
	if rc.Namespace == "" {
		return nil, fmt.Errorf("namespace is empty")
	}

	if rc.TableName == "" {
		rc.TableName = defaultTableName
	}
	err := rc.createRetryTable()
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func getCallbackKey(namespace, retryType string) string {
	if namespace == "" || retryType == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", namespace, retryType)
}

func Register(namespace, retryType string, fun func(param []any) (any, error)) error {
	keyCache := getCallbackKey(namespace, retryType)
	if keyCache == "" {
		return fmt.Errorf("namespace or retryType is empty")
	}
	registorCallback[keyCache] = fun
	return nil
}
func (rc *RetryConfig) getCallback(namespace, retryType string) (func(param []any) (any, error), error) {
	keyCache := getCallbackKey(namespace, retryType)
	if keyCache == "" {
		return nil, fmt.Errorf("namespace or retryType is empty")
	}
	if f, ok := registorCallback[keyCache]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("find no function")
}

// 创建请求记录表
func (rc *RetryConfig) createRetryTable() error {
	creatSql := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id bigint AUTO_INCREMENT PRIMARY KEY,
            namespace VARCHAR(100) NOT NULL,
            retry_type VARCHAR(100) NOT NULL,
            param TEXT NOT NULL,
            errors TEXT,
            response TEXT,
            retries INT DEFAULT 0,
            max_retries INT DEFAULT %d,
            interval_second bigint DEFAULT %d,
            next_retry DATETIME NOT NULL,
            status VARCHAR(20) DEFAULT '%s',
            create_time DATETIME,
            update_time DATETIME
        )
    `, rc.TableName, maxRetries, defaultInterval, retryStatusPending)
	return rc.SqlExec(creatSql)
}

// AddRecord 插入请求记录到数据库
func (rc *RetryConfig) AddRecord(r *RetryRecord) error {
	if r.Interval == 0 {
		r.Interval = defaultInterval * time.Second
	}
	if r.MaxRetries == 0 {
		r.MaxRetries = maxRetries
	}
	if r.RetryType == "" {
		return fmt.Errorf("retryType is empty")
	}

	f, err := rc.getCallback(rc.Namespace, r.RetryType)
	if err != nil || f == nil {
		if r.Execute == nil {
			if err != nil {
				return err
			}
			err = fmt.Errorf("has no execute")
			return err
		}
		Register(rc.Namespace, r.RetryType, r.Execute)
	}

	nextRetry := time.Now().Add(r.Interval)
	query, data, err := sqlstatement.NewSqlStruct(sqlstatement.SetTableName(rc.TableName), sqlstatement.SetStructData(retryRecord{})).InsertSql(&retryRecord{
		Namespace:      rc.Namespace,
		RetryType:      r.RetryType,
		Param:          conv.String(r.Param),
		MaxRetries:     r.MaxRetries,
		IntervalSecond: int64(r.Interval.Seconds()),
		NextRetry:      nextRetry,
		Status:         retryStatusPending,
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	})
	if err != nil {
		return err
	}

	return rc.SqlExec(query, data...)
}

// 扫描需要重试的请求记录
func (rc *RetryConfig) scanRequestRecords() ([]map[string]string, error) {
	whereCond := sqlstatement.LogicCondition{
		Conditions: []any{
			sqlstatement.Condition{
				Field:    "namespace",
				Operator: "=",
				Value:    rc.Namespace,
			},
			sqlstatement.Condition{
				Field:    "status",
				Operator: "=",
				Value:    retryStatusPending,
			},
			sqlstatement.Condition{
				Field: "`retries`<`max_retries`",
			},
			sqlstatement.Condition{
				Field:    "next_retry",
				Operator: "<=",
				Value:    conv.String(time.Now()),
			},
		},
		Operator: "AND",
	}
	st := sqlstatement.Statement{}
	whereQuery, data := st.GenerateWhereClause(whereCond)
	sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE %s order by next_retry asc limit 50", rc.TableName, whereQuery)

	rows, err := rc.SqlQuery(sqlStr, data...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]string
	// 为每一行创建一个 []any 数组
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// 遍历结果集
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}
		row := make(map[string]string)
		lo.ForEach(columns, func(item string, index int) {
			row[item] = conv.String(values[index])
		})
		result = append(result, row)
	}

	return result, nil
}

// 执行失败
func (rc *RetryConfig) failExec(item *retryRecord, err error) error {
	errList := make([]string, 0)
	if item.Errors != "" {
		_ = conv.Unmarshal(item.Errors, &errList)
	}

	failSql := fmt.Sprintf(`
        UPDATE %s
        SET errors = ?, update_time=?, status=IF(retries+1 >= max_retries, '%s', '%s'), retries = retries + 1, next_retry = '%s' + INTERVAL %d SECOND
        WHERE id = ?
    `, rc.TableName, retryStatusFailure, retryStatusPending, conv.String(time.Now()), item.IntervalSecond)

	errList = append(errList, err.Error())

	return rc.SqlExec(failSql, conv.String(errList), conv.String(time.Now()), item.Id)
}

// 执行成功
func (rc *RetryConfig) successExec(item *retryRecord, resp string) error {
	successSql := fmt.Sprintf(`
        UPDATE %s
        SET response = ?, retries = retries + 1, update_time=?, status="%s"
        WHERE id = ?
    `, rc.TableName, retryStatusSuccess)
	return rc.SqlExec(successSql, resp, conv.String(time.Now()), item.Id)
}

// 执行请求
func (rc *RetryConfig) executeRequest() error {
	list, err := rc.scanRequestRecords()
	if err != nil {
		return err
	}
	dataList := make([]*retryRecord, 0)
	err = conv.Unmarshal(list, &dataList)
	if err != nil {
		return err
	}

	lo.ForEach(dataList, func(item *retryRecord, i int) {
		paramList := make([]any, 0)
		err := conv.Unmarshal(item.Param, &paramList)
		if err != nil {
			_ = rc.failExec(item, err)
			return
		}
		f, err := rc.getCallback(item.Namespace, item.RetryType)
		if err != nil {
			_ = rc.failExec(item, err)
			return
		}
		response, err := f(paramList)
		if err != nil {
			_ = rc.failExec(item, err)
			return
		}
		_ = rc.successExec(item, conv.String(response))
	})
	return nil
}

func (rc *RetryConfig) Start() {
	goroutines.GoAsync(func(param ...any) {
		// 定时扫描需要重试的请求记录
		ticker := time.NewTicker(defaultInterval * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := rc.executeRequest()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}, nil)
}
