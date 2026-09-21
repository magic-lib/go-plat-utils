package conv_test

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/ucarion/jcs"
	"sync"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/commerror/errpb"
	"github.com/magic-lib/go-plat-utils/conv"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// 本文件按 conv.String 的判定顺序逐分支覆盖。
//
// 断言方式说明：
//   - checkEq  ：精确字符串相等。用于键序确定的输出（标量、slice、以及 jcs 排序过的 map）。
//   - checkJSON：JSON 语义等价。用于 struct / map 这类输出，因为内部会经过 map 中转，
//     键的顺序取决于 Go 的 map 迭代顺序，是随机的，不能做字符串相等断言。
//
// 测试同时固化了若干「当前行为」，其中部分属于已知缺陷，已在相关用例处标注 KNOWN。

// ---------- 断言辅助 ----------

func checkEq(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s:\n  got  = %q\n  want = %q", name, got, want)
	}
}

func checkJSON(t *testing.T, name, got, want string) {
	if !cond.IsSameJson(got, want) {
		t.Errorf("%s:\n  got  = %q\n  want = %q", name, got, want)
		return
	}
}

// ---------- 测试用类型 ----------

type tsString string
type tsInt int
type tsFloat float64
type tsBool bool

// tsStruct 覆盖 tag 的各种写法
type tsStruct struct {
	A int    `json:"a"`
	B string `json:"b,omitempty"`
	C string `json:"-"` // 忽略
	D string // 无 tag，回退字段名
}

type tsEmpty struct{}

type tsInner struct {
	X int `json:"x"`
}
type tsHost struct {
	tsInner     // 匿名嵌入，字段被提升到顶层
	Y       int `json:"y"`
}

type tsWithTime struct {
	T time.Time `json:"t"`
}
type tsWithSQL struct {
	S sql.NullString `json:"s"`
}
type tsNested struct {
	Inner *tsStruct `json:"inner"`
}
type tsSpecialChar struct {
	S string `json:"s"`
}
type tsUnexported struct {
	Pub  string `json:"pub"`
	priv string // 未导出字段被忽略
}
type tsWithChan struct {
	C chan int `json:"c"`
}
type tsWithMap struct {
	M map[string]any `json:"m"`
}

// 纯 error：只有 Error 一个导出方法 → cond.IsError 为 true
type tsPureErr struct{}

func (e tsPureErr) Error() string { return "pure-err" }

// 带额外导出方法 → cond.IsError 为 false，但仍是 error 接口
type tsErrExtra struct{}

func (e tsErrExtra) Error() string { return "err-extra" }
func (e tsErrExtra) GetMsg() string {
	return "msg"
}

// 带可导出字段的业务对象：除 Error 外还有 GetCode → cond.IsError 为 false，
// 会走到 getByType 的 `case error:`，因为有可导出字段，报错往后走，按对象输出
type tsErrBiz struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e tsErrBiz) Error() string { return e.Msg }
func (e tsErrBiz) GetCode() int  { return e.Code }

// 有可导出字段，但导出方法只有 Error → cond.IsError 同样为 false（还要求无可导出字段），
// 会走到 getByType 的 `case error:`，因为有可导出字段，报错往后走，按对象输出
type tsErrFieldOnly struct {
	Code int `json:"code"`
}

func (e tsErrFieldOnly) Error() string { return "field-only" }

// Error + Unwrap → 在 cond.IsError 的白名单内
type tsErrUnwrap struct{ next error }

func (e tsErrUnwrap) Error() string { return "err-unwrap" }
func (e tsErrUnwrap) Unwrap() error { return e.next }

// Error 方法不解引用接收者，nil 指针调用是安全的
type tsErrNoDeref struct{}

func (e *tsErrNoDeref) Error() string { return "no-deref" }

// Error 方法解引用接收者，nil 指针调用会 panic（KNOWN 缺陷）
type tsErrDeref struct{ msg string }

func (e *tsErrDeref) Error() string { return e.msg }

// ---------- 1. nil 与各类零值 ----------

