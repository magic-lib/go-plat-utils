package conv_test

import (
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestAnyToBool(t *testing.T) {
	iPtr := 90
	testCases := []*utils.TestStruct{
		{"bool true", []any{true}, []any{true, true}, conv.Bool},
		{"bool false", []any{false}, []any{false}, conv.Bool},
		{"int -1", []any{int(-1)}, []any{true}, conv.Bool},
		{"int 1", []any{int(1)}, []any{true}, conv.Bool},
		{"int 0", []any{int(0)}, []any{false}, conv.Bool},
		{"int8 1", []any{int8(1)}, []any{true}, conv.Bool},
		{"int8 0", []any{int8(0)}, []any{false}, conv.Bool},
		{"int16 1", []any{int16(1)}, []any{true}, conv.Bool},
		{"int16 0", []any{int16(0)}, []any{false}, conv.Bool},
		{"int32 1", []any{int32(1)}, []any{true}, conv.Bool},
		{"int32 0", []any{int32(0)}, []any{false}, conv.Bool},
		{"int64 1", []any{int64(1)}, []any{true}, conv.Bool},
		{"int64 100", []any{int64(100)}, []any{true}, conv.Bool},
		{"int64 0", []any{int64(0)}, []any{false}, conv.Bool},
		{"uint 1", []any{uint(1)}, []any{true}, conv.Bool},
		{"uint 0", []any{uint(0)}, []any{false}, conv.Bool},
		{"uint8 1", []any{uint8(1)}, []any{true}, conv.Bool},
		{"uint8 0", []any{uint8(0)}, []any{false}, conv.Bool},
		{"uint16 1", []any{uint16(1)}, []any{true}, conv.Bool},
		{"uint16 0", []any{uint16(0)}, []any{false}, conv.Bool},
		{"uint32 1", []any{uint32(1)}, []any{true}, conv.Bool},
		{"uint32 0", []any{uint32(0)}, []any{false}, conv.Bool},
		{"uint64 1", []any{uint64(1)}, []any{true}, conv.Bool},
		{"uint64 0", []any{uint64(0)}, []any{false}, conv.Bool},
		{"float32 1.0", []any{float32(1.0)}, []any{true}, conv.Bool},
		{"float32 0.0", []any{float32(0.0)}, []any{false}, conv.Bool},
		{"float64 1.0", []any{float64(1.0)}, []any{true}, conv.Bool},
		{"float64 0.0", []any{float64(0.0)}, []any{false}, conv.Bool},
		{"string abc", []any{"abc"}, []any{true}, conv.Bool},
		{"string true", []any{"true"}, []any{true}, conv.Bool},
		{"string false", []any{"false"}, []any{false}, conv.Bool},
		{"empty string", []any{""}, []any{false}, conv.Bool},
		{"nil value", []any{nil}, []any{false}, conv.Bool},
		{"complex64 1+1i", []any{complex64(1 + 1i)}, []any{true}, conv.Bool},
		{"complex64 0+0i", []any{complex64(0 + 0i)}, []any{false}, conv.Bool},
		{"complex128 1+1i", []any{complex128(1 + 1i)}, []any{true}, conv.Bool},
		{"complex128 0+0i", []any{complex128(0 + 0i)}, []any{false}, conv.Bool},
		{"nil pointer", []any{(*int)(nil)}, []any{false}, conv.Bool},
		{"non-nil pointer", []any{&iPtr}, []any{true}, conv.Bool},
		{"empty slice no value", []any{[]int{}}, []any{false, true}, conv.Bool},
		{"empty slice has value", []any{[]int{1, 2, 3}}, []any{true, true}, conv.Bool},
	}
	utils.TestFunction(t, testCases, conv.Bool)
}
