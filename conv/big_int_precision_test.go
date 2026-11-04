package conv_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

type bigIntT struct {
	Id     int64   `json:"id"`
	Uid    uint64  `json:"uid,omitempty"`
	Name   string  `json:"name"`
	Amount int64   `json:"amount,omitempty"`
	Rate   float64 `json:"rate"`
	Any    any     `json:"any"`
}

// struct 转字符串时，超过 2^53 的整数不能丢精度，输出按 key 排序且不转义 &
func TestStringStructBigInt(t *testing.T) {
	s := bigIntT{Id: 1897690027819798529, Uid: 18446744073709551615, Name: "a&b", Amount: -1897690027819798529, Rate: 1.5, Any: 3}
	want := `{"amount":-1897690027819798529,"any":3,"id":1897690027819798529,"name":"a&b","rate":1.5,"uid":18446744073709551615}`
	if got := conv.String(s); got != want {
		t.Errorf("String(struct) =\n%s\nwant\n%s", got, want)
	}
	if got := conv.String(&s); got != want {
		t.Errorf("String(&struct) = %s", got)
	}

	// 不含大整数时输出与原来一致（仍走 jcs）
	small := bigIntT{Id: 12, Uid: 3, Name: "x", Amount: 4, Rate: 0.1, Any: 2}
	if got, want := conv.String(small), `{"amount":4,"any":2,"id":12,"name":"x","rate":0.1,"uid":3}`; got != want {
		t.Errorf("String(small struct) = %s, want %s", got, want)
	}
}

// 字符串反序列化到 struct 时大整数不丢精度，float64 与 any 字段类型不变
func TestUnmarshalStringBigInt(t *testing.T) {
	js := `{"id":1897690027819798529,"uid":18446744073709551615,"name":"a","amount":-1897690027819798529,"rate":2.5,"any":7}`
	var a bigIntT
	if err := conv.Unmarshal(js, &a); err != nil {
		t.Fatalf("Unmarshal err: %v", err)
	}
	if a.Id != 1897690027819798529 || a.Uid != 18446744073709551615 || a.Amount != -1897690027819798529 {
		t.Errorf("大整数丢精度: %+v", a)
	}
	if a.Rate != 2.5 || a.Name != "a" {
		t.Errorf("普通字段: %+v", a)
	}
	if v, ok := a.Any.(float64); !ok || v != 7 {
		t.Errorf("any 字段应仍为 float64(7), got %v(%T)", a.Any, a.Any)
	}

	var list []bigIntT
	if err := conv.Unmarshal(`[{"id":1897690027819798529}]`, &list); err != nil || len(list) != 1 || list[0].Id != 1897690027819798529 {
		t.Errorf("Unmarshal 数组 = %+v, %v", list, err)
	}
}
