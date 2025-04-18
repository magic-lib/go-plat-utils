package conv

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conf"
	jsoniterForNil "github.com/magic-lib/go-plat-utils/internal/jsoniter/go"
	"github.com/samber/lo"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// String 转换为string
func String(src any) string {
	if src == nil {
		return ""
	}

	strType := reflect.TypeOf(src)
	strValue := reflect.ValueOf(src)
	if strType.Kind() == reflect.Ptr {
		if strValue.IsNil() {
			return ""
		}
		return String(strValue.Elem().Interface())
	}

	// 常用特殊类型
	if strValue.Type().String() == "sync.Map" {
		retStr := ""
		if synMap, ok := src.(sync.Map); ok {
			retStr = String(getBySyncMap(&synMap))
		}
		return retStr
	}

	if strType.Kind() == reflect.Map {
		if strValue.IsNil() {
			return ""
		}
		retStr, newMap, err := getByMap(src)
		if err == nil {
			return retStr
		}
		src = newMap
	}

	if strType.Kind() == reflect.Slice {
		if strValue.IsNil() {
			return ""
		}
		retStr, newList, err := getBySlice(src)
		if err == nil {
			return retStr
		}
		src = newList
	}

	retStr, err := getByKind(src)
	if err == nil {
		return retStr
	}

	retStr, err = getByType(src)
	if err == nil {
		return retStr
	}

	retStr, err = getByTypeString(src)
	if err == nil {
		return retStr
	}

	retStr, err = getByCopy(src) //concurrent map read and map write
	if err == nil {
		return retStr
	}

	fmt.Printf("jsoniter.Marshal error:%s", err.Error())
	return fmt.Sprintf("%v", src)
}

func getBySyncMap(synMap *sync.Map) map[any]any {
	newMap := make(map[any]any)
	defer func() {
		if err := recover(); any(err) != nil {
			fmt.Println("getBySyncMap error:", err)
			return
		}
	}()
	fmt.Println("getBySyncMap 1:")
	synMap.Range(func(key, value any) bool {
		fmt.Println("getBySyncMap 2:")
		//newMap[key] = value
		return true
	})
	fmt.Println("getBySyncMap 3:")
	return newMap
}
func getByMap(src any) (string, map[any]any, error) {
	retStr, err := getStringFromJson(src)
	if err == nil {
		return retStr, nil, nil
	}

	strValue := reflect.ValueOf(src)

	newMap := make(map[any]any)
	iter := strValue.MapRange()
	for iter.Next() {
		newMap[iter.Key().Interface()] = iter.Value().Interface()
	}

	retStr, err = getStringFromJson(newMap)
	if err == nil {
		return retStr, newMap, nil
	}

	return "", newMap, err
}
func getBySlice(src any) (string, []any, error) {
	//如果是[]byte，则直接转为string
	if strByte, ok := src.([]byte); ok {
		return string(strByte), nil, nil
	}

	json, err := getStringFromJson(src)
	if err == nil {
		return json, nil, nil
	}
	strValue := reflect.ValueOf(src)

	newMap := make([]any, 0)
	for i := 0; i < strValue.Len(); i++ {
		oneItem := strValue.Index(i).Interface()
		newMap = append(newMap, oneItem)
	}

	retStr, err := getStringFromJson(newMap)
	if err == nil {
		return retStr, newMap, nil
	}

	return "", newMap, err
}

func getByKind(i any) (string, error) {
	if i == nil {
		return "", nil
	}

	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Complex64:
		return fmt.Sprintf("(%g+%gi)", real(v.Complex()), imag(v.Complex())), nil
	case reflect.Complex128:
		return fmt.Sprintf("(%g+%gi)", real(v.Complex()), imag(v.Complex())), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	default:
		return "", fmt.Errorf("kind error")
	}
}

func getByType(src any) (string, error) {
	switch src.(type) {
	case []byte:
		return string(src.([]byte)), nil
	case byte:
		return string(src.(byte)), nil
	case string:
		return src.(string), nil
	case int:
		return strconv.Itoa(src.(int)), nil
	case int64:
		return strconv.FormatInt(src.(int64), 10), nil
	case error:
		err, _ := src.(error)
		return err.Error(), nil
	//case float64:
	//	return strconv.FormatFloat(str.(float64), 'g', -1, 64)
	case time.Time:
		{
			oneTime := src.(time.Time)
			//如果为空时间，则返回空字符串
			if cond.IsTimeEmpty(oneTime) {
				//return "", nil
			}
			if cst := conf.TimeLocation(); cst != nil {
				return oneTime.In(cst).Format(fullTimeForm), nil
			}
			return oneTime.Format(fullTimeForm), nil
		}
	}
	return "", fmt.Errorf("type error")
}

func getByTypeString(src any) (string, error) {
	strType := fmt.Sprintf("%T", src)
	if strType == "errors.errorString" {
		errTemp := fmt.Sprintf("%v", src)
		if len(errTemp) <= 2 {
			return "", nil
		}
		return errTemp[1 : len(errTemp)-1], nil
	}

	//看看是否是数组类型
	if len(strType) >= 2 {
		subTemp := lo.Substring(strType, 0, 2)
		if subTemp == "[]" && strType != "[]string" {
			arrTemp := reflect.ValueOf(src)
			newArrTemp := make([]any, 0)
			for i := 0; i < arrTemp.Len(); i++ {
				oneTemp := arrTemp.Index(i).Interface()
				newArrTemp = append(newArrTemp, oneTemp)
			}
			retStr, _, err := getBySlice(newArrTemp)
			return retStr, err
		}
	}

	return "", fmt.Errorf("typeString error")
}
func getByCopy(src any) (string, error) {
	newStrTemp := mapDeepCopy(src) //concurrent map read and map write

	retStr, err := getStringFromJson(newStrTemp)
	if err == nil {
		return retStr, nil
	}
	return "", fmt.Errorf("copy error")
}

func getStringFromJson(src any) (string, error) {
	json, err := jsoniterForNil.MarshalToString(src)
	if err == nil {
		if len(json) >= 2 { //解决返回字符串首位带"的问题
			match, errTemp := regexp.MatchString(`^".*"$`, json)
			if errTemp == nil {
				if match {
					json = json[1 : len(json)-1]
				}
			}
		}
		//解决 & 会转换成 \u0026 的问题
		return strFix(json), nil
	}
	//nolint:goerr113
	return "", fmt.Errorf(errStrGetStringFromJson, err)
}

func strFix(s string) string {
	// https://stackoverflow.com/questions/28595664/how-to-stop-json-marshal-from-escaping-and/28596225
	if strings.Contains(s, "\\u0026") {
		s = strings.Replace(s, "\\u0026", "&", -1)
	}
	if strings.Contains(s, "\\u003c") {
		s = strings.Replace(s, "\\u003c", "<", -1)
	}
	if strings.Contains(s, "\\u003e") {
		s = strings.Replace(s, "\\u003e", ">", -1)
	}
	return s
}

func mapDeepCopy(value any) any {
	switch v := value.(type) {
	case map[string]any:
		newMap := make(map[string]any)
		for k, v := range v {
			newMap[k] = mapDeepCopy(v)
		}
		return newMap
	case []any:
		newSlice := make([]any, len(v))
		for k, v := range v {
			newSlice[k] = mapDeepCopy(v)
		}
		return newSlice
	default:
		return value
	}
}
