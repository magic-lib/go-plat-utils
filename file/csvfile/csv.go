package csvfile

import (
	"encoding/csv"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	"strings"
)

// CsvConfig 用于配置 CSV 读取选项
type CsvConfig[T any] struct {
	Delimiter       rune   // 分隔符，默认为逗号
	HeaderRowIndex  int    // 表头位于第几行 -1，表示没有表头， 默认为0
	DataRowIndex    [2]int // 数据起始行索引，默认为 0, 数据结束行索引，0表示读取到文件末尾
	DataColumnIndex [2]int // 数据起始列索引，默认为 0, 数据结束行索引，0表示读取到文件列尾
	ColumnNumber    int    // 列数，如果有不满足的列，可能存在分隔符的问题，就得进行特殊处理，如果为0，就忽略检查
	TrimSpaces      bool   // 是否去除字段前后空格
	SkipEmptyRows   bool   // 是否跳过空行
	csvReader       *csv.Reader
	csvAllRecords   [][]string // 保留所有数据，避免每次读文件
}

// DefaultConfig 返回默认配置
func DefaultConfig[T any]() *CsvConfig[T] {
	return &CsvConfig[T]{
		Delimiter:       ',',
		HeaderRowIndex:  0,
		DataRowIndex:    [2]int{0, 0},
		DataColumnIndex: [2]int{0, 0},
		ColumnNumber:    0,
		TrimSpaces:      true,
		SkipEmptyRows:   true,
	}
}

func (c *CsvConfig[T]) validate() error {
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

	if c.DataColumnIndex[0] < 0 {
		c.DataColumnIndex[0] = 0
	}

	if c.DataColumnIndex[1] < 0 ||
		c.DataColumnIndex[1] <= c.DataColumnIndex[0] {
		c.DataColumnIndex[1] = 0
	}

	return nil
}

func (c *CsvConfig[T]) SetCsvReader(csvReader *csv.Reader) *CsvConfig[T] {
	if csvReader == nil {
		return c
	}

	if len(c.csvAllRecords) == 0 {
		records, err := csvReader.ReadAll()
		if err != nil {
			fmt.Println("WithCsvReader error:", err)
			return c
		}
		c.csvAllRecords = records
	}

	c.csvReader = csvReader
	return c
}
func (c *CsvConfig[T]) ExchangeList() ([]map[string]string, error) {
	if c.csvReader == nil && len(c.csvAllRecords) == 0 {
		return nil, fmt.Errorf("请先调用 SetCsvReader 设置 CSV 读取器")
	}
	return c.readToMap(c.csvReader)
}
func (c *CsvConfig[T]) ExchangeOne(rowIndex, columnIndex int) (string, error) {
	if c.csvReader == nil && len(c.csvAllRecords) == 0 {
		return "", fmt.Errorf("请先调用 SetCsvReader 设置 CSV 读取器")
	}
	if rowIndex <= 0 && columnIndex <= 0 {
		return "", fmt.Errorf("请指定行和列位置索引，从1开始: %d, %d", rowIndex, columnIndex)
	}
	records := c.csvAllRecords
	if len(records) == 0 {
		return "", fmt.Errorf("读取记录出错")
	}
	if len(records) < rowIndex {
		return "", fmt.Errorf("行索引超出范围: %d, %d", rowIndex, len(records))
	}

	columList := records[rowIndex-1]
	if len(columList) < columnIndex {
		return "", fmt.Errorf("列索引超出范围: %d, %d", columnIndex, len(columList))
	}

	return columList[columnIndex-1], nil
}

