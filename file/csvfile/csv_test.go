package csvfile_test

import (
	"encoding/csv"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/file/csvfile"
	"os"
	"testing"
)

// processFields 处理字段，去除空格等
func TestProcessFields(t *testing.T) {
	filePath := "/Volumes/MacintoshData/tianlin0/Downloads/111.csv"
	configTemp := csvfile.DefaultConfig[string]()

	configTemp.HeaderRowIndex = 0
	configTemp.DataRowIndex = [2]int{13, 20}
	configTemp.DataColumnIndex = [2]int{1, 7}
	configTemp.ColumnNumber = 7

	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()

	configTemp.SetCsvReader(csv.NewReader(file))

	aa, err := configTemp.ExchangeList()

	bb, err := configTemp.ExchangeOne(5, 5)

	fmt.Println(conv.String(aa))
	fmt.Println(conv.String(bb))
	fmt.Println(conv.String(err))
}
