package templates

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	"reflect"
	"regexp"
)

type JsonMapTemplate struct {
	prefixString string
	suffixString string
}

func NewJsonMapTemplate(fixString ...string) *JsonMapTemplate {
	prefixString := prefixDefault
	suffixString := suffixDefault
	if len(fixString) > 1 {
		prefixString = fixString[0]
		suffixString = fixString[1]
	}
	if prefixString == "" {
		prefixString = prefixDefault
	}
	if suffixString == "" {
		suffixString = suffixDefault
	}
	return &JsonMapTemplate{
		prefixString: prefixString,
		suffixString: suffixString,
	}
}

func (j *JsonMapTemplate) Replace(args any, bindings ...map[string]any) (any, error) {
	argsStr := conv.String(args)
	allParamStrRet := trimTemplateSpaces(j.prefixString, j.suffixString, argsStr)

	if !hasTemplatePatternStrict(j.prefixString, j.suffixString, argsStr) {
		return args, nil
	}

	var err error
	lo.ForEachWhile(bindings, func(binding map[string]any, _ int) bool {
		tmp := NewTemplate(allParamStrRet, j.prefixString, j.suffixString)
		allParamStrRet = tmp.Replace(binding)
		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return args, fmt.Errorf("JsonMapTemplate Replace: %w", err)
	}
	if cond.IsPointer(args) {
		_ = conv.Unmarshal(allParamStrRet, args)
		return args, nil
	}

	retInfo, err := conv.ConvertForType(reflect.TypeOf(args), allParamStrRet)
	if err == nil {
		return retInfo, nil
	}
	return args, nil
}

// trimTemplateSpaces 去除{{后的空格和}}前的空格
func trimTemplateSpaces(prefixString, suffixString string, input string) string {
	// 正则表达式解释：
	// {{\s+  匹配{{后面跟一个或多个空白字符
	// (\S.*?) 捕获非空白字符开始的内容（非贪婪模式）
	// \s+}}  匹配一个或多个空白字符后面跟}}
	escapedPrefix := regexp.QuoteMeta(prefixString)
	escapedSuffix := regexp.QuoteMeta(suffixString)
	compileStr := fmt.Sprintf(`%s\s+(\S.*?)\s+%s`, escapedPrefix, escapedSuffix)
	re := regexp.MustCompile(compileStr)
	return re.ReplaceAllString(input, fmt.Sprintf("%s$1%s", prefixString, suffixString))
}

func hasTemplatePatternStrict(prefixString, suffixString, input string) bool {
	escapedPrefix := regexp.QuoteMeta(prefixString)
	escapedSuffix := regexp.QuoteMeta(suffixString)
	pattern := fmt.Sprintf(`%s\s*\S+?\s*%s`, escapedPrefix, escapedSuffix)
	matched, _ := regexp.MatchString(pattern, input)
	return matched
}