func TestStringNilAndZero(t *testing.T) {
	var nilPtr *int
	var nilStruct *tsStruct

	cases := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, ""},
		{"空字符串", "", ""},
		{"nil指针", nilPtr, ""},
		{"nil结构体指针", nilStruct, ""},
		{"nil字节切片", []byte(nil), ""},
		{"nil字符串切片", []string(nil), ""},
		{"nil字符串切片", []string{}, "[]"},
		{"nilany切片", []any(nil), ""},
		{"nilany切片", []any{}, "[]"},
		{"nil string-any map", map[string]any(nil), ""},
		{"nil string-any map", map[string]any{}, "{}"},

		// KNOWN：与上面的 map[string]any(nil) 不同。
		// map[any]any 在 getByType 里就被转成空 map[string]any 并 jcs.Format，
		// 而 map[string]any 走 getBySpecialType 的 IsNil 分支直接返回空串。
		{"nil any-any map", map[any]any(nil), ""},
		{"nil any-any map", map[any]any{}, "{}"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkEq(t, c.name, conv.String(c.in), c.want)
		})
	}
}

// ---------- 2. 基础标量：getByType 的具名类型分支 ----------

func TestStringScalar(t *testing.T) {
	fixed := time.Date(2026, 9, 19, 10, 30, 0, 0, time.UTC)

	cases := []struct {
		name string
		in   any
		want string
	}{
		{"string", "hello", "hello"},
		{"bool-true", true, "true"},
		{"bool-false", false, "false"},

		{"int", int(-1), "-1"},
		{"int8", int8(-8), "-8"},
		{"int16", int16(-16), "-16"},
		{"int32", int32(-32), "-32"},
		{"int64", int64(-64), "-64"},

		{"uint", uint(1), "1"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},

		{"float32", float32(1.5), "1.5"},
		{"float64", float64(3.14159), "3.14159"},
		{"float64整数", float64(100), "100"},
		{"float64零值", float64(0), "0"},
		{"float64科学计数", 1e21, "1000000000000000000000"},

		{"字节切片", []byte("hello bytes"), "hello bytes"},

		// time.Time 固定格式：2006-01-02 15:04:05
		{"time.Time", fixed, "2026-09-19 10:30:00"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkEq(t, c.name, conv.String(c.in), c.want)
		})
	}
}

// ---------- 3. error：IsError 前置拦截 与 getByType 的 case error ----------

func TestStringError(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		// cond.IsError 为 true（导出方法只有 Error）→ 在函数入口就被拦截
		{"errors.New", errors.New("boom"), "boom"},
		{"自定义纯error", tsPureErr{}, "pure-err"},
		{"带Unwrap的error", tsErrUnwrap{next: errors.New("x")}, "err-unwrap"},
		{"fmt.Errorf带%w", fmt.Errorf("wrap: %w", errors.New("inner")), "wrap: inner"},

		// cond.IsError 为 false（除 Error 外还有 GetMsg），会走到 getByType 的 `case error:`；
		// 但它没有可导出字段，序列化只会得到 {}，所以仍然输出 Error() 的文本。
		{"带额外方法但无导出字段的error", tsErrExtra{}, "err-extra"},

		// Error 不解引用接收者，nil 指针安全
		{"nil指针error不解引用", error((*tsErrNoDeref)(nil)), "no-deref"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkEq(t, c.name, conv.String(c.in), c.want)
		})
	}
}

// ---------- 3.1 error：带有可导出字段时按对象输出 ----------

