package conv

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conf"
	jsoniterForNil "github.com/magic-lib/go-plat-utils/internal/jsoniter/go"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// String 转换为string
func String(src any) string {
	if src == nil {
		return ""
	}

	var retStr string
	var ok bool
	src, retStr, ok = getBySpecialType(src)
	if ok {
		return retStr
	}

	retStr, err := getByKind(src)
	if err == nil {
		return retStr
	}

	retStr, err = getByType(src)
	if err == nil {
		return retStr
	}

	retStr, err = getBySqlType(src)
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
	retStr, err = cast.ToStringE(src)
	if err == nil {
		return retStr
	}

	fmt.Printf("jsoniter.Marshal error:%s", err.Error())
	return fmt.Sprintf("%v", src)
}

func getBySpecialType(src any) (any, string, bool) {
	strType := reflect.TypeOf(src)
	strValue := reflect.ValueOf(src)
	if strType.Kind() == reflect.Ptr {
		if strValue.IsNil() {
			return src, "", true
		}
		return src, String(strValue.Elem().Interface()), true
	}

	// 常用特殊类型
	if strValue.Type().String() == "sync.Map" {
		retStr := ""
		if synMap, ok := src.(sync.Map); ok {
			retStr = String(getBySyncMap(&synMap))
		}
		return src, retStr, true
	}

	if strType.Kind() == reflect.Map {
		if strValue.IsNil() {
			return src, "", true
		}
		retStr, newMap, err := getByMap(src)
		if err == nil {
			return src, retStr, true
		}
		src = newMap
	}

	if strType.Kind() == reflect.Slice {
		if strValue.IsNil() {
			return src, "", true
		}
		retStr, newList, err := getBySlice(src)
		if err == nil {
			return src, retStr, true
		}
		src = newList
	}

	strValue = reflect.ValueOf(src)
	for strValue.Kind() == reflect.Ptr {
		strValue = strValue.Elem()
		if !strValue.IsValid() {
			break
		}
		src = strValue.Interface()
		if src == nil {
			break
		}
		strValue = reflect.ValueOf(src)
	}
	if src != nil {
		if v, ok := src.(timestamppb.Timestamp); ok {
			src = v.AsTime()
			return src, String(src), true
		}
	}

	return src, "", false
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

func isNumeric(data []uint8) bool {
	if len(data) == 0 {
		return false
	}
	for _, b := range data {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func isPrintableASCII(data []uint8) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

func isUTF8String(data []uint8) bool {
	return utf8.Valid(data)
}

// 尝试解析为 uint32（大端序）
func tryParseUint32(data []uint8) (bool, uint32) {
	if len(data) != 4 {
		return false, 0
	}
	return true, binary.BigEndian.Uint32(data)
}

// 尝试解析为 float64（大端序）
func tryParseFloat64(data []uint8) (bool, float64) {
	if len(data) != 8 {
		return false, 0
	}
	bits := binary.BigEndian.Uint64(data)
	f := math.Float64frombits(bits)
	return true, f
}
func detectDataType(data []uint8) any {
	// 优先判断是否为合法的 UTF-8 字符串
	if isUTF8String(data) {
		// 检查是否为纯数字字符串
		if isNumeric(data) {
			return "number"
		}
		return "string"
	}

	// 尝试解析为二进制数字
	if len(data) == 4 {
		if ok, _ := tryParseUint32(data); ok {
			return "binary uint32"
		}
	} else if len(data) == 8 {
		if ok, f := tryParseFloat64(data); ok {
			// 检查是否为非 NaN 和非无穷大的有效浮点数
			if !math.IsNaN(f) && !math.IsInf(f, 0) {
				return "binary float64"
			}
		}
	}

	// 默认视为二进制数据
	return "binary data"
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
				// 将时间转换为字符串时，如果加了时区，则显示的结果和预期不一致了，所以这里不处理
				//return oneTime.In(cst).Format(fullTimeForm), nil
			}
			return oneTime.Format(fullTimeForm), nil
		}
	}
	return "", fmt.Errorf("type error")
}
func getBySqlType(src any) (string, error) {
	if strNull, ok := src.(sql.NullString); ok {
		if strNull.Valid {
			return strNull.String, nil
		}
		return "", nil
	}

	return "", fmt.Errorf("sql type error")
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
