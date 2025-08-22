package csvfile

import (
	"encoding/csv"
	"fmt"
	"strings"
)

// CsvConfig 用于配置 CSV 读取选项
type CsvConfig[T any] struct {
	Delimiter      rune   // 分隔符，默认为逗号
	HeaderRowIndex int    // 表头位于第几行 -1，表示没有表头， 默认为0
	DataRowIndex   [2]int // 数据起始行索引，默认为 1, 数据结束行索引，0表示读取到文件末尾
	ColumnNumber   int    // 列数，如果有不满足的列，可能存在分隔符的问题，就得进行特殊处理，如果为0，就忽略检查
	TrimSpaces     bool   // 是否去除字段前后空格
	SkipEmptyRows  bool   // 是否跳过空行
}

// DefaultConfig 返回默认配置
func DefaultConfig[T any]() CsvConfig[T] {
	return CsvConfig[T]{
		Delimiter:      ',',
		HeaderRowIndex: 0,
		DataRowIndex:   [2]int{1, 0},
		ColumnNumber:   0,
		TrimSpaces:     true,
		SkipEmptyRows:  true,
	}
}

func (c CsvConfig[T]) checkConfig() error {
	if c.HeaderRowIndex > c.DataRowIndex[0] {
		return fmt.Errorf("数据表头不能大于数据开始行")
	}

	if c.DataRowIndex[0] < 0 {
		c.DataRowIndex[0] = 0
	}

	if c.DataRowIndex[1] < 0 ||
		c.DataRowIndex[1] <= c.DataRowIndex[0] {
		c.DataRowIndex[1] = -1
	}
	return nil
}

// ReadToList 读取 CSV 文件返回数据和表头
func (c CsvConfig[T]) ReadToList(reader *csv.Reader) ([][]string, []string, error) {
	// 读取所有记录
	if c.Delimiter > 0 {
		reader.Comma = c.Delimiter
	}
	err := c.checkConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("配置错误: %w", err)
	}
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("读取记录出错: %w", err)
	}

	headers, err := c.headerFields(records)
	if err != nil {
		return nil, headers, fmt.Errorf("读取表头出错: %w", err)
	}

	result := make([][]string, 0)
	lastRowIndex := c.DataRowIndex[1]
	if lastRowIndex < 0 {
		lastRowIndex = len(records) - 1
	}

	for i := c.DataRowIndex[0]; i <= lastRowIndex; i++ {
		record := records[i]
		if len(record) <= 1 {
			if c.SkipEmptyRows {
				continue
			}
		}
		if c.ColumnNumber > 0 && len(record) != c.ColumnNumber {
			return nil, headers, fmt.Errorf("第 %d 行字段数量与列数不匹配: 列数 %d, 行 %d 个",
				i+1+c.DataRowIndex[0],
				c.ColumnNumber, len(record))
		}
		result = append(result, processFields(record, c.TrimSpaces))
	}

	return result, headers, nil
}

// headerFields 获取表头字段
func (c CsvConfig[T]) headerFields(records [][]string) ([]string, error) {
	if c.HeaderRowIndex >= 0 {
		if c.HeaderRowIndex >= len(records) {
			return nil, fmt.Errorf("表头行索引超出范围")
		}
		headers := records[c.HeaderRowIndex]
		headers = processFields(headers, c.TrimSpaces)
		return headers, nil
	}
	return nil, nil
}

// processFields 处理字段，去除空格等
func processFields(fields []string, trimSpaces bool) []string {
	result := make([]string, len(fields))
	for i, field := range fields {
		if trimSpaces {
			result[i] = strings.TrimSpace(field)
		} else {
			result[i] = field
		}
	}
	return result
}
