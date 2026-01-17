package mask_test

import (
	"github.com/magic-lib/go-plat-utils/mask"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestMask(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"Phone", []any{"1"}, []any{"1"}, mask.Phone},
		{"Phone", []any{"22"}, []any{"*2"}, mask.Phone},
		{"Phone", []any{"333"}, []any{"*33"}, mask.Phone},
		{"Phone", []any{"4444"}, []any{"**44"}, mask.Phone},
		{"Phone", []any{"55555"}, []any{"**555"}, mask.Phone},
		{"Phone", []any{"666666"}, []any{"***666"}, mask.Phone},
		{"Phone", []any{"7777777"}, []any{"***7777"}, mask.Phone},
		{"Phone", []any{"88888888"}, []any{"8***8888"}, mask.Phone},
		{"Phone", []any{"999999999"}, []any{"99***9999"}, mask.Phone},
		{"Phone", []any{"0000000000"}, []any{"00****0000"}, mask.Phone},
		{"Phone", []any{"13722223345"}, []any{"137****3345"}, mask.Phone},
		{"Email", []any{"zhangsan@go-mall.com"}, []any{"zh****an@go-mall.com"}, mask.Email},
		{"Email", []any{"you@go-mall.com"}, []any{"y*u@go-mall.com"}, mask.Email},
		{"Email", []any{"dear@go-mall.com"}, []any{"d**r@go-mall.com"}, mask.Email},
		{"RealName", []any{"张三"}, []any{"张*"}, mask.RealName},
		{"RealName", []any{"赵丽颖"}, []any{"赵*颖"}, mask.RealName},
	}
	utils.TestFunction(t, testCases, nil)
}