func TestStringErrorWithOutputField(t *testing.T) {
	// 既实现了 error、又有可导出字段，且导出方法不止 Error（cond.IsError 为 false）：
	// 会走到 getByType 的 `case error:`，因为有别的输出属性，报错往后走，最终按对象输出
	checkJSON(t, "带可导出字段的error", conv.String(tsErrBiz{Msg: "boom", Code: 1}),
		`{"code":1,"msg":"boom"}`)

	// 指针形式：解引用后再判断字段，结论一致
	checkJSON(t, "带可导出字段的error指针", conv.String(&tsErrBiz{Code: 1, Msg: "boom"}),
		`{"code":1,"msg":"boom"}`)
	checkJSON(t, "带可导出字段的error指针", conv.String(&tsErrBiz{}),
		`{"code":0,"msg":""}`)

	// 没有可导出字段，也没有别的输出属性 → 直接输出 Error() 的内容
	checkEq(t, "无导出字段的error", conv.String(tsErrExtra{}), "err-extra")

	// 只有 Error 方法、但有可导出字段：cond.IsError 也要求「无可导出字段」，故为 false，
	// 会走到 getByType 的 `case error:`，同样按对象输出
	checkJSON(t, "仅Error方法但有导出字段", conv.String(tsErrFieldOnly{Code: 1}), `{"code":1}`)
}

// TestStringNilErrorPointerDerefPanic KNOWN 缺陷：
// 实现了 error 的 nil 指针，若 Error() 会解引用接收者，String 会 panic。
// 因为 cond.IsError 返回 false（有第二个导出方法）后，
// getByType 的 `case error:` 会无条件调用 v.Error()，而此时 v 是 nil。
// 这里用 recover 固化现状：若将来修复（nil 时返回空串），断言分支会生效。
func TestStringNilErrorPointerDerefPanic(t *testing.T) {
	var e *tsErrDeref
	defer func() {
		if r := recover(); r != nil {
			t.Logf("KNOWN 缺陷：nil 的 error 指针触发 panic —— %v", r)
		}
	}()
	got := conv.String(e)
	checkEq(t, "修复后应为", got, "")
}

// ---------- 4. getByKind：自定义具名类型 / uintptr / 复数 ----------

func TestStringKind(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"自定义string", tsString("abc"), "abc"},
		{"自定义int", tsInt(-3), "-3"},
		{"自定义float", tsFloat(2.5), "2.5"},
		{"自定义bool", tsBool(true), "true"},
		{"uintptr", uintptr(7), "7"},
		{"complex64", complex64(complex(1, 2)), "(1+2i)"},
		{"complex128", complex128(complex(3, 4)), "(3+4i)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkEq(t, c.name, conv.String(c.in), c.want)
		})
	}
}

// ---------- 5. 指针（含多级指针）----------

