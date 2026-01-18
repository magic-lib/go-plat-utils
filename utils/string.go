package utils

import (
	"fmt"
	"github.com/samber/lo"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/magic-lib/go-plat-utils/internal"
)

// VariableType snake 和 camel 类型
type VariableType string

const (
	Snake  = VariableType(internal.Snake)
	Camel  = VariableType(internal.Camel)
	Pascal = VariableType(internal.Pascal)
	Lower  = VariableType(internal.Lower)
	Upper  = VariableType(internal.Upper)
)

// RandomString 生成随机字符串
func RandomString(l int, sourceStr ...string) string {
	if l <= 0 {
		return ""
	}
	var str = append(lo.NumbersCharset, lo.LowerCaseLettersCharset...)
	if len(sourceStr) > 0 {
		sourceArr := make([]rune, 0, len(sourceStr))
		for _, one := range sourceStr {
			sourceArr = append(sourceArr, []rune(one)...)
		}
		if len(sourceArr) > 0 {
			str = sourceArr
		}
	}
	return lo.RandomString(l, str)
}

func RandomStringInt(l int) string {
	if l <= 0 {
		return ""
	}
	return lo.RandomString(l, append(lo.NumbersCharset))
}

// UnicodeDecodeString 解码unicode
func UnicodeDecodeString(s string) string {
	if s == "" {
		return s
	}
	newStr := make([]string, 0)
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		newStr = append(newStr, fmt.Sprintf("%c", r))
		i += n
	}
	if len(newStr) == 0 {
		return s
	}
	return strings.Join(newStr, "")
}

// VarNameConverter 将驼峰与小写互转
func VarNameConverter(varName string, toType ...VariableType) string {
	if len(toType) == 1 {
		inType := internal.VariableType(toType[0])
		return internal.VarNameConverter(varName, inType)
	}
	return internal.VarNameConverter(varName)
}

// ReplaceDynamicVariables 将动态分隔符包裹的变量替换为 [变量名] 格式
// 参数：
// - input：原始字符串
// - startDelimiter：起始分隔符
// - endDelimiter：结束分隔符
func ReplaceDynamicVariables(input string, startDelimiter, endDelimiter string, replaceStartDelimiter, replaceEndDelimiter string) string {
	// 转义分隔符中的特殊字符，确保它们在正则表达式中被正确处理
	escapedStart := regexp.QuoteMeta(startDelimiter)
	escapedEnd := regexp.QuoteMeta(endDelimiter)

	// 构建正则表达式：匹配 startDelimiter + 变量名 + endDelimiter
	regexPattern := fmt.Sprintf(`%s(.*?)%s`, escapedStart, escapedEnd)
	re := regexp.MustCompile(regexPattern)

	// 替换匹配到的变量
	output := re.ReplaceAllStringFunc(input, func(match string) string {
		// 提取变量名（去掉分隔符）
		varName := strings.TrimPrefix(strings.TrimSuffix(match, endDelimiter), startDelimiter)
		return fmt.Sprintf("%s%s%s", replaceStartDelimiter, varName, replaceEndDelimiter)
	})

	return output
}

// LeftPadding 字符串左侧补全
func LeftPadding(length int, str string, complement rune) string {
	if len(str) >= length {
		return str
	}
	padding := strings.Repeat(string(complement), length-len(str))
	return fmt.Sprintf("%s%s", padding, str)
}
