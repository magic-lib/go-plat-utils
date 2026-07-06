package activity

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/samber/lo"
	"log"
	"regexp"
)

var (
	activityIdRegExpMap = cmap.New[*regexp.Regexp]() //缓存，提高效率
)

func CloneMap(data map[string]any) map[string]any {
	m := make(map[string]any)
	err := conv.Unmarshal(conv.String(data), &m)
	if err == nil {
		return m
	}
	return lo.Assign(data)
}

func MergeNewArguments(oldArgs map[string]any, args map[string]any) map[string]any {
	if len(args) == 0 {
		return oldArgs
	}
	jsonTemplate := templates.NewJsonMapTemplate()
	newArgs, err := jsonTemplate.Replace(oldArgs, args)
	if err == nil {
		_ = conv.Unmarshal(newArgs, &oldArgs)
	}
	for key, val := range args {
		if _, ok := oldArgs[key]; !ok {
			oldArgs[key] = val
		}
	}
	return oldArgs
}

func getResponseRegExp(keyPrefix string) *regexp.Regexp {
	re, ok := activityIdRegExpMap.Get(keyPrefix)
	if ok && re != nil {
		return re
	}
	retStr := fmt.Sprintf(`\{\{%s([a-zA-Z0-9_]+)\.(%s|%s)\.`, keyPrefix, Arguments, Responses)
	re = regexp.MustCompile(retStr)
	activityIdRegExpMap.Set(keyPrefix, re)
	return re
}

func logDebug(str ...any) {
	if !cond.IsOpenLog() {
		return
	}
	strArr := make([]any, 0)
	strArr = append(strArr, "[logDebug]")
	strArr = append(strArr, str...)
	log.Println(strArr...)
}

// ExtractDependsActivityIds 解析出待依赖的所有ActivityId
func ExtractDependsActivityIds(condition string, keyPrefix string) []string {
	idList := make([]string, 0)
	re := getResponseRegExp(keyPrefix)
	matches := re.FindAllStringSubmatch(condition, -1)
	for _, match := range matches {
		if len(match) > 1 {
			idList = append(idList, match[1])
		}
	}
	return lo.Uniq(idList)
}
