package utils_test

import (
	"strings"
	"testing"

	"github.com/magic-lib/go-plat-utils/utils"
)

func TestFuncExecutePanic(t *testing.T) {
	result, err := utils.FuncExecute(func(a int) int {
		if a > 0 {
			panic("boom")
		}
		return a
	}, 1)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("被调函数 panic 时应返回错误, got result=%v err=%v", result, err)
	}
	if result != nil {
		t.Fatalf("panic 时 result 应为 nil, got %v", result)
	}

	result, err = utils.FuncExecute(func(a int) int { return a + 1 }, 1)
	if err != nil || len(result) != 1 || result[0] != 2 {
		t.Fatalf("正常调用 result=%v err=%v", result, err)
	}
}

func TestFuncExecuteNilArg(t *testing.T) {
	result, err := utils.FuncExecute(func(v any, s *string, rest ...any) bool {
		return v == nil && s == nil && len(rest) == 1 && rest[0] == nil
	}, nil, nil, nil)
	if err != nil || len(result) != 1 || result[0] != true {
		t.Fatalf("nil 实参应按形参类型传零值, got result=%v err=%v", result, err)
	}
}
