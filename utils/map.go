package utils

import (
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/json-iterator/go"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"log"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// FilterMap map有很多字段，需要通过Mapstruct过滤，主要用在查询数据库时，
// 因为前端传的参数可能很多，需要过滤出数据库包含的字段才可以用
func FilterMap(oldMap map[string]interface{}, mapStruct interface{}) (map[string]interface{}, []string) {
	modelMapStr, err := jsoniter.Marshal(mapStruct)
	if err != nil {
		return oldMap, nil
	}
	newMap1 := make(map[string]interface{})
	err = jsoniter.Unmarshal(modelMapStr, &newMap1)
	if err != nil {
		return oldMap, nil
	}
	newMap := make(map[string]interface{})
	keyList := make([]string, 0)
	for key := range oldMap {
		isFind := false
		for key2 := range newMap1 {
			if key == key2 {
				isFind = true
				break
			}
		}
		if isFind {
			newMap[key] = oldMap[key]
			keyList = append(keyList, key)
		}
	}
	return newMap, keyList
}

// FillMap 将oldMap中的数据填充到mapStruct中
func FillMap(oldMap map[string]interface{}, mapStruct interface{}) interface{} {
	modelMapStr, err := jsoniter.Marshal(mapStruct)
	if err != nil {
		return mapStruct
	}
	newMapAll := make(map[string]interface{})
	err = jsoniter.Unmarshal(modelMapStr, &newMapAll)
	if err != nil {
		return mapStruct
	}

	for key := range oldMap {
		for key2 := range newMapAll {
			if key == key2 {
				newMapAll[key] = oldMap[key]
				break
			}
		}
	}

	modelMapAllStr, err := jsoniter.Marshal(newMapAll)
	if err != nil {
		return mapStruct
	}
	err = jsoniter.Unmarshal(modelMapAllStr, mapStruct)
	if err != nil {
		return mapStruct
	}
	return mapStruct
}

// ShowColumn 将一个数组中很多字段，过滤出showList里包含的字段，返回map或数组
func ShowColumn(data interface{}, showList []string) interface{} {
	if data == nil {
		return nil
	}
	oldStr := reflect.TypeOf(data).Kind()
	if oldStr == reflect.Struct || oldStr == reflect.Map {
		byte1, err := jsoniter.Marshal(data)
		if err == nil {
			retData := make(map[string]interface{}, 0)
			byte2 := string(byte1)
			for _, one := range showList {
				retData[one] = gjson.Get(byte2, one).Value()
			}
			return &retData
		}
	}

	if oldStr == reflect.Slice {
		byte1, err := jsoniter.Marshal(data)
		if err == nil {
			retData := make([]map[string]interface{}, 0)
			byte2 := string(byte1)

			lens := gjson.Get(byte2, "#").Int()
			for i := 0; i < int(lens); i++ {
				tempData := make(map[string]interface{}, 0)
				retData = append(retData, tempData)
			}

			for _, one := range showList {
				tempList := gjson.Get(byte2, "#."+one).Array()
				for i, two := range tempList {
					retData[i][one] = two.Value()
				}
			}
			return &retData
		}
	}

	log.Println(oldStr)

	return nil
}

// HideColumn 与上相反
func HideColumn(data interface{}, hideList []string) interface{} {
	if data == nil {
		return nil
	}

	oldStr := reflect.TypeOf(data).Kind()
	if oldStr == reflect.Struct || oldStr == reflect.Map {
		byte1, err := jsoniter.Marshal(data)
		if err == nil {
			retData := make(map[string]interface{}, 0)
			byte2 := string(byte1)
			for _, one := range hideList {
				byte2, _ = sjson.Delete(byte2, one)
			}
			_ = jsoniter.UnmarshalFromString(byte2, &retData)
			return &retData
		}
	}

	if oldStr == reflect.Slice {
		byte1, err := jsoniter.Marshal(data)
		if err == nil {
			retData := make([]map[string]interface{}, 0)
			byte2 := string(byte1)

			lens := gjson.Get(byte2, "#").Int()
			for i := 0; i < int(lens); i++ {
				for _, one := range hideList {
					byte2, _ = sjson.Delete(byte2, strconv.Itoa(i)+"."+one)
				}
			}

			_ = jsoniter.UnmarshalFromString(byte2, &retData)
			return &retData
		}
	}

	log.Println(oldStr)

	return nil
}

// MapSort 按key排序 isDesc 是否降序
func MapSort(oldData map[string]interface{}, isDesc ...bool) map[string]interface{} {
	if oldData == nil {
		return nil
	}
	o := orderedmap.New()
	for k, v := range oldData {
		o.Set(k, v)
	}
	desc := false
	if len(isDesc) >= 1 {
		desc = isDesc[0]
	}

	if !desc {
		o.SortKeys(sort.Strings)
	} else {
		o.SortKeys(func(keys []string) {
			sort.Slice(keys, func(i, j int) bool {
				return keys[i] > keys[j]
			})
		})
	}
	var newMap map[string]interface{}
	err := conv.Unmarshal(o, &newMap)
	if err != nil {
		return oldData
	}
	return newMap
}

// GetJsonKeyMaps 取得一个struct的所有key，批量输出到前端使用
func GetJsonKeyMaps(bean interface{}) map[string]string {
	cType := reflect.TypeOf(bean)
	cValue := reflect.ValueOf(bean)
	if cValue.Kind() != reflect.Ptr {
		return nil
	}

	nunLen := reflect.Indirect(cValue)
	if nunLen.Kind() == reflect.Interface {
		return nil
	}

	jsonMap := make(map[string]string, 0)
	for i := 0; i < nunLen.NumField(); i++ {
		field := cType.Elem().Field(i)
		name := field.Tag.Get("json")
		if name == "-" || name == "" {
			name = ChangeVariableName(field.Name, "lower")
		}
		jsonMap[field.Name] = name
	}
	return jsonMap
}

// DelKey 批量删除map里多个字段
func DelKey(val map[string]interface{}, fields []string) map[string]interface{} {
	if fields == nil || len(fields) == 0 {
		return val
	}
	for _, key := range fields {
		delete(val, key)
	}
	return val
}

// ToMapFromKeyList "app.mm" : 1 ==> {"app":{"mm":1}}
func ToMapFromKeyList(keyMapJsonObject interface{}) map[string]interface{} {
	keyMapJson := conv.String(keyMapJsonObject)

	keyMap := make(map[string]interface{})
	_ = jsoniter.UnmarshalFromString(keyMapJson, &keyMap)
	newMap := make(map[string]interface{})
	for key, val := range keyMap {
		keyList := strings.Split(key, ".")
		toMapFromString(keyList, val, newMap)
	}
	return newMap
}

func toMapFromString(keyList []string, val interface{}, oneMap map[string]interface{}) {
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
				oneMap[realKey] = make(map[string]interface{})
			}
			tempMap := oneMap[realKey]
			if one, ok := tempMap.(map[string]interface{}); ok {
				oneMap = one
			}
		} else {
			if _, ok := oneMap[realKey]; !ok {
				oneMap[realKey] = make([]interface{}, 0)
			}
			if arr, ok := oneMap[realKey].([]interface{}); ok {
				if len(arr) <= indexNumber {
					newArr := make([]interface{}, indexNumber+1)
					copy(newArr, arr)
					arr = newArr
				}

				if isEnd {
					arr[indexNumber] = val
					oneMap[realKey] = arr
					continue
				}

				var target map[string]interface{}
				if one, ok := arr[indexNumber].(map[string]interface{}); ok {
					target = one
				} else {
					arr[indexNumber] = make(map[string]interface{})
					target, _ = arr[indexNumber].(map[string]interface{})
				}

				oneMap[realKey] = arr
				oneMap = target
			}
		}

	}
	return
}

