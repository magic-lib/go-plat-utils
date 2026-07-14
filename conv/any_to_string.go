package conv

import (
	"database/sql"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	jsoniterForNil "github.com/magic-lib/go-plat-utils/internal/jsoniter/go"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/viant/toolbox"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	if cond.IsError(src) {
		return src.(error).Error()
	}

	retStr, err := baseAsString(src)
	if err == nil {
		return retStr
	}

	var ok bool
	src, retStr, ok = getBySpecialType(src)
	if ok {
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
	return toolbox.AsString(src)
}

func baseAsString(src any) (string, error) {
	retStr, err := getByType(src)
	if err == nil {
		return retStr, nil
	}
	retStr, err = getByKind(src)
	if err == nil {
		return retStr, nil
	}
	return "", err
}

func mustBaseAsString(src any) string {
	retStr, err := baseAsString(src)
	if err == nil {
		return retStr
	}
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

func hasCustomJSONTag(msg proto.Message) bool {
	val := reflect.ValueOf(msg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return false
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			tagName := strings.Split(jsonTag, ",")[0]
			if tagName != "" && tagName != field.Name {
				return true
			}
		}
	}

	return false
}

func getBySyncMap(synMap *sync.Map) map[any]any {
	newMap := make(map[any]any)
	defer func() {
		if err := recover(); err != nil {
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

//func isNumeric(data []uint8) bool {
//	if len(data) == 0 {
//		return false
//	}
//	for _, b := range data {
//		if b < '0' || b > '9' {
//			return false
//		}
//	}
//	return true
//}

//func isPrintableASCII(data []uint8) bool {
//	for _, b := range data {
//		if b < 32 || b > 126 {
//			return false
//		}
//	}
//	return true
//}

//func isUTF8String(data []uint8) bool {
//	return utf8.Valid(data)
//}
//
//// 尝试解析为 uint32（大端序）
//func tryParseUint32(data []uint8) (bool, uint32) {
//	if len(data) != 4 {
//		return false, 0
//	}
//	return true, binary.BigEndian.Uint32(data)
//}
//
//// 尝试解析为 float64（大端序）
//func tryParseFloat64(data []uint8) (bool, float64) {
//	if len(data) != 8 {
//		return false, 0
//	}
//	bits := binary.BigEndian.Uint64(data)
//	f := math.Float64frombits(bits)
//	return true, f
//}
//func detectDataType(data []uint8) any {
//	// 优先判断是否为合法的 UTF-8 字符串
//	if isUTF8String(data) {
//		// 检查是否为纯数字字符串
//		if isNumeric(data) {
//			return "number"
//		}
//		return "string"
//	}
//
//	// 尝试解析为二进制数字
//	if len(data) == 4 {
//		if ok, _ := tryParseUint32(data); ok {
//			return "binary uint32"
//		}
//	} else if len(data) == 8 {
//		if ok, f := tryParseFloat64(data); ok {
//			// 检查是否为非 NaN 和非无穷大的有效浮点数
//			if !math.IsNaN(f) && !math.IsInf(f, 0) {
//				return "binary float64"
//			}
//		}
//	}
//
//	// 默认视为二进制数据
//	return "binary data"
//}

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
	if src == nil {
		return "", nil
	}

	switch v := src.(type) {
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case int:
		return strconv.Itoa(v), nil
	case int8:
		return strconv.Itoa(int(v)), nil
	case int16:
		return strconv.Itoa(int(v)), nil
	case int32:
		return strconv.Itoa(int(v)), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	case error:
		return v.Error(), nil
	case time.Time:
		{
			return v.Format(time.DateTime), nil
		}
	case proto.Message:
		{
			if !hasCustomJSONTag(v) {
				b, err := protojson.Marshal(v)
				if err == nil {
					return string(b), nil
				}
			}
		}
	case map[any]any:
		// 直接创建临时map进行类型转换
		stringMap := make(map[string]any)
		for key, value := range v {
			stringMap[String(key)] = value
		}
		if data, err := jsoniterForNil.Marshal(stringMap); err == nil {
			return string(data), nil
		} else {
			return "", err
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

// unwrapSqlTypes 递归展开 sql.Null* 类型为底层值，解决嵌套在 struct/map/slice 中
// 时 JSON 序列化为 {"String":"...","Valid":true} 而非实际值的问题
func unwrapSqlTypes(src any) any {
	if src == nil {
		return nil
	}

	// 直接处理各类 sql.Null* 类型
	switch v := src.(type) {
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return ""
	case sql.NullInt64:
		if v.Valid {
			return v.Int64
		}
		return int64(0)
	case sql.NullFloat64:
		if v.Valid {
			return v.Float64
		}
		return float64(0)
	case sql.NullBool:
		if v.Valid {
			return v.Bool
		}
		return false
	case sql.NullInt32:
		if v.Valid {
			return v.Int32
		}
		return int32(0)
	case sql.NullInt16:
		if v.Valid {
			return v.Int16
		}
		return int16(0)
	case sql.NullByte:
		if v.Valid {
			return v.Byte
		}
		return byte(0)
	case sql.NullTime:
		if v.Valid {
			return v.Time
		}
		return time.Time{}
	}
	return src
}

func getStringFromJson(src any) (string, error) {
	src = unwrapSqlTypes(src)
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
