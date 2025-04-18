package utils

import (
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestStruct struct {
	Name     string
	Inputs   []any
	Expected []any
	Func     any //每个执行方法不同的情况
}

// TestFunction 统一检查函数
func TestFunction(t *testing.T, paramList []*TestStruct, checkFunc any) {
	for _, tc := range paramList {
		t.Run(tc.Name, func(t *testing.T) {
			var checkFuncTemp any
			if tc.Func != nil {
				checkFuncTemp = tc.Func
			} else {
				if checkFunc != nil {
					checkFuncTemp = checkFunc
				} else {
					t.Errorf("[%s] error: not set CheckFunc, Input: (%v)", tc.Name, conv.String(tc.Inputs))
					return
				}
			}

			result, err := FuncExecute(checkFuncTemp, tc.Inputs...)
			funcName, _ := GetFuncName(checkFuncTemp)
			if err != nil {
				t.Errorf("[%s] funcName: %s, error: (%v), Input: (%v)", tc.Name, funcName, err, conv.String(tc.Inputs))
				return
			}
			//比较最小的内容，有可能有后面返回error，不写出来的情况
			//也有可能以前写的方法返回3个值，后面改为2个
			minNum := len(tc.Expected)
			if len(tc.Expected) > len(result) {
				minNum = len(result)
			}

			for i := 0; i < minNum; i++ {
				assert.Equal(t, tc.Expected[i], result[i])
			}
		})
	}
}
