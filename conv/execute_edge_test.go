package conv_test

import (
	"reflect"
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

func TestParamsTypeChangeEdge(t *testing.T) {
	fn := reflect.ValueOf(func(a int64, b any, c int) {})
	// 实参少于形参、实参不是数字、接口参数为无效值：都不能 panic
	params := []reflect.Value{reflect.ValueOf(int32(5)), {}}
	got := conv.ParamsTypeChange(fn, params)
	if got[0].Kind() != reflect.Int64 || got[0].Int() != 5 {
		t.Fatalf("int32 应转为 int64, got %v", got[0])
	}
	if !got[1].IsValid() || got[1].Type() != reflect.TypeOf((*any)(nil)).Elem() {
		t.Fatalf("无效的接口参数应补为接口零值, got %v", got[1])
	}
	strParams := conv.ParamsTypeChange(fn, []reflect.Value{reflect.ValueOf("x")})
	if strParams[0].Kind() != reflect.String {
		t.Fatalf("非数字实参应保持原样, got %v", strParams[0].Kind())
	}
}

func TestGetWantedValueConvert(t *testing.T) {
	cases := []struct {
		in   any
		to   reflect.Type
		want any
	}{
		{int32(7), reflect.TypeOf(int64(0)), int64(7)},
		{int32(7), reflect.TypeOf(uint64(0)), uint64(7)},
		{int(3), reflect.TypeOf(float64(0)), float64(3)},
		{float32(2), reflect.TypeOf(int(0)), int(2)},
	}
	for _, c := range cases {
		got, err := conv.GetWantedValue(reflect.ValueOf(c.in), c.to)
		if err != nil || got.Interface() != c.want {
			t.Errorf("GetWantedValue(%T(%v), %s) = %v(%T), %v", c.in, c.in, c.to, got.Interface(), got.Interface(), err)
		}
	}
}
