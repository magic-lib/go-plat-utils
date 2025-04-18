package mask_test

import (
	"github.com/magic-lib/go-plat-utils/mask"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestMask(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"Phone", []any{"13722223345"}, []any{"137****3345"}, mask.Phone},
		{"Phone", []any{"13722"}, []any{"137**"}, mask.Phone},
		{"Phone", []any{"13"}, []any{"13"}, mask.Phone},
		{"Email", []any{"zhangsan@go-mall.com"}, []any{"zh****an@go-mall.com"}, mask.Email},
		{"Email", []any{"you@go-mall.com"}, []any{"y*u@go-mall.com"}, mask.Email},
		{"Email", []any{"dear@go-mall.com"}, []any{"d**r@go-mall.com"}, mask.Email},
		{"RealName", []any{"张三"}, []any{"张*"}, mask.RealName},
		{"RealName", []any{"赵丽颖"}, []any{"赵*颖"}, mask.RealName},
	}
	utils.TestFunction(t, testCases, nil)
}
