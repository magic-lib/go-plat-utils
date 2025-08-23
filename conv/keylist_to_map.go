package conv

import (
	"github.com/json-iterator/go"
	"regexp"
	"strconv"
	"strings"
)

// MapFromKeyList "app.mm" : 1 ==> {"app":{"mm":1}}
func MapFromKeyList(keyMapJsonObject any) map[string]any {
	keyMapJson := String(keyMapJsonObject)

	keyMap := make(map[string]any)
	_ = jsoniter.UnmarshalFromString(keyMapJson, &keyMap)
	newMap := make(map[string]any)
	for key, val := range keyMap {
		keyList := strings.Split(key, ".")
		toMapFromString(keyList, val, newMap)
	}
	return newMap
}

func toMapFromString(keyList []string, val any, oneMap map[string]any) {
	var re, _ = regexp.Compile(`\[[0-9]+]$`)

	for i, key := range keyList {
		if key == "" {
			continue
		}
		//如果要显示的数组
		index := re.FindString(key)
		realKey := key
		isArray := false
		isEnd := i == (len(keyList) - 1)
		var indexNumber int
		if index != "" {
			realKey = strings.Replace(key, index, "", -1)
			isArray = true
			inStr := strings.ReplaceAll(index, "[", "")
			inStr = strings.ReplaceAll(inStr, "]", "")
			indexNumber, _ = strconv.Atoi(inStr)
		}

		if !isArray {
			if isEnd {
				if val != nil {
					oneMap[realKey] = val
				}
				continue
			}
			if _, ok := oneMap[realKey]; !ok {
				oneMap[realKey] = make(map[string]any)
			}
			tempMap := oneMap[realKey]
			if one, ok := tempMap.(map[string]any); ok {
				oneMap = one
			}
		} else {
			if _, ok := oneMap[realKey]; !ok {
				oneMap[realKey] = make([]any, 0)
			}
			if arr, ok := oneMap[realKey].([]any); ok {
				if len(arr) <= indexNumber {
					newArr := make([]any, indexNumber+1)
					copy(newArr, arr)
					arr = newArr
				}

				if isEnd {
					arr[indexNumber] = val
					oneMap[realKey] = arr
					continue
				}

				var target map[string]any
				if one, ok := arr[indexNumber].(map[string]any); ok {
					target = one
				} else {
					arr[indexNumber] = make(map[string]any)
					target, _ = arr[indexNumber].(map[string]any)
				}

				oneMap[realKey] = arr
				oneMap = target
			}
		}

	}
	return
}
