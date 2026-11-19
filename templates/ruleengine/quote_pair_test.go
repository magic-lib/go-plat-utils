package ruleengine

import "testing"

// 字符串只在遇到与开头相同的引号时结束
func TestStringQuotePairing(t *testing.T) {
	r := NewEngineLogic()
	params := map[string]any{"title": `say "hi"`, "word": "it's", "backslash": `a\`}
	cases := []struct {
		rule string
		want any
	}{
		{`"it's" == word`, true},         // 双引号里直接写单引号
		{`'say "hi"' == title`, true},    // 单引号里直接写双引号
		{`JavaScript("1 == '1'")`, true}, // 脚本里的引号不用再转义
		{`JavaScript('"a" + 5')`, "a5"},
		{`JavaScript('1 == \"1\"')`, true}, // 原来的转义写法仍然可用
		{`Contains('a"b', '"')`, true},
		// 与外层相同的引号转义后不会截断字符串，继续往后找真正的结束引号
		{`'it\'s' == word`, true},
		{`"say \"hi\"" == title`, true},
		{`JavaScript('1 == \'1\'')`, true},
		{`JavaScript("\"a\" + 5")`, "a5"},
		{`Contains('a\'b\'c', "'")`, true},
		{`'a\\' == backslash`, true},  // 转义反斜杠本身：'a\\' 表示 a\，结尾的引号正常结束
		{`'abc' == "abc"`, true},      // 两种引号的字符串值相同
		{`In('b', ['a', "b"])`, true}, // 数组里混用两种引号
	}
	for _, c := range cases {
		ret, err := r.EvaluateString(c.rule, params)
		if err != nil {
			t.Fatalf("%s: %v", c.rule, err)
		}
		if ret != c.want {
			t.Errorf("%s = %#v, want %#v", c.rule, ret, c.want)
		}
	}
	// 引号不配对时报未闭合错误
	for _, rule := range []string{`'abc" == title`, `"abc' == title`} {
		if _, err := r.EvaluateString(rule, params); err == nil {
			t.Errorf("%s: expected unclosed string error", rule)
		}
	}
}
