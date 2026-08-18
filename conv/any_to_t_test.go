package conv_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	jsoniterForNil "github.com/magic-lib/go-plat-utils/internal/jsoniter/go"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/shopspring/decimal"
	"regexp"
	"testing"
	"time"
)

func TestAnyToBool(t *testing.T) {
	iPtr := 90
	testCases := []*utils.TestStruct{
		{"bool true", []any{true}, []any{true}, nil},
		{"bool false", []any{false}, []any{false}, nil},
		{"int -1", []any{int(-1)}, []any{true}, nil},
		{"int 1", []any{int(1)}, []any{true}, nil},
		{"int 0", []any{int(0)}, []any{false}, nil},
		{"int8 1", []any{int8(1)}, []any{true}, nil},
		{"int8 0", []any{int8(0)}, []any{false}, nil},
		{"int16 1", []any{int16(1)}, []any{true}, nil},
		{"int16 0", []any{int16(0)}, []any{false}, nil},
		{"int32 1", []any{int32(1)}, []any{true}, nil},
		{"int32 0", []any{int32(0)}, []any{false}, nil},
		{"int64 1", []any{int64(1)}, []any{true}, nil},
		{"int64 100", []any{int64(100)}, []any{true}, nil},
		{"int64 0", []any{int64(0)}, []any{false}, nil},
		{"uint 1", []any{uint(1)}, []any{true}, nil},
		{"uint 0", []any{uint(0)}, []any{false}, nil},
		{"uint8 1", []any{uint8(1)}, []any{true}, nil},
		{"uint8 0", []any{uint8(0)}, []any{false}, nil},
		{"uint16 1", []any{uint16(1)}, []any{true}, nil},
		{"uint16 0", []any{uint16(0)}, []any{false}, nil},
		{"uint32 1", []any{uint32(1)}, []any{true}, nil},
		{"uint32 0", []any{uint32(0)}, []any{false}, nil},
		{"uint64 1", []any{uint64(1)}, []any{true}, nil},
		{"uint64 0", []any{uint64(0)}, []any{false}, nil},
		{"float32 1.0", []any{float32(1.0)}, []any{true}, nil},
		{"float32 0.0", []any{float32(0.0)}, []any{false}, nil},
		{"float64 1.0", []any{float64(1.0)}, []any{true}, nil},
		{"float64 0.0", []any{float64(0.0)}, []any{false}, nil},
		{"string abc", []any{"abc"}, []any{false}, nil},
		{"string true", []any{"true"}, []any{true}, nil},
		{"string false", []any{"false"}, []any{false}, nil},
		{"empty string", []any{""}, []any{false}, nil},
		{"nil value", []any{nil}, []any{false}, nil},
		{"complex64 1+1i", []any{complex64(1 + 1i)}, []any{true}, nil},
		{"complex64 0+0i", []any{complex64(0 + 0i)}, []any{false}, nil},
		{"complex128 1+1i", []any{complex128(1 + 1i)}, []any{true}, nil},
		{"complex128 0+0i", []any{complex128(0 + 0i)}, []any{false}, nil},
		{"nil pointer", []any{(*int)(nil)}, []any{false}, nil},
		{"non-nil pointer", []any{&iPtr}, []any{true}, nil},
		{"empty slice no value", []any{[]int{}}, []any{false}, nil},
		{"empty slice has value", []any{[]int{1, 2, 3}}, []any{true}, nil},
	}
	utils.TestFunction(t, testCases, conv.Convert[bool])
}

