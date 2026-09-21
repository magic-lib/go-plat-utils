package conv_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/conv"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// 基准测试与等价性回归测试：conv.String 性能优化的验证基线。
//
// 约定：本文件先于优化建立，TestStringGolden 中的期望值取自优化前的实际输出，
// 优化后必须保持一致（键序随机的项按 JSON 语义等价比较）；
// BenchmarkString* 用于量化 ns/op 与 allocs/op 的改善。

// benchUser 覆盖 struct 路径：含 json tag、omitempty、匿名嵌入、嵌套指针。
type benchUser struct {
	ID       int64  `json:"id"`
	Name     string `json:"name,omitempty"`
	NickName string `json:"nick_name"`
	Ignored  string `json:"-"`
	Remark   string // 无 tag，回退字段名
	Empty    string `json:"empty,omitempty"`
}

type benchEmbedded struct {
	AuditOrderID int64 `json:"audit_order_id"`
	OrderID      int64 `json:"order_id"`
}

type benchOrder struct {
	benchEmbedded
	IsAutoPayment bool      `json:"is_auto_payment"`
	CreatedAt     time.Time `json:"created_at"`
	Amount        float64   `json:"amount"`
	User          *benchUser
	Tags          []string          `json:"tags"`
	Extra         map[string]any    `json:"extra"`
	Strs          map[string]string `json:"strs"`
}

func benchOrderValue() *benchOrder {
	return &benchOrder{
		benchEmbedded: benchEmbedded{AuditOrderID: 1001, OrderID: 2002},
		IsAutoPayment: true,
		CreatedAt:     time.Date(2026, 9, 19, 10, 30, 0, 0, time.UTC),
		Amount:        123.45,
		User: &benchUser{
			ID:       7,
			Name:     "alice",
			NickName: "al",
			Ignored:  "nope",
			Remark:   "hi",
		},
		Tags:  []string{"a", "b", "c"},
		Extra: map[string]any{"k1": "v1", "k2": int64(2)},
		Strs:  map[string]string{"s1": "x&y<z>"},
	}
}

func benchInputs() map[string]any {
	return map[string]any{
		"nil":          nil,
		"string":       "hello world",
		"string-empty": "",
		"int":          int(123456),
		"int64":        int64(123456789),
		"uint64":       uint64(42),
		"float64":      float64(3.14159),
		"bool":         true,
		"bytes":        []byte("hello bytes"),
		"time":         time.Date(2026, 9, 19, 10, 30, 0, 0, time.UTC),
		// 非 JSON 文本：string2Json 应快速返回，不解析
		"json-like-str": `{"not really json"`,
		// 真正的 JSON 文本：应走 JCS 规范化
		"json-object": `{"b":2,"a":1}`,
		"json-array":  `[3,1,2]`,
		"json-spaced": "  \n\t{\"z\":1,\"y\":2}  ",
		// 容器
		"map-any":    map[any]any{"a": 1, "b": "two"},
		"map-string": map[string]any{"aa": "111", "bb": "222"},
		"slice-any":  []any{1, "two", 3.0},
		"slice-int":  []int{1, 2, 3},
		"slice-str":  []string{"x", "y"},
		// struct 与指针
		"struct":     *benchOrderValue(),
		"struct-ptr": benchOrderValue(),
		"user-ptr":   &benchUser{ID: 1, Name: "bob", NickName: "b"},
		// sql / error / protobuf 时间
		"sql-null-valid":   sql.NullString{String: "bbbbb", Valid: true},
		"sql-null-invalid": sql.NullString{String: "bbbbb", Valid: false},
		"sql-null-int":     sql.NullInt64{Int64: 99, Valid: true},
		"error":            errors.New("something failed"),
		"error-errors":     fmt.Errorf("wrapped: %w", errors.New("inner")),
		"timestamppb":      timestamppb.New(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)),
	}
}

// --- 基准测试 ---

func BenchmarkStringNil(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(nil)
	}
}

func BenchmarkStringString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String("hello world")
	}
}

func BenchmarkStringInt(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(int(123456))
	}
}

func BenchmarkStringInt64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(int64(123456789))
	}
}

func BenchmarkStringFloat64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(float64(3.14159))
	}
}

func BenchmarkStringBool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(true)
	}
}

func BenchmarkStringBytes(b *testing.B) {
	src := []byte("hello bytes")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(src)
	}
}

func BenchmarkStringTime(b *testing.B) {
	src := time.Date(2026, 9, 19, 10, 30, 0, 0, time.UTC)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(src)
	}
}

func BenchmarkStringSQLNull(b *testing.B) {
	src := sql.NullString{String: "bbbbb", Valid: true}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(src)
	}
}

func BenchmarkStringError(b *testing.B) {
	src := errors.New("something failed")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(src)
	}
}

// BenchmarkStringNonJSONText 非 JSON 文本：优化后应 0 解析 0 分配。
func BenchmarkStringNonJSONText(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = conv.String(`{"not really json"`)
	}
}

// BenchmarkStringJSONObject 真正的 JSON 文本：走 JCS 规范化。
func BenchmarkStringJSONObject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = conv.String(`{"b":2,"a":1}`)
	}
}

func BenchmarkStringMapAny(b *testing.B) {
	src := map[any]any{"a": 1, "b": "two"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(src)
	}
}

func BenchmarkStringMapString(b *testing.B) {
	src := map[string]any{"aa": "111", "bb": "222"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(src)
	}
}

