package conv

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"strings"
)

// KeyListFromMap 参数传字符串，避免不是map结构
// {"app":{"mm":1}} ==> "app.mm" : 1
func KeyListFromMap(keyMapJsonObject any) map[string]any {
	keyMapJson := String(keyMapJsonObject)

	allMap := make(map[string]any)
	if keyMapJson == "" {
		return allMap
	}
	keyMap := make(map[string]any)
	keyList := make([]any, 0)
	var err1, err2 error
	err1 = jsoniter.UnmarshalFromString(keyMapJson, &keyMap)
	if err1 != nil {
		err2 = jsoniter.UnmarshalFromString(keyMapJson, &keyList)
		if err2 == nil {
			toStringFromList(keyList, "", nil, 0, allMap)
		}
	} else {
		toStringFromMap(keyMap, nil, 0, allMap)
	}
	if err1 != nil && err2 != nil {
		allMap["."] = keyMapJson
	}
	return allMap
}

func toStringFromList(oneList []any, lastKey string, keyList []string, index int,
	allMap map[string]any) {
	if keyList == nil {
		keyList = make([]string, 0)
	}
	for i, one := range oneList {
		newKey := fmt.Sprintf("%s[%d]", lastKey, i)
		if target2, ok := one.(map[string]any); ok {
			keyList = append(keyList, newKey)
			index = index + 1
			toStringFromMap(target2, keyList, index, allMap)
			index = index - 1
			keyList = append(keyList[:index])
		} else if target3, ok := one.([]any); ok {
			toStringFromList(target3, newKey, keyList, index, allMap)
		} else {
			keyList = append(keyList, newKey)
			keyStr := strings.Join(keyList, ".")
			allMap[keyStr] = one
			keyList = append(keyList[:index])
		}
	}
}

func toStringFromMap(oneMap map[string]any, keyList []string, index int, allMap map[string]any) {
	if keyList == nil {
		keyList = make([]string, 0)
	}
	for key, val := range oneMap {
		if target, ok := val.(map[string]any); ok {
			keyList = append(keyList, key)
			index = index + 1
			toStringFromMap(target, keyList, index, allMap)
			index = index - 1
			keyList = append(keyList[:index])
		} else {
			if list, ok := val.([]any); ok {
				toStringFromList(list, key, keyList, index, allMap)
				continue
			}
			keyList = append(keyList, key)
			keyStr := strings.Join(keyList, ".")
			allMap[keyStr] = val
			keyList = append(keyList[:index])
		}
	}
	return
}
