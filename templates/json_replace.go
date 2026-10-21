package templates

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// compileRe 按 pattern 缓存已编译的正则，避免模板替换时重复编译相同的 prefix/suffix 模式
var reCache sync.Map // map[string]*regexp.Regexp

func compileRe(pattern string) *regexp.Regexp {
	if v, ok := reCache.Load(pattern); ok {
		return v.(*regexp.Regexp)
	}
	re := regexp.MustCompile(pattern)
	reCache.Store(pattern, re)
	return re
}

type JsonMapTemplate struct {
	prefixString string
	suffixString string
	PathSplitter string // 路径分割符，默认是 /
}

func NewJsonMapTemplate(fixString ...string) *JsonMapTemplate {
	prefixString := DefaultPrefix
	suffixString := DefaultSuffix
	if len(fixString) > 1 {
		prefixString = fixString[0]
		suffixString = fixString[1]
	}
	return &JsonMapTemplate{
		prefixString: prefixString,
		suffixString: suffixString,
		PathSplitter: "/",
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

// ReplacePath 替换路径
// 路径变量替换：把 /api/project/:project/env/:env/redis-config 中的变量替换为实际值
// 支持 :name（单段）和 *name（通配）
func (j *JsonMapTemplate) ReplacePath(path string, params map[string]any) (string, error) {
	segs := strings.Split(path, j.PathSplitter)
	var err error
	for i, seg := range segs {
		if len(seg) > 1 {
			oneSeg := trimTemplateSpaces(j.prefixString, j.suffixString, seg)
			if !hasTemplatePatternStrict(j.prefixString, j.suffixString, oneSeg) {
				continue
			}
			isFind := false
			for k, v := range params {
				newKey := fmt.Sprintf("%s%s%s", j.prefixString, k, j.suffixString)
				if newKey == oneSeg {
					segs[i] = conv.String(v)
					isFind = true
					break
				}
			}
			if !isFind {
				err = multierror.Append(err, fmt.Errorf("path: %s, param: %s not found", path, seg))
			}
		}
	}
	return strings.Join(segs, j.PathSplitter), err
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
	re := compileRe(compileStr)
	return re.ReplaceAllString(input, fmt.Sprintf("%s$1%s", prefixString, suffixString))
}

func hasTemplatePatternStrict(prefixString, suffixString, input string) bool {
	escapedPrefix := regexp.QuoteMeta(prefixString)
	escapedSuffix := regexp.QuoteMeta(suffixString)
	pattern := fmt.Sprintf(`%s\s*\S+?\s*%s`, escapedPrefix, escapedSuffix)
	return compileRe(pattern).MatchString(input)
}
