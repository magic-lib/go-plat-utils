package csvfile_test

import (
	"encoding/csv"
	"fmt"
	"github.com/magic-lib/go-plat-utils/file/csvfile"
	"os"
	"testing"
)

// processFields 处理字段，去除空格等
func TestProcessFields(t *testing.T) {
	filePath := "/Volumes/MacintoshData/tianlin0/Downloads/111.csv"
	configTemp := csvfile.DefaultConfig[string]()

	configTemp.HeaderRowIndex = 11
	configTemp.DataRowIndex = [2]int{12, 19}

	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	kk := csv.NewReader(file)

	aa, bb, cc := configTemp.ReadToList(kk)
	fmt.Println(aa, bb, cc)
}
