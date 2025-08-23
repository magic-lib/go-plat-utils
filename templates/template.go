package templates

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/crypto"
	gCache "github.com/patrickmn/go-cache"
	"regexp"
	"strings"
	tmpl "text/template"
	"time"
)

const (
	prefixDefault = "{{"
	suffixDefault = "}}"
)

var (
	templateCacheMap = gCache.New(30*time.Minute, 100*time.Minute)
)

type template interface {
	Replace(replaceMap ...any) string
	Template(replaceMap any) (string, error)
	ParseVars(templateStrings ...string) []string
}

type impl struct {
	prefix string
	suffix string
	s      string
}

// NewTemplate 新建一个模板
func NewTemplate(s string, fix ...string) template {
	prefix, suffix := getPreAndSuffix(fix...)
	return &impl{prefix: prefix, suffix: suffix, s: s}
}

func getPreAndSuffix(delis ...string) (string, string) {
	startDelim := prefixDefault
	endDelim := suffixDefault

	switch len(delis) {
	case 1:
		if delis[0] != "" {
			startDelim = delis[0]
		}
	case 2:
		if delis[0] != "" {
			startDelim = delis[0]
		}
		if delis[1] != "" {
			endDelim = delis[1]
		}
	}
	return startDelim, endDelim
}

// Replace 替换模板中的变量占位符为实际值
// data: 用于替换的数据源，可以是单个值或多个值的列表
// 返回值: 替换完成后的字符串
func (t *impl) Replace(data ...any) string {
	templateStr := t.s // 原始模板字符串

	// 提取模板中所有的变量占位符
	varKeys := t.ParseVars(templateStr)
	if len(varKeys) == 0 {
		return templateStr // 无占位符直接返回原始字符串
	}

	// 标记是否需要处理全量替换（{{.}}）
	hasFullReplace := false
	for _, key := range varKeys {
		if key == "." {
			hasFullReplace = true
			break
		}
	}

	// 预编译正则表达式模板，避免重复编译
	regexPattern := fmt.Sprintf(`%s\s*%%s\s*%s`,
		regexp.QuoteMeta(t.prefix),
		regexp.QuoteMeta(t.suffix))

	// 遍历所有数据源，替换对应变量
	for _, item := range data {
		// 将数据项转换为JSON字符串，便于通过gjson解析
		itemJSON := conv.String(item)

		for _, key := range varKeys {
			if key == "." {
				continue // 全量替换在后面单独处理
			}
			// 获取JSON中的值
			value, exists := JsonGet(itemJSON, key)
			if !exists {
				continue // 键不存在则跳过
			}

			// 编译当前key的正则表达式并执行替换
			keyRegex := regexp.MustCompile(fmt.Sprintf(regexPattern, regexp.QuoteMeta(key)))
			templateStr = keyRegex.ReplaceAllString(templateStr, value.String())
		}
	}

	// 处理全量替换（{{.}}）
	if hasFullReplace {
		var replaceStr string
		switch len(data) {
		case 1:
			replaceStr = conv.String(data[0])
		case 0:
			replaceStr = "" // 无数据时替换为空
		default:
			replaceStr = conv.String(data) // 多数据时转为数组字符串
		}

		// 编译全量替换的正则表达式并执行替换
		fullReplaceRegex := regexp.MustCompile(fmt.Sprintf(regexPattern, `\.`))
		templateStr = fullReplaceRegex.ReplaceAllString(templateStr, replaceStr)
	}

	return templateStr
}

// ParseVars 从模板字符串中提取所有被分隔符包裹的关键字
// templateStrings: 包含模板关键字的字符串列表
// 返回值: 去重后的模板关键字列表
func (t *impl) ParseVars(templateStrings ...string) []string {
	// 编译正则表达式，用于匹配分隔符包裹的内容
	pattern := fmt.Sprintf(`%s\s*(.*?)\s*%s`, regexp.QuoteMeta(t.prefix), regexp.QuoteMeta(t.suffix))
	matcher := regexp.MustCompile(pattern)

	// 用map存储关键字实现去重
	uniqueKeys := make(map[string]struct{})

	// 遍历所有输入字符串，提取关键字
	for _, str := range templateStrings {
		// 查找所有匹配的内容
		matches := matcher.FindAllStringSubmatch(str, -1)

		for _, match := range matches {
			// 确保有捕获到内容（至少包含整个匹配和分组内容）
			if len(match) < 2 {
				continue
			}

			// 处理关键字：去除首尾空格并检查非空
			key := strings.TrimSpace(match[1])
			if key != "" {
				uniqueKeys[key] = struct{}{}
			}
		}
	}

	// 将map中的关键字转换为切片返回
	keys := make([]string, 0, len(uniqueKeys))
	for key := range uniqueKeys {
		keys = append(keys, key)
	}

	return keys
}

// Template 从模板字符串中提取所有被分隔符包裹的关键字
func (t *impl) Template(data any) (string, error) {
	templateStr := t.s // 原始模板字符串
	return Template(templateStr, data, t.prefix, t.suffix)
}

// Template 模版填充
/*
定义变量：{{$article := "hello"}}  {{$article := .ArticleContent}}
调用方法：{{functionName .arg1 .arg2}}
条件判断：
{{if .condition1}}
{{else if .condition2}}
{{end}}

逻辑关系：
or and not : 或, 与, 非
eq ne lt le gt ge: 等于, 不等于, 小于, 小于等于, 大于, 大于等于
示例：
{{if ge .var1 .var2}}
{{end}}

循环：
{{range $i, $v := .slice}}
{{end}}

{{range .slice}}
{{.field}} //获取对象里的变量
{{$.ArticleContent}}  //访问循环外的全局变量的方式
{{end}}
*/
func Template(format string, data interface{}, delis ...string) (string, error) {
	startDelim, endDelim := getPreAndSuffix(delis...)

	// 生成唯一模板名称（使用MD5哈希避免名称冲突）
	tempName := "utils-template-" + crypto.Md5(format)
	var parsedTpl *tmpl.Template
	var err error

	parsedTplTemp, ok := templateCacheMap.Get(tempName)
	if ok && parsedTplTemp != nil {
		parsedTpl = parsedTplTemp.(*tmpl.Template)
	} else {
		// 创建模板实例并配置
		tpl := tmpl.New(tempName).
			Funcs(sprig.FuncMap()).
			Delims(startDelim, endDelim).
			Option("missingkey=zero") // 缺失键时返回零值而非报错

		// 解析模板
		parsedTpl, err = tpl.Parse(format)
		if err != nil {
			return "", fmt.Errorf("模板解析失败: %w", err)
		}
		templateCacheMap.Set(tempName, parsedTpl, 30*time.Minute)
	}

	// 执行模板渲染
	var buf bytes.Buffer
	if err := parsedTpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("模板渲染失败: %w", err)
	}
	result := buf.String()

	// 如果没有值也不会报错，所以这里需要处理一下
	HasNoValueIndex := strings.Index(result, "<no value>")
	if HasNoValueIndex < 0 {
		return result, nil
	}
	return result, fmt.Errorf("模板有未覆盖的变量")
}