func FormatFullTime(t time.Time) string {
	// 自定义布局：覆盖年月日时分秒毫秒 + 时区偏移 + 时区缩写
	const fullLayout = "2006-01-02 15:04:05.000 -07:00 (MST)"
	return t.Format(fullLayout)
}
func TestAnyToTime(t *testing.T) {
	timeFunc := func(val string) bool {
		tt, err := conv.Convert[time.Time](val)
		fmt.Println("当前完整时间：", val)
		fmt.Println("当前完整时间：", FormatFullTime(tt))
		if err == nil {
			return true
		}
		return false
	}

	testCases := []*utils.TestStruct{
		{"2025-01-02T15:04:05Z07:00", []any{"2025-01-02T18:04:05Z07:00"}, []any{true}, nil},
		{"02/01/2006", []any{"02/01/2036"}, []any{true}, nil},
		{"02/1/2006", []any{"02/1/2036"}, []any{true}, nil},
		{"02/1/2006 15:04:05", []any{"02/1/2026 18:04:05"}, []any{true}, nil},
		{time.DateOnly, []any{time.DateOnly}, []any{true}, nil},
		{"2006.01", []any{"2026.01"}, []any{true}, nil},
		{"2006/1/02 15:04:05", []any{"2026/1/02 18:04:05"}, []any{true}, nil},
		{"2006/1/02", []any{"2026/1/02"}, []any{true}, nil},
		{"2006-01-02 15:04:05", []any{"2026-01-02 18:04:05"}, []any{true}, nil},
		{"2006/1/02 15:04", []any{"2026/1/02 18:04"}, []any{true}, nil},
		{"2006/1/2 15:04", []any{"2016/10/20 18:04"}, []any{true}, nil},
		{"02-Jan-2006", []any{"24-Jan-2016"}, []any{true}, nil},
		{"2-Jan-2006", []any{"24-Jan-2026"}, []any{true}, nil},
		{"2006/1/02 15:04:05:00", []any{"2017/10/02 15:04:05:00"}, []any{true}, nil},
		{"Jan 02, 2006", []any{"Jan 25, 2016"}, []any{true}, nil},
		{"2025/12/5", []any{"2025/12/5"}, []any{true}, nil},
		{"2025/9/5 11:28", []any{"2025/9/5 11:28"}, []any{true}, nil},
		{"2025/09/05 11:28", []any{"2025/09/05 11:28"}, []any{true}, nil},
		{"2025/9/5 9:13", []any{"2025/9/5 9:13"}, []any{true}, nil},
		{"2025-9-12 12:04:07", []any{"2025-9-12 12:04:07"}, []any{true}, nil},
		{"2025/12/5", []any{"2025/12/5"}, []any{true}, nil},
		{"2025-12", []any{"2025-12"}, []any{true}, nil},
	}
	utils.TestFunction(t, testCases, timeFunc)
}

func TestAnyToNumber(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"string 1 to int 1", []any{"1"}, []any{1}, conv.Convert[int]},
		{"bool to int 0", []any{false}, []any{0}, conv.Convert[int]},
		{"int 0 to int8 0", []any{0}, []any{int8(0)}, conv.Convert[int8]},
		{"int 0 to int16 0", []any{0}, []any{int16(0)}, conv.Convert[int16]},
		{"int 0 to int32 0", []any{0}, []any{int32(0)}, conv.Convert[int32]},
		{"int -5 to int64", []any{-5}, []any{int64(-5)}, conv.Convert[int64]},
		{"int 5 to uint", []any{5}, []any{uint(5)}, conv.Convert[uint]},
		{"int 5 to uint8", []any{5}, []any{uint8(5)}, conv.Convert[uint8]},
		{"int 5 to uint16", []any{5}, []any{uint16(5)}, conv.Convert[uint16]},
		{"int 5 to uint32", []any{5}, []any{uint32(5)}, conv.Convert[uint32]},
		{"int 5 to uint64", []any{5}, []any{uint64(5)}, conv.Convert[uint64]},
		{"int -5 to float32", []any{-5}, []any{float32(-5)}, conv.Convert[float32]},
		{"int -5 to float64", []any{-5}, []any{float64(-5)}, conv.Convert[float64]},
		{"int -5 to string", []any{-5}, []any{"-5"}, conv.Convert[string]},
	}
	utils.TestFunction(t, testCases, nil)
}