// GetStructInfoByTag converts golang struct field into slice string.
// tagNames 会依次按传的顺序获取
func GetStructInfoByTag(in any, f func(string) string, tagNames ...string) (structName string, fieldMap map[string]interface{}, err error) {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil, fmt.Errorf("input is a nil pointer")
		}
		v = v.Elem()
	}
	// we only accept structs
	if v.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("ToMap only accepts structs; got %T", v)
	}
	if f == nil {
		f = func(s string) string {
			return s
		}
	}

	newTagNames := make([]string, 0)
	if tagNames != nil && len(tagNames) > 0 {
		lo.ForEach(tagNames, func(item string, index int) {
			item = strings.TrimSpace(item)
			if item != "" {
				newTagNames = append(newTagNames, item)
			}
		})
	}

	typ := v.Type()
	structName = typ.Name()

	fieldMap = make(map[string]any)

	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		vi := v.Field(i)

		//首字母大写的才能导出
		if strings.ToUpper(fi.Name[:1]) != fi.Name[:1] {
			continue
		}

		oneKey := ""
		findTagName := false
		if len(newTagNames) > 0 {
			for j := 0; j < len(newTagNames); j++ {
				tempTag := strings.TrimSpace(newTagNames[j])
				name, err := getOneTag(fi, tempTag)
				if err != nil {
					continue
				}
				findTagName = true
				oneKey = name
				break
			}
		}
		if !findTagName {
			oneKey = f(fi.Name)
		}

		if oneKey != "" { //有可能tag设置为-的情况
			if vi.IsValid() {
				fieldMap[oneKey] = vi.Interface()
			}
		}
	}
	return
}

func getOneTag(fi reflect.StructField, oneTag string) (string, error) {
	if oneTag == "" {
		return fi.Name, nil
	}
	tagV := fi.Tag.Get(oneTag)
	if strings.Contains(tagV, ",") {
		tagV = strings.TrimSpace(strings.Split(tagV, ",")[0])
	}
	if tagV == "-" {
		return "", nil //表示不能用
	}
	if tagV == "" {
		return fi.Name, fmt.Errorf("tag %s not found", oneTag)
	}
	return tagV, nil
}

// StructInfo 用于存储结构体的通用信息
type StructInfo struct {
	PackageName string
	TypeName    string
	Fields      []FieldInfo
}

// FieldInfo 用于存储结构体字段的信息
type FieldInfo struct {
	Name string
	Type string
	Tag  string
}

// GetStructInfo 函数用于获取结构体的详细信息
func GetStructInfo(obj interface{}) StructInfo {
	// 获取对象的反射类型
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 获取包名和类型名
	pkgPath := t.PkgPath()
	if pkgPath == "" {
		// 如果是未命名包，使用默认值
		pkgPath = ""
	}
	pkgName := runtime.FuncForPC(reflect.ValueOf(obj).Pointer()).Name()
	if lastSlash := len(pkgPath) - 1; lastSlash > 0 {
		pkgName = pkgPath[lastSlash+1:]
	}
	typeName := t.Name()

	// 遍历结构体的字段
	var fields []FieldInfo
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldInfo := FieldInfo{
			Name: field.Name,
			Type: field.Type.String(),
			Tag:  string(field.Tag),
		}
		fields = append(fields, fieldInfo)
	}

	return StructInfo{
		PackageName: pkgName,
		TypeName:    typeName,
		Fields:      fields,
	}
}
