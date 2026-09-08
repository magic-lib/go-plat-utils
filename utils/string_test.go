package utils_test

import (
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

// TestReplaceDynamicVariables 验证将分隔符包裹的动态变量替换为指定格式
func TestReplaceDynamicVariables(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		start  string
		end    string
		repPre string
		repEnd string
		want   string
	}{
		{
			name:   "多变量基本替换",
			input:  "年龄大于${age}且小于${max}",
			start:  "${",
			end:    "}",
			repPre: "[",
			repEnd: "]",
			want:   "年龄大于[age]且小于[max]",
		},
		{
			name:   "变量位于开头",
			input:  "${name} 的余额",
			start:  "${",
			end:    "}",
			repPre: "[",
			repEnd: "]",
			want:   "[name] 的余额",
		},
		{
			name:   "变量位于结尾",
			input:  "状态为${status}",
			start:  "${",
			end:    "}",
			repPre: "[",
			repEnd: "]",
			want:   "状态为[status]",
		},
		{
			name:   "整串仅为单个变量",
			input:  "${user.name}",
			start:  "${",
			end:    "}",
			repPre: "[",
			repEnd: "]",
			want:   "[user.name]",
		},
		{
			name:   "连续变量",
			input:  "(${a})-(${b})",
			start:  "${",
			end:    "}",
			repPre: "[",
			repEnd: "]",
			want:   "([a])-([b])",
		},
		{
			name:   "反向替换：方括号变量还原为${}",
			input:  "[age] > 18 && [name] == 'a'",
			start:  "[",
			end:    "]",
			repPre: "${",
			repEnd: "}",
			want:   "${age} > 18 && ${name} == 'a'",
		},
		{
			name:   "分隔符为正则特殊字符(*)",
			input:  "单价*price*元，数量*count*",
			start:  "*",
			end:    "*",
			repPre: "[",
			repEnd: "]",
			want:   "单价[price]元，数量[count]",
		},
		{
			name:   "多字符模板分隔符",
			input:  "欢迎<%=user%>光临",
			start:  "<%=",
			end:    "%>",
			repPre: "${",
			repEnd: "}",
			want:   "欢迎${user}光临",
		},
		{
			name:   "含正则元字符的多字符分隔符",
			input:  "金额$$(amount)元",
			start:  "$$(",
			end:    ")",
			repPre: "[",
			repEnd: "]",
			want:   "金额[amount]元",
		},
		{
			name:   "无变量原样返回",
			input:  "没有变量的文本",
			start:  "${",
			end:    "}",
			repPre: "[",
			repEnd: "]",
			want:   "没有变量的文本",
		},
		{
			name:   "空字符串",
			input:  "",
			start:  "${",
			end:    "}",
			repPre: "[",
			repEnd: "]",
			want:   "",
		},
		{
			name:   "起始分隔符存在但未闭合不替换",
			input:  "余额${amount 未闭合",
			start:  "${",
			end:    "}",
			repPre: "[",
			repEnd: "]",
			want:   "余额${amount 未闭合",
		},
		{
			name:   "仅文本中出现分隔符片段",
			input:  "a$b 与 {c} 是普通字符",
			start:  "${",
			end:    "}",
			repPre: "[",
			repEnd: "]",
			want:   "a$b 与 {c} 是普通字符",
		},
	}
	for _, c := range cases {
		got := utils.ReplaceDynamicVariables(c.input, c.start, c.end, c.repPre, c.repEnd)
		if got != c.want {
			t.Errorf("%s: ReplaceDynamicVariables(%q, %q, %q, %q, %q)=%q, want %q",
				c.name, c.input, c.start, c.end, c.repPre, c.repEnd, got, c.want)
		}
	}
}
