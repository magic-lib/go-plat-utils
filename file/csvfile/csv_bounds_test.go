package csvfile

import (
	"encoding/csv"
	"strings"
	"testing"
)

// 默认配置（DataRowIndex 为 0）、自定义分隔符、列索引超出范围都不能 panic
func TestCsvDefaultConfigAndDelimiter(t *testing.T) {
	cfg := DefaultConfig[any]()
	cfg.Delimiter = ';'
	cfg.DataColumnIndex = [2]int{1, 10}
	cfg.SetCsvReader(csv.NewReader(strings.NewReader("a;b\n1;2\n")))
	list, err := cfg.ExchangeList()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[1]["A"] != "1" || list[1]["B"] != "2" {
		t.Fatalf("got %#v", list)
	}
}

func TestCsvExchangeOneZeroIndex(t *testing.T) {
	cfg := DefaultConfig[any]()
	cfg.SetCsvReader(csv.NewReader(strings.NewReader("a,b\n")))
	if _, err := cfg.ExchangeOne(0, 1); err == nil {
		t.Fatal("rowIndex 0 should return error")
	}
	if v, err := cfg.ExchangeOne(1, 2); err != nil || v != "b" {
		t.Fatalf("got %q %v", v, err)
	}
}
