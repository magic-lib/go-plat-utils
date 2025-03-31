package compress_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/compress"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestCompress(t *testing.T) {

	testCases := []*utils.TestStruct{
		{"BrCompress", []any{"tianlin0"}, []any{[]byte{139, 3, 128, 116, 105, 97, 110, 108, 105, 110, 48}}, func(input string) []byte {
			str, err := compress.BrCompress([]byte(input))
			if err != nil {
				return nil
			}
			fmt.Println("BrCompress:", []byte(input), str)
			return str
		}},
		{"BrUnCompress", []any{"tianlin0"}, []any{"tianlin0"}, func(input string) string {

			strCom, err := compress.BrCompress([]byte(input))
			if err != nil {
				return ""
			}

			str, err := compress.BrUnCompress(strCom)
			if err != nil {
				return ""
			}
			fmt.Println("BrUnCompress:", input, str)
			return string(str)
		}},
		{"GZipCompress", []any{"tianlin0"}, []any{[]byte{31, 139, 8, 0, 0, 0, 0, 0, 2, 255, 42, 201, 76, 204, 203, 201, 204, 51, 0, 0, 0, 0, 255, 255, 1, 0, 0, 255, 255, 119, 146, 237, 93, 8, 0, 0, 0}}, func(input string) []byte {
			str, err := compress.GZipCompress([]byte(input))
			if err != nil {
				return nil
			}
			fmt.Println("GZipCompress:", []byte(input), str)
			return str
		}},
		{"GZipUnCompress", []any{"tianlin0"}, []any{"tianlin0"}, func(input string) string {

			strCom, err := compress.GZipCompress([]byte(input))
			if err != nil {
				return ""
			}

			str, err := compress.GZipUnCompress(strCom)
			if err != nil {
				t.Error(err)
				return ""
			}
			fmt.Println("GZipUnCompress:", input, str)
			return string(str)
		}},
	}
	utils.TestFunction(t, testCases, nil)
}