func TestConvertForTypeString(t *testing.T) {
	// 类型为空：原样返回
	if got, ok := conv.ConvertForTypeString("", "abc"); !ok || got != "abc" {
		t.Errorf("ConvertForTypeString(\"\") = (%v, %v), want (abc, true)", got, ok)
	}

	cases := []struct {
		name string
		typ  string
		raw  any
		want any
	}{
		{"string", "string", 123, "123"},
		{"bool", "bool", "true", true},
		{"int", "int", "42", int(42)},
		{"int8", "int8", "8", int8(8)},
		{"int16", "int16", "16", int16(16)},
		{"int32", "int32", "32", int32(32)},
		{"int64", "int64", "64", int64(64)},
		{"uint", "uint", "7", uint(7)},
		{"uint8", "uint8", "8", uint8(8)},
		{"uint16", "uint16", "16", uint16(16)},
		{"uint32", "uint32", "32", uint32(32)},
		{"uint64", "uint64", "64", uint64(64)},
		{"float32", "float32", "3.5", float32(3.5)},
		{"float64", "float64", "3.5", float64(3.5)},
		{"decimal", "decimal", "1.23", decimal.RequireFromString("1.23")},
		{"nil", "nil", 123, nil},
		{"nil with case/space", "  NIL ", "abc", nil},
	}
	for _, c := range cases {
		got, ok := conv.ConvertForTypeString(c.typ, c.raw)
		if !ok {
			t.Errorf("%s: ConvertForTypeString(%q, %v) ok = false, want true", c.name, c.typ, c.raw)
			continue
		}
		// decimal.Decimal 需用 Equal 比较（内部为指针字段，== 不可用）
		if dv, isDec := got.(decimal.Decimal); isDec {
			wantD, ok := c.want.(decimal.Decimal)
			if !ok || !dv.Equal(wantD) {
				t.Errorf("%s: ConvertForTypeString(%q, %v) = %v (%T), want %v (%T)", c.name, c.typ, c.raw, got, got, c.want, c.want)
			}
			continue
		}
		if got != c.want {
			t.Errorf("%s: ConvertForTypeString(%q, %v) = %v (%T), want %v (%T)", c.name, c.typ, c.raw, got, got, c.want, c.want)
		}
	}

	// time / time.Time：只校验解析结果的时间字段（时区由 conf.TimeLocation 决定，不与固定位置比较）
	for _, typ := range []string{"time"} {
		got, ok := conv.ConvertForTypeString(typ, "2026-08-15 12:00:00")
		if !ok {
			t.Errorf("%s: ConvertForTypeString(%q) ok = false, want true", typ, typ)
			continue
		}
		tv, ok := got.(time.Time)
		if !ok {
			t.Errorf("%s: ConvertForTypeString(%q) 返回 %T, want time.Time", typ, typ, got)
			continue
		}
		if tv.Year() != 2026 || tv.Month() != 8 || tv.Day() != 15 || tv.Hour() != 12 {
			t.Errorf("%s: ConvertForTypeString(%q) = %v, want 2026-08-15 12:00:00", typ, typ, tv)
		}
	}

	// 转换失败：返回 (raw, false)
	if got, ok := conv.ConvertForTypeString("int64", "abc"); ok || got != "abc" {
		t.Errorf("ConvertForTypeString(int64, abc) = (%v, %v), want (abc, false)", got, ok)
	}
	// 未知类型：返回 (raw, false)
	if got, ok := conv.ConvertForTypeString("unknown", 1); ok || got != 1 {
		t.Errorf("ConvertForTypeString(unknown, 1) = (%v, %v), want (1, false)", got, ok)
	}
}