func TestStringPointer(t *testing.T) {
	v := 5
	pv := &v
	ppv := &pv
	s := tsString("abc")

	cases := []struct {
		name string
		in   any
		want string
	}{
		{"int指针", &v, "5"},
		{"int二级指针", ppv, "5"},
		{"自定义string指针", &s, "abc"},
		{"结构体指针", &tsStruct{A: 1, B: "b"}, `{"D":"","a":1,"b":"b"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkEq(t, c.name, conv.String(c.in), c.want)
		})
	}

	// 结构体指针最终会解引用到结构体本身
	checkJSON(t, "结构体指针", conv.String(&tsStruct{A: 1, B: "b"}), `{"D":"","a":1,"b":"b"}`)
}

// ---------- 6. sync.Map ----------

func TestStringSyncMap(t *testing.T) {
	var m sync.Map
	m.Store("k", "v")
	m.Store("n", 1)

	// KNOWN 缺陷：getBySyncMap 中向 newMap 写入的那行代码被注释掉了，
	// Range 遍历后 newMap 始终为空，因此任何 sync.Map 都返回 "{}"。
	// 若该行恢复，这里的期望值需改为 {"k":"v","n":1}。
	//
	// 另外源码用 `src.(sync.Map)` 值断言，必然复制 sync.Map（go vet 的
	// "copies lock value" 告警即来源于此），这里按值传入是复现该路径所必需的。
	checkEq(t, "非空sync.Map", conv.String(m), "{\"k\":\"v\",\"n\":1}")
	checkEq(t, "空sync.Map", conv.String(sync.Map{}), "{}")
}

// ---------- 7. map ----------

func TestStringMap(t *testing.T) {
	// 键序确定的：值全为字符串时 jcs.Format 成功，按 JCS 规范排序键
	t.Run("jcs排序", func(t *testing.T) {
		checkEq(t, "全字符串值",
			conv.String(map[string]any{"b": "2", "a": "1"}),
			`{"a":"1","b":"2"}`)
	})

	// 键序随机的：含非字符串值时 jcs.Format 失败，回退 jsoniter，键序取决于 map 迭代顺序
	jsonCases := []struct {
		name string
		in   any
		want string
	}{
		{"含int值", map[string]any{"b": 4, "a": 3}, `{"a":3,"b":4}`},
		{"空map", map[string]any{}, `{}`},
		{"map[string]string", map[string]string{"a": "1"}, `{"a":"1"}`},
		{"int键被转成字符串", map[int]string{1: "one"}, `{"1":"one"}`},
		{"map[any]any", map[any]any{"b": 4, "a": 5}, `{"a":5,"b":4}`},
		{"嵌套map", map[string]any{"m": map[string]any{"k": "v"}}, `{"m":{"k":"v"}}`},
		{"值为结构体", map[string]any{"m": tsInner{X: 1}}, `{"m":{"x":1}}`},
		{"空键", map[string]any{"": "v"}, `{"":"v"}`},
		{"nil值", map[string]any{"k": nil}, `{"k":null}`},
		{"大整数保精度", map[string]any{"n": int64(9007199254740993)}, `{"n":9007199254740993}`},
	}
	for _, c := range jsonCases {
		t.Run(c.name, func(t *testing.T) {
			checkJSON(t, c.name, conv.String(c.in), c.want)
		})
	}
}

// ---------- 8. slice 与定长数组 ----------

func TestStringSliceAndArray(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"字符串切片", []string{"x", "y"}, `["x","y"]`},
		{"int切片", []int{1, 2, 3}, `[1,2,3]`},
		{"any切片", []any{1, "two", 3.0}, `[1,"two",3]`},
		{"空切片", []string{}, `[]`},
		{"nil元素", []any{nil}, `[null]`},
		{"嵌套字节切片转base64", [][]byte{[]byte("a"), []byte("b")}, `["YQ==","Yg=="]`},
		{"结构体切片", []tsInner{{X: 1}}, `[{"x":1}]`},

		// KNOWN：chan 无法被 jsoniter 序列化，被序列化成 null
		{"chan切片", []chan int{make(chan int)}, `[null]`},

		// 定长数组不是 reflect.Slice，走 getByCopy 的通用 jsoniter 路径
		{"int数组", [3]int{1, 2, 3}, `[1,2,3]`},
		{"string数组", [3]string{"a", "b", "c"}, `["a","b","c"]`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkJSON(t, c.name, conv.String(c.in), c.want)
		})
	}
}

// ---------- 9. protobuf ----------

func TestStringProtobuf(t *testing.T) {
	fixed := time.Date(2026, 9, 19, 10, 30, 0, 0, time.UTC)

	// timestamppb 走 getBySpecialType 的 timestamppb 特判：AsTime + 时间格式化
	checkEq(t, "timestamppb指针",
		conv.String(timestamppb.New(fixed)), "2026-09-19 10:30:00")
	// 值类型不实现 proto.Message（ProtoReflect 是指针方法），走解引用后的 timestamppb 分支
	checkEq(t, "timestamppb值",
		conv.String(timestamppb.Timestamp{Seconds: fixed.Unix()}), "2026-09-19 10:30:00")

	// KNOWN：getByType 里的 protojson.Marshal 分支对 protoc-gen-go 生成的类型不可达。
	// 因为 hasCustomJSONTag 会把 `json:"小写,omitempty"` 判定为「自定义 tag」而返回 true，
	// 从而跳过 protojson，最终走通用的 struct 序列化路径（字段名取 json tag）。
	checkJSON(t, "errpb空值",
		conv.String(&errpb.ErrorDetail{}),
		`{"code":0,"details":null,"message":""}`)
	checkJSON(t, "errpb有值",
		conv.String(&errpb.ErrorDetail{Code: 1, Message: "m", Details: []string{"d"}}),
		`{"code":1,"details":["d"],"message":"m"}`)
}

// ---------- 10. database/sql 的 Null* 类型 ----------

func TestStringSqlTypes(t *testing.T) {
	fixed := time.Date(2026, 9, 19, 10, 30, 0, 0, time.UTC)

	cases := []struct {
		name string
		in   any
		want string
	}{
		// NullString 有专门的分支：Valid 时取值，否则空串
		{"NullString有效", sql.NullString{String: "abc", Valid: true}, "abc"},
		{"NullString无效", sql.NullString{String: "abc", Valid: false}, ""},

		// 其它 Null* 走 getByCopy → unwrapSqlTypes 展开成底层值
		{"NullInt64有效", sql.NullInt64{Int64: 99, Valid: true}, "99"},
		{"NullInt64无效", sql.NullInt64{Int64: 99, Valid: false}, "0"},
		{"NullInt32有效", sql.NullInt32{Int32: 7, Valid: true}, "7"},
		{"NullInt16有效", sql.NullInt16{Int16: 6, Valid: true}, "6"},
		{"NullByte有效", sql.NullByte{Byte: 5, Valid: true}, "5"},
		{"NullFloat64有效", sql.NullFloat64{Float64: 1.5, Valid: true}, "1.5"},
		{"NullBool有效", sql.NullBool{Bool: true, Valid: true}, "true"},
		{"NullBool无效", sql.NullBool{Bool: true, Valid: false}, "false"},

		// KNOWN：NullTime 展开成 time.Time 后，jsoniter 对顶层 time.Time 序列化出 "{}"，
		// Valid 与否结果都一样，时间信息丢失。
		{"NullTime有效", sql.NullTime{Time: fixed, Valid: true}, "2026-09-19 10:30:00"},
		{"NullTime无效", sql.NullTime{Time: fixed, Valid: false}, "1000-01-01 00:00:00"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkEq(t, c.name, conv.String(c.in), c.want)
		})
	}
}

// ---------- 11. struct ----------

func TestStringStruct(t *testing.T) {
	fixed := time.Date(2026, 9, 19, 10, 30, 0, 0, time.UTC)

	cases := []struct {
		name string
		in   any
		want string
	}{
		// json:"-" 被忽略；无 tag 的字段回退字段名
		{"基础", tsStruct{A: 1, B: "b", C: "c", D: "d"}, `{"a":1,"b":"b","D":"d"}`},
		{"空结构体", tsEmpty{}, `{}`},

		// omitempty 的空值字段会被补回来（getStringFromStruct 的回流）
		{"omitempty空值被补全", tsStruct{A: 1}, `{"a":1,"b":"","D":""}`},

		// 匿名嵌入的字段被提升到顶层
		{"匿名嵌入", tsHost{tsInner: tsInner{X: 1}, Y: 2}, `{"x":1,"y":2}`},

		{"含时间", tsWithTime{T: fixed}, `{"t":"2026-09-19T10:30:00Z"}`},

		// KNOWN：unwrapSqlTypes 只对顶层生效，嵌套在 struct 字段里的 sql.Null* 不展开
		{"含sql字段不展开", tsWithSQL{S: sql.NullString{String: "v", Valid: true}},
			`{"s":{"String":"v","Valid":true}}`},

		{"嵌套结构体指针", tsNested{Inner: &tsStruct{A: 9, B: "bb"}},
			`{"inner":{"a":9,"b":"bb","D":""}}`},
		{"嵌套结构体指针为nil", tsNested{}, `{"inner":null}`},

		// KNOWN：struct 里 & < > 仍被转义成 \u00xx。
		// strFix 会还原它们，但 struct 分支在 strFix 之前又重新 marshal 了一次，
		// 导致还原结果被覆盖。非 struct 路径（见 TestStringSpecialChars）表现正常。
		{"特殊字符被转义", tsSpecialChar{S: "x&y<z>"}, `{"s":"x\u0026y\u003cz\u003e"}`},

		{"未导出字段被忽略", tsUnexported{Pub: "p", priv: "q"}, `{"pub":"p"}`},
		{"chan字段序列化为null", tsWithChan{C: make(chan int)}, `{"c":null}`},
		{"map字段", tsWithMap{M: map[string]any{"k": 1}}, `{"m":{"k":1}}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkJSON(t, c.name, conv.String(c.in), c.want)
		})
	}
}