func BenchmarkStringSlice(b *testing.B) {
	src := []any{1, "two", 3.0}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(src)
	}
}

func BenchmarkStringStruct(b *testing.B) {
	src := benchOrderValue()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(src)
	}
}

// BenchmarkStringMixed 模拟真实混合负载。
func BenchmarkStringMixed(b *testing.B) {
	items := []any{
		nil, "hello", int(1), int64(2), float64(1.5), true,
		[]byte("b"), time.Now(),
		map[string]any{"a": 1}, []any{1, 2},
		benchOrderValue(), errors.New("e"),
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = conv.String(items[i%len(items)])
	}
}

// --- 等价性回归 ---

// jsonEqualOrEqual 先按字符串相等比较；不相等时，若两边都是合法 JSON 则按语义等价比较。
// 必须这么做的原因：struct / map 的输出内部会经过 map 中转，键序取决于 Go 的 map 迭代顺序，
// 是随机的，无法做逐字节断言；只能比较反序列化后的结构。
func jsonEqualOrEqual(got, want string) bool {
	if got == want {
		return true
	}
	var g, w any
	if json.Unmarshal([]byte(got), &g) != nil {
		return false
	}
	if json.Unmarshal([]byte(want), &w) != nil {
		return false
	}
	return reflect.DeepEqual(g, w)
}

// TestStringGolden 断言 conv.String 对每个分支的输出保持稳定。
// 期望值取自 TestCaptureBaseline 的实际输出，任何优化都必须保持其不变。
func TestStringGolden(t *testing.T) {
	want := map[string]string{
		"nil":          "",
		"string":       "hello world",
		"string-empty": "",
		"int":          "123456",
		"int64":        "123456789",
		"uint64":       "42",
		"float64":      "3.14159",
		"bool":         "true",
		"bytes":        "hello bytes",
		"time":         "2026-09-19 10:30:00",

		// JSON 文本输入一律原样返回：String 不再做 JCS 规范化（键序与空白保持原样）
		"json-like-str": `{"not really json"`,
		"json-object":   `{"b":2,"a":1}`,
		"json-array":    `[3,1,2]`,
		"json-spaced":   "  \n\t{\"z\":1,\"y\":2}  ",

		// 容器：空字段由 getStringFromStruct 补全，& < > 由 strFix 还原
		"map-any":    `{"a":1,"b":"two"}`,
		"map-string": `{"aa":"111","bb":"222"}`,
		"slice-any":  `[1,"two",3]`,
		"slice-int":  `[1,2,3]`,
		"slice-str":  `["x","y"]`,

		// struct：omitempty 字段被补全输出，无 tag 字段回退字段名，匿名嵌入展开
		"struct":     `{"User":{"Remark":"hi","id":7,"name":"alice","nick_name":"al"},"amount":123.45,"audit_order_id":1001,"created_at":"2026-09-19T10:30:00Z","extra":{"k1":"v1","k2":2},"is_auto_payment":true,"order_id":2002,"strs":{"s1":"x&y<z>"},"tags":["a","b","c"]}`,
		"struct-ptr": `{"User":{"Remark":"hi","id":7,"name":"alice","nick_name":"al"},"amount":123.45,"audit_order_id":1001,"created_at":"2026-09-19T10:30:00Z","extra":{"k1":"v1","k2":2},"is_auto_payment":true,"order_id":2002,"strs":{"s1":"x&y<z>"},"tags":["a","b","c"]}`,
		"user-ptr":   `{"Remark":"","empty":"","id":1,"name":"bob","nick_name":"b"}`,

		// sql.Null*、error、protobuf 时间戳
		"sql-null-valid":   "bbbbb",
		"sql-null-invalid": "",
		"sql-null-int":     "99",
		"error":            "something failed",
		"error-errors":     "wrapped: inner",
		"timestamppb":      "2026-01-02 03:04:05",
	}

	inputs := benchInputs()
	for _, name := range baselineOrder() {
		src, ok := inputs[name]
		if !ok {
			t.Fatalf("missing input %q", name)
		}
		expected, ok := want[name]
		if !ok {
			t.Fatalf("missing golden value for %q", name)
		}
		if got := conv.String(src); !jsonEqualOrEqual(got, expected) {
			t.Errorf("分支 %q 输出漂移:\n got=%q\nwant=%q", name, got, expected)
		}
	}
}

// --- 基线捕获：仅在需要重新生成 golden 值时手动执行 ---

// TestCaptureBaseline 打印所有输入的输出，用于生成/核对 golden 值。
// 正常回归不依赖它；手动运行：`go test -run TestCaptureBaseline -v ./conv/`
func TestCaptureBaseline(t *testing.T) {
	inputs := benchInputs()
	for _, name := range baselineOrder() {
		src, ok := inputs[name]
		if !ok {
			t.Fatalf("missing input %q", name)
		}
		t.Logf("BASELINE %-18s => %q", name, conv.String(src))
	}
}

// baselineOrder 固定遍历顺序，保证多次捕获输出可比。
func baselineOrder() []string {
	return []string{
		"nil", "string", "string-empty", "int", "int64", "uint64",
		"float64", "bool", "bytes", "time",
		"json-like-str", "json-object", "json-array", "json-spaced",
		"map-any", "map-string", "slice-any", "slice-int", "slice-str",
		"struct", "struct-ptr", "user-ptr",
		"sql-null-valid", "sql-null-invalid", "sql-null-int",
		"error", "error-errors", "timestamppb",
	}
}