func TestConvertForTypeJs(t *testing.T) {
	// 类型为空：原样返回
	if got, ok := conv.ConvertForTypeJs("", 123); !ok || got != 123 {
		t.Errorf("ConvertForTypeJs(\"\") = (%v, %v), want (123, true)", got, ok)
	}

	cases := []struct {
		name string
		typ  string
		raw  any
		want any
	}{
		{"string", "string", 123, "123"},
		{"boolean", "boolean", "true", true},
		{"boolean with case", " BOOLEAN ", "false", false},
		{"number int", "number", 42, int64(42)},
		{"number float64 int", "number", float64(42), int64(42)},
		{"number float64", "number", 42.5, float64(42.5)},
		{"number string int", "number", "42", int64(42)},
		{"number string float", "number", "3.14", float64(3.14)},
		{"number string exponent", "number", "1e3", float64(1000)},
		{"bigint", "bigint", "9007199254740993", int64(9007199254740993)},
		{"null", "null", 123, nil},
		{"undefined", "undefined", "abc", nil},
		{"array", "array", `[1,2,3]`, []any{float64(1), float64(2), float64(3)}},
		{"object", "object", `{"a":1}`, map[string]any{"a": float64(1)}},
		{"date", "date", "2026-08-15 12:00:00", time.Time{}},
	}
	for _, c := range cases {
		got, ok := conv.ConvertForTypeJs(c.typ, c.raw)
		if !ok {
			t.Errorf("%s: ConvertForTypeJs(%q, %v) ok = false, want true", c.name, c.typ, c.raw)
			continue
		}
		switch want := c.want.(type) {
		case time.Time:
			// 时区由 conf.TimeLocation 决定，只校验字段
			tv, isTime := got.(time.Time)
			if !isTime || tv.Year() != 2026 || tv.Month() != 8 || tv.Day() != 15 || tv.Hour() != 12 {
				t.Errorf("%s: ConvertForTypeJs(%q) = %v (%T), want 2026-08-15 12:00:00", c.name, c.typ, got, got)
			}
		case map[string]any:
			gm, isMap := got.(map[string]any)
			if !isMap {
				t.Errorf("%s: ConvertForTypeJs(%q) = %T, want map[string]any", c.name, c.typ, got)
				continue
			}
			if len(gm) != len(want) {
				t.Errorf("%s: ConvertForTypeJs(%q) = %v, want %v", c.name, c.typ, got, want)
			}
		default:
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", c.want) {
				t.Errorf("%s: ConvertForTypeJs(%q, %v) = %v (%T), want %v (%T)", c.name, c.typ, c.raw, got, got, c.want, c.want)
			}
		}
	}

	// 未知类型（symbol/function 等）：返回 (raw, false)
	if got, ok := conv.ConvertForTypeJs("symbol", 1); ok || got != 1 {
		t.Errorf("ConvertForTypeJs(symbol, 1) = (%v, %v), want (1, false)", got, ok)
	}
}

type Table1 struct {
	CreateTime sql.NullTime   `db:"create_time" json:"create_time"`
	Name       sql.NullString `db:"name" json:"name"`
}

func insertSql(in any) {
	if cond.IsArray(in) {
		inList1, err := conv.Convert[[]Table1](in)
		if err == nil {
			fmt.Println(inList1)
		}

		inList2, err := conv.Convert[[]*Table1](in)
		if err == nil {
			fmt.Println(inList2)
		}

		inList3, err := conv.Convert[[]any](in)
		if err == nil {
			fmt.Println(inList3)
		}
		num := []int64{1, 2, 3}
		inList4, err := conv.Convert[[]string](num)
		if err == nil {
			fmt.Println(inList4)
		}
	}
	fmt.Println("errr")
}

func TestAnyToConvert(t *testing.T) {
	aa := []Table1{
		{
			CreateTime: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			Name: sql.NullString{
				String: "aaaaa",
				Valid:  true,
			},
		},
		{
			CreateTime: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			Name: sql.NullString{
				String: "bbbbb",
				Valid:  true,
			},
		},
	}
	insertSql(aa)
}
func TestAnyToString(t *testing.T) {
	name := sql.NullString{
		String: "bbbbb",
		Valid:  true,
	}
	aa := conv.String(name)
	fmt.Println(aa)
}

type CollectionHistory struct {
	AppointTime time.Time `db:"appoint_time" json:"appoint_time"` // 预约时间
}

func TestTimeToString(t *testing.T) {
	name := time.Now() // 2026-06-20T06:30:43+08:00
	bb, _ := json.Marshal(name)

	// 2026-06-22T17:09:41.47956+08:00
	fmt.Println(string(bb))

	cc, _ := jsoniterForNil.MarshalToString(name)
	fmt.Println(string(cc))

	mm := &CollectionHistory{
		AppointTime: time.Now(),
	}

	fmt.Println(conv.String(mm))

	aa := conv.String(name)
	fmt.Println(aa)

	resp := "aaaaaaaa 2026-06-22T17:09:41.47956+08:00 bbbbbb 2026-06-20T06:30:43+08:00"
	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})[^\s"]*`)
	resp = re.ReplaceAllString(resp, "$1 $2")
	fmt.Println(resp)
}

//func TestString(t *testing.T) {
//	lastFinishedOrder := &creditpb.LastFinishedOrder{
//		LoanAmount: 1000,
//
//		ActiveRepaidAmount:  34,
//		PassiveRepaidAmount: 13,
//		OrderEndTime:        &timestamppb.Timestamp{},
//		OrderFinishTime:     &timestamppb.Timestamp{},
//		IsActiveRepaid:      true,
//		ActualDueDayDiff:    13,
//	}
//
//	str := conv.String(lastFinishedOrder)
//	fmt.Println(str)
//}