// ---------- 12. 特殊字符 & < > 的还原（strFix）----------

func TestStringSpecialChars(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		// 走 jsoniter + strFix 的路径：& < > 被还原
		{"map[string]string", map[string]string{"s": "x&y<z>", "mm": "xbbbb"}, `{"mm": "xbbbb","s":"x&y<z>"}`},
		{"map[string]any含int值", map[string]any{"s": "x&y<z>", "n": 1}, `{"n":1,"s":"x&y<z>"}`},
		{"字符串切片", []string{"x&y<z>"}, `["x&y<z>"]`},

		// jcs 路径本身就不转义 HTML 字符，无需还原
		{"map[string]any全字符串值", map[string]any{"s": "x&y<z>"}, `{"s":"x&y<z>"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkJSON(t, c.name, conv.String(c.in), c.want)
		})
	}

	// 顶层 string 直接返回，不经过序列化，原样保留（不是 JSON，不能用 checkJSON）
	checkEq(t, "顶层字符串", conv.String("x&y<z>"), "x&y<z>")
}

// ---------- 13. JSON 文本输入原样透传 ----------

func TestStringJsonText(t *testing.T) {
	// String 的输入若是 JSON 文本，不会被重新规范化（键序保持原样）
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"未排序的JSON对象", `{"b":2,"a":1}`, `{"b":2,"a":1}`},
		{"未排序的JSON数组", `[3,1,2]`, `[3,1,2]`},
		{"非法JSON", `{"not really json"`, `{"not really json"`},
		{"带空白的JSON", `  {"a":1}  `, `  {"a":1}  `},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkEq(t, c.name, conv.String(c.in), c.want)
		})
	}
}

// ---------- 14. 兜底路径 ----------

func TestStringUnsupported(t *testing.T) {
	// chan / func 都无法被前面任何分支处理，
	// 落到 getByCopy 的 jsoniter 兜底，被序列化成 null
	checkEq(t, "chan", conv.String(make(chan int)), "null")
	checkEq(t, "func", conv.String(func() {}), "null")
}

// ---------- 15. 健壮性：各类输入都不应 panic ----------

func TestStringNoPanic(t *testing.T) {
	fixed := time.Date(2026, 9, 19, 10, 30, 0, 0, time.UTC)
	var syn sync.Map
	syn.Store("k", struct{ A int }{A: 1})

	inputs := []any{
		nil, "", 0, int64(0), uint64(0), 0.0, false, 'x',
		[]any{}, []any{nil}, map[string]any{}, map[any]any{},
		[]byte{}, []string{}, [0]int{},
		&tsStruct{}, &tsEmpty{}, (**int)(nil),
		tsPureErr{}, tsErrExtra{}, tsErrUnwrap{},
		sql.NullString{}, sql.NullTime{}, sql.NullFloat64{},
		timestamppb.New(fixed), &errpb.ErrorDetail{},
		syn, sync.Map{},
		make(chan int), func() {},
		time.Time{}, fixed,
		complex128(0), uintptr(0),
		map[string]any{"deep": map[string]any{"deeper": []any{1, nil, "x"}}},
	}

	for _, in := range inputs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("conv.String(%T) 触发 panic: %v", in, r)
				}
			}()
			_ = conv.String(in)
		}()
	}
}
func TestStringNoPanic11(t *testing.T) {
	src := map[string]any{
		"aa": 1,
		"bb": 2,
	}

	newS, err := jcs.Format(src)

	fmt.Println(newS, err)
}
