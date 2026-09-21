package internal_test

import (
	"encoding/json"
	"github.com/magic-lib/go-plat-utils/internal"
	"strings"
	"testing"

	"github.com/magic-lib/go-plat-utils/cond"
)

// TestSimpleCheckJson 各种合法/非法输入
func TestSimpleCheckJson(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		// ---- 合法：应为 true ----
		{"空对象", `{}`, true},
		{"空数组", `[]`, true},
		{"简单对象", `{"a":1}`, true},
		{"字符串值", `{"a":"b"}`, true},
		{"多键值", `{"a":1,"b":"x","c":true,"d":false,"e":null}`, true},
		{"简单数组", `[1,2,3]`, true},
		{"对象数组", `[{"a":1},{"b":2}]`, true},
		{"嵌套混合", `{"a":{"b":[1,2,{"c":null}]}}`, true},
		{"数组嵌套空值", `[[]]`, true},
		{"前后空白", `  {"a":1}  `, true},
		{"换行制表空白", "\n\t{\"a\":1}\r\n", true},
		{"字符串内含闭合符", `{"a":"}]"}`, true},
		{"转义反斜杠", `{"a":"\\"}`, true},
		{"转义引号", `{"a":"\""}`, true},
		{"unicode转义", `{"a":"你好"}`, true},
		{"科学计数", `{"a":1e10}`, true},
		{"负数小数", `{"a":-1.5}`, true},
		{"中文字段名", `{"键":"值"}`, true},

		// ---- 非法：应为 false ----
		{"空串", ``, false},
		{"纯空白", `   `, false},
		{"仅左括号", `{`, false},
		{"仅右括号", `}`, false},
		{"顶层数字", `123`, false},
		{"顶层字符串", `"str"`, false},
		{"顶层true", `true`, false},
		{"顶层null", `null`, false},
		{"花括号配方括号", `{]`, false},
		{"方括号配花括号", `[}`, false},
		{"对象尾逗号", `{"a":1,}`, false},
		{"数组尾逗号", `[1,2,]`, false},
		{"对象起始逗号", `{,}`, false},
		{"数组起始逗号", `[,]`, false},
		{"重复逗号", `{"a":1,,}`, false},
		{"缺失逗号", `{"a" 1}`, false},
		{"重复冒号", `{"a"::1}`, false},
		{"尾冒号", `{"a":}`, false},
		{"数组内冒号", `[:1]`, false},
		{"对象未闭合", `{"a":1`, false},
		{"数组未闭合", `[1,2`, false},
		{"尾随内容", `{"a":1} {"b":2}`, false},
		{"尾随字符", `{"a":1}extra`, false},
		{"多余闭合符", `[]]`, false},
		{"数组元素缺逗号", `[1 2]`, false},
		{"对象元素缺逗号", `{"a":1 "b":2}`, false},
		{"两个顶层值", `{"a":1} "x"`, false},
		{"串外控制字符", "{\"a\":\x01}", false},
		{"串内控制字符", "{\"a\":\"\x01\"}", false},
		{"引号未闭合", `{"a":"b}`, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := internal.SimpleCheckJsonArrayOrObject(c.in); got != c.want {
				t.Errorf("SimpleCheckJsonArrayOrObject(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

// TestSimpleCheckJsonDeepNesting 嵌套深度超过内部栈数组容量时会扩容，结果应不受影响
func TestSimpleCheckJsonDeepNesting(t *testing.T) {
	for depth := 1; depth <= 64; depth *= 2 {
		text := strings.Repeat(`{"a":`, depth) + `{}` + strings.Repeat(`}`, depth)
		var v any
		if err := json.Unmarshal([]byte(text), &v); err != nil {
			t.Fatalf("用例自身有误，嵌套%d层不是合法JSON: %v", depth, err)
		}
		if got := internal.SimpleCheckJsonArrayOrObject(text); !got {
			t.Errorf("SimpleCheckJson 嵌套%d层 = false, want true", depth)
		}

		// 少一个闭合符 => 非法
		bad := strings.Repeat(`{"a":`, depth) + `{}` + strings.Repeat(`}`, depth-1)
		if got := internal.SimpleCheckJsonArrayOrObject(bad); got {
			t.Errorf("SimpleCheckJson 嵌套%d层缺闭合符 = true, want false", depth)
		}
	}
}

// TestSimpleCheckJsonNoFalseNegative 交叉验证：不存在假阴性。
// 凡 json.Unmarshal 能成功且顶层是对象/数组的输入，SimpleCheckJson 必须返回 true。
// 这是该方法的核心正确性保证——它只能比 Unmarshal 宽松，不能更严格。
func TestSimpleCheckJsonNoFalseNegative(t *testing.T) {
	valids := []string{
		`{}`, `[]`, `{"a":1}`, `[1]`, `{"a":[]}`, `[{}]`,
		`{"a":{"b":{"c":[1,2,{"d":null}]}},"e":"文字"}`,
		`[1,2,3,4,5]`, `{"a":true,"b":false,"c":null,"d":-1.25e+3}`,
		"{\"a\":\t\"b\",\n\"c\":  [ 1 , 2 ]  }",
		`{"url":"https://x.com/a?b=1&c=2"}`,
		`{"regex":"[,{}\":\\]","tail":"]}"}`, // 字符串内满是结构性字符
		`{"emoji":"😀"}`,
		`{"a":"{}","b":"[]","c":"[}"}`,
	}

	for _, s := range valids {
		var v any
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			t.Fatalf("用例自身有误，%q 不是合法JSON: %v", s, err)
		}
		switch v.(type) {
		case map[string]any, []any:
		default:
			t.Fatalf("用例自身有误，%q 顶层不是对象或数组", s)
		}
		if got := internal.SimpleCheckJsonArrayOrObject(s); !got {
			t.Errorf("假阴性：Unmarshal 成功但 SimpleCheckJson(%q) = false", s)
		}
	}
}

// TestSimpleCheckJsonFilterPower 验证预筛确实能在调用 Unmarshal 之前拦下常见非法输入
func TestSimpleCheckJsonFilterPower(t *testing.T) {
	mustReject := []string{
		`hello world`,
		`<html><body>x</body></html>`,
		`2024-01-01`,
		`{"a":1,}`,
		`[1,2,]`,
		`{"a":1} trailing`,
		`{"a" 1}`,
		`not-a-json-string-but-long-enough-to-be-a-realistic-payload`,
	}
	for _, s := range mustReject {
		var v any
		if err := json.Unmarshal([]byte(s), &v); err == nil {
			t.Errorf("用例自身有误，%q 竟然被当成合法JSON", s)
		}
		if got := internal.SimpleCheckJsonArrayOrObject(s); got {
			t.Errorf("SimpleCheckJson(%q) = true, 应在 Unmarshal 之前被拦截", s)
		}
	}
}

// ---- 基准测试：对比 SimpleCheckJson 与 json.Unmarshal 的开销 ----

var (
	benchValid   = `{"name":"tom","age":18,"tags":["a","b","c"],"addr":{"city":"sh","zip":200000}}`
	benchInvalid = `this is definitely not a json payload and it goes on and on ...`
)

func BenchmarkSimpleCheckJsonValid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = internal.SimpleCheckJsonArrayOrObject(benchValid)
	}
}

func BenchmarkJsonUnmarshalValid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var v any
		_ = json.Unmarshal([]byte(benchValid), &v)
	}
}

func BenchmarkSimpleCheckJsonInvalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = internal.SimpleCheckJsonArrayOrObject(benchInvalid)
	}
}

func BenchmarkJsonUnmarshalInvalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var v any
		_ = json.Unmarshal([]byte(benchInvalid), &v)
	}
}

// BenchmarkIsJsonInvalid 走 getAnyFromJsonString 的完整路径：预筛 + Unmarshal
func BenchmarkIsJsonInvalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = cond.IsJson(benchInvalid)
	}
}