// ReadToList 读取 CSV 文件返回数据和表头
func (c *CsvConfig[T]) readToList(reader *csv.Reader) ([][]string, []string, error) {
	// 读取所有记录
	if c.Delimiter > 0 {
		reader.Comma = c.Delimiter
	}

	if err := c.validate(); err != nil {
		return nil, nil, fmt.Errorf("配置错误: %w", err)
	}

	records := c.csvAllRecords
	if len(records) == 0 {
		return nil, nil, fmt.Errorf("读取记录出错")
	}

	headers, err := c.extractHeaders(records)
	if err != nil {
		return nil, headers, fmt.Errorf("读取表头出错: %w", err)
	}

	result := make([][]string, 0)
	lastRowIndex := c.DataRowIndex[1] - 1
	if lastRowIndex <= 0 {
		lastRowIndex = len(records) - 1
	}

	for i := c.DataRowIndex[0] - 1; i <= lastRowIndex; i++ {
		record := records[i]
		if len(record) <= 1 {
			if c.SkipEmptyRows {
				continue
			}
		}
		newRecords, err := c.extractData(record)
		if err != nil {
			return result, headers, fmt.Errorf("读取数据出错: %w", err)
		}
		if c.ColumnNumber > 0 && len(newRecords) != c.ColumnNumber {
			return result, headers, fmt.Errorf("第 %d 行字段数量与列数不匹配: 列数 %d, 行 %d 个",
				c.DataRowIndex[0],
				c.ColumnNumber, len(newRecords))
		}
		result = append(result, newRecords)
	}

	return result, headers, nil
}

func (c *CsvConfig[T]) readToMap(reader *csv.Reader) ([]map[string]string, error) {
	dataList, headers, err := c.readToList(reader)
	if err != nil {
		return nil, err
	}
	allDataList := make([]map[string]string, 0)
	var retErr error
	lo.ForEachWhile(dataList, func(item []string, index int) bool {
		if len(headers) > 0 {
			if len(item) != len(headers) {
				retErr = fmt.Errorf("数据与header数量不统一: %s, %s", conv.String(item), conv.String(headers))
				return false
			}
			oneMap := make(map[string]string)
			lo.ForEach(item, func(oneItem string, j int) {
				oneMap[headers[j]] = oneItem
			})
			allDataList = append(allDataList, oneMap)
			return true
		} else {
			oneMap := make(map[string]string)
			lo.ForEach(item, func(oneItem string, j int) {
				dataKey := getExcelColumnName(j + 1)
				oneMap[dataKey] = oneItem
			})
			allDataList = append(allDataList, oneMap)
		}
		return true
	})
	if retErr != nil {
		return nil, retErr
	}
	return allDataList, nil
}

// headerFields 获取表头字段
func (c *CsvConfig[T]) extractHeaders(records [][]string) ([]string, error) {
	// 无表头配置
	if c.HeaderRowIndex < 0 {
		return nil, nil
	}

	// 检查表头行索引有效性
	if c.HeaderRowIndex >= len(records) {
		return nil, fmt.Errorf("表头行索引(%d)超出记录范围(共%d行)",
			c.HeaderRowIndex, len(records))
	}
	headerIndex := c.HeaderRowIndex - 1
	if headerIndex < 0 {
		return nil, nil
	}

	headerAllList := processFields(records[headerIndex], c.TrimSpaces)
	if c.DataColumnIndex[0] > 0 && c.DataColumnIndex[1] > c.DataColumnIndex[0] {
		headerColIndex := c.DataColumnIndex[0] - 1
		if headerColIndex >= 0 {
			headerAllList = headerAllList[headerColIndex:c.DataColumnIndex[1]]
		}
	}
	return headerAllList, nil
}

// extractData 获取表头字段
func (c *CsvConfig[T]) extractData(records []string) ([]string, error) {
	newRecords := processFields(records, c.TrimSpaces)
	if c.DataColumnIndex[0] > 0 && c.DataColumnIndex[1] > c.DataColumnIndex[0] {
		headerColIndex := c.DataColumnIndex[0] - 1
		if headerColIndex >= 0 {
			newRecords = newRecords[headerColIndex:c.DataColumnIndex[1]]
		}
	}
	return newRecords, nil
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

// 生成Excel列名，n表示第n列（从1开始）
func getExcelColumnName(n int) string {
	var result strings.Builder

	for n > 0 {
		// 调整为0-25的范围
		n--
		// 计算当前字符
		result.WriteByte('A' + byte(n%26))
		// 继续处理更高位
		n = n / 26
	}

	// 反转结果，因为我们是从低位开始计算的
	return reverseString(result.String())
}

// 反转字符串
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
