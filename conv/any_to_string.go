package conv

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/magic-lib/go-plat-utils/cond"
	jsoniterForNil "github.com/magic-lib/go-plat-utils/internal/jsoniter/go"
	"github.com/spf13/cast"
	"github.com/ucarion/jcs"
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

var (
	// quotedJSONRegexp 预编译：getStringFromJson 位于 map/slice/struct 的必经路径，
	// 原先每次调用都执行 regexp.MatchString 重新编译正则（约 1-2µs 且多次分配）。
	quotedJSONRegexp = regexp.MustCompile(`^".*"$`)

	// 以下为静态哨兵错误。所有调用点仅判断 err == nil，从不使用错误文本，
	// 因此用包级变量代替 fmt.Errorf，可消除失败路径上的字符串与 error 分配
	// （struct 分支单次转换会连续触发其中 4-5 次失败）。
	errKind       = errors.New("kind error")
	errType       = errors.New("type error")
	errSQLType    = errors.New("sql type error")
	errTypeString = errors.New("typeString error")
	errCopy       = errors.New("copy error")
)

const MysqlZeroTime = "1000-01-01 00:00:00"

// String 转换为string
func String(src any) string {
	if src == nil {
		return ""
	}

	// 特殊处理error类型
	if cond.IsError(src) {
		return src.(error).Error()
	}

	// IgnoreOmitempty 为 true 时忽略 json tag 中的 omitempty，空值字段也会输出（String 显示全部字段）
	ignoreOmitempty := true

	retStr, err := baseAsString(src)
	if err == nil {
		return retStr
	}

	var ok bool
	src, retStr, ok = getBySpecialType(src, ignoreOmitempty)
	if ok {
		return retStr
	}

	retStr, err = getBySqlType(src)
	if err == nil {
		return retStr
	}

	retStr, err = getByTypeString(src, ignoreOmitempty)
	if err == nil {
		return retStr
	}

	retStr, err = getByCopy(src, ignoreOmitempty) //concurrent map read and map write
	if err == nil {
		return retStr
	}
	retStr, err = cast.ToStringE(src)
	if err == nil {
		return retStr
	}

	retStr = toolbox.AsString(src)
	return retStr
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

func getBySpecialType(src any, ignoreOmitempty bool) (any, string, bool) {
	strType := reflect.TypeOf(src)
	strValue := reflect.ValueOf(src)
	if strType.Kind() == reflect.Ptr {
		if strValue.IsNil() {
			return src, "", true
		}
		return src, String(strValue.Elem().Interface()), true
	}

	// 常用特殊类型：直接用类型断言判定，避免 strValue.Type().String()
	// 每次生成完整类型名字符串（一次分配 + 字符串比较）后再与紧随其后的断言重复判断
	if synMap, ok := src.(sync.Map); ok {
		return src, String(getBySyncMap(&synMap)), true
	}

	if strType.Kind() == reflect.Map {
		if strValue.IsNil() {
			return src, "", true
		}
		retStr, newMap, err := getByMap(src, ignoreOmitempty)
		if err == nil {
			return src, retStr, true
		}
		src = newMap
	}

	if strType.Kind() == reflect.Slice {
		if strValue.IsNil() {
			return src, "", true
		}
		retStr, newList, err := getBySlice(src, ignoreOmitempty)
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
			// 用 Index 截断代替 strings.Split，避免为每个字段分配切片
			tagName := jsonTag
			if idx := strings.Index(tagName, ","); idx >= 0 {
				tagName = tagName[:idx]
			}
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
	synMap.Range(func(key, value any) bool {
		newMap[key] = value
		return true
	})
	return newMap
}
func getByMap(src any, ignoreOmitempty bool) (string, map[any]any, error) {
	retStr, err := getStringFromJson(src, ignoreOmitempty)
	if err == nil {
		return retStr, nil, nil
	}

	strValue := reflect.ValueOf(src)

	newMap := make(map[any]any)
	iter := strValue.MapRange()
	for iter.Next() {
		newMap[iter.Key().Interface()] = iter.Value().Interface()
	}

	retStr, err = getStringFromJson(newMap, ignoreOmitempty)
	if err == nil {
		return retStr, newMap, nil
	}

	return "", newMap, err
}
func getBySlice(src any, ignoreOmitempty bool) (string, []any, error) {
	//如果是[]byte，则直接转为string
	if strByte, ok := src.([]byte); ok {
		return string(strByte), nil, nil
	}

	jsonStr, err := getStringFromJson(src, ignoreOmitempty)
	if err == nil {
		return jsonStr, nil, nil
	}
	strValue := reflect.ValueOf(src)

	newMap := make([]any, 0)
	for i := 0; i < strValue.Len(); i++ {
		oneItem := strValue.Index(i).Interface()
		newMap = append(newMap, oneItem)
	}

	retStr, err := getStringFromJson(newMap, ignoreOmitempty)
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
		return "", errKind
	}
}

// outputFieldCache 缓存 hasOutputField 的判定结果，key 为 reflect.Type，value 为 bool。
// 判定只与动态类型有关，缓存后可避免每次转换都遍历结构体字段。
var outputFieldCache sync.Map

// hasOutputField 判断 val（解引用后）是否含有「可导出且会被输出」的字段，json:"-" 的字段不算。
// 用于区分「纯粹的错误对象」与「碰巧实现了 error 接口的业务结构体」：
//   - 业务结构体带有可导出字段，序列化后是有内容的 JSON，此时应当按对象输出；
//   - 纯粹的错误对象（errors.New、fmt.Errorf 等）没有可导出字段，
//     序列化只会得到 {}，此时输出 Error() 的文本才有意义。
func hasOutputField(val any) bool {
	v := reflect.ValueOf(val)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	if cached, ok := outputFieldCache.Load(v.Type()); ok {
		return cached.(bool)
	}
	t := v.Type()
	has := false
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() || f.Tag.Get("json") == "-" {
			continue
		}
		has = true
		break
	}
	outputFieldCache.Store(t, has)
	return has
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
	case error:
		// 如果该对象包含了别的可导出的值，则打印这个值，否则输出v.Error()内容
		if hasOutputField(v) {
			// 还有可导出的字段，说明这是「碰巧实现了 error 的业务对象」而不是纯粹的错误，
			// 返回 errType 让流程继续往后走，最终按对象序列化输出
			return "", errType
		}
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
		if cond.IsNil(src) {
			return "", nil
		}
		// 直接创建临时map进行类型转换
		stringMap := make(map[string]any, len(v))
		for key, value := range v {
			stringMap[String(key)] = value
		}
		if newS, err := jcs.Format(stringMap); err == nil {
			return newS, nil
		}
		if data, err := jsoniterForNil.Marshal(stringMap); err == nil {
			return string(data), nil
		} else {
			return "", err
		}
	}

	return "", errType
}
func getBySqlType(src any) (string, error) {
	if strNull, ok := src.(sql.NullString); ok {
		if strNull.Valid {
			return strNull.String, nil
		}
		return "", nil
	}

	return "", errSQLType
}

func getByTypeString(src any, ignoreOmitempty bool) (string, error) {
	strType := reflect.TypeOf(src).String()
	if strType == "errors.errorString" {
		errTemp := fmt.Sprintf("%v", src)
		if len(errTemp) <= 2 {
			return "", nil
		}
		return errTemp[1 : len(errTemp)-1], nil
	}

	//看看是否是数组类型
	if len(strType) >= 2 {
		subTemp := strType[:2]
		if subTemp == "[]" && strType != "[]string" {
			arrTemp := reflect.ValueOf(src)
			newArrTemp := make([]any, 0, arrTemp.Len())
			for i := 0; i < arrTemp.Len(); i++ {
				oneTemp := arrTemp.Index(i).Interface()
				newArrTemp = append(newArrTemp, oneTemp)
			}
			retStr, _, err := getBySlice(newArrTemp, ignoreOmitempty)
			return retStr, err
		}
	}

	return "", errTypeString
}
func getByCopy(src any, ignoreOmitempty bool) (string, error) {
	newStrTemp := mapDeepCopy(src) //concurrent map read and map write

	retStr, err := getStringFromJson(newStrTemp, ignoreOmitempty)
	if err == nil {
		return retStr, nil
	}
	return "", errCopy
}

// unwrapSqlTypes 递归展开 sql.Null* 类型为底层值，解决嵌套在 struct/map/slice 中
// 时 JSON 序列化为 {"String":"...","Valid":true} 而非实际值的问题
func unwrapSqlTypes(src any) (any, bool) {
	if src == nil {
		return nil, false
	}

	// 直接处理各类 sql.Null* 类型
	switch v := src.(type) {
	case sql.NullString:
		if v.Valid {
			return v.String, true
		}
		return "", true
	case sql.NullInt64:
		if v.Valid {
			return v.Int64, true
		}
		return int64(0), true
	case sql.NullFloat64:
		if v.Valid {
			return v.Float64, true
		}
		return float64(0), true
	case sql.NullBool:
		if v.Valid {
			return v.Bool, true
		}
		return false, true
	case sql.NullInt32:
		if v.Valid {
			return v.Int32, true
		}
		return int32(0), true
	case sql.NullInt16:
		if v.Valid {
			return v.Int16, true
		}
		return int16(0), true
	case sql.NullByte:
		if v.Valid {
			return v.Byte, true
		}
		return byte(0), true
	case sql.NullTime:
		if v.Valid {
			return v.Time, true
		}
		return MysqlZeroTime, true
	}
	return src, false
}

func getStringFromJson(src any, ignoreOmitempty bool) (string, error) {
	_, ok1 := src.(map[string]any)
	_, ok2 := src.(map[string]string)
	_, ok3 := src.(map[any]any)
	_, ok4 := src.([]any)
	if ok1 || ok2 || ok3 || ok4 {
		if newS, err := jcs.Format(src); err == nil {
			return newS, nil
		} else {
			if ok1 {
				omIns := orderedmap.New()
				for k, v := range src.(map[string]any) {
					omIns.Set(k, v)
				}
				bs, err := json.Marshal(omIns)
				if err == nil {
					return string(bs), nil
				}
			}
			if ok2 {
				omIns := orderedmap.New()
				for k, v := range src.(map[string]string) {
					omIns.Set(k, v)
				}
				bs, err := json.MarshalIndent(omIns, "", "")
				if err == nil {
					return string(bs), nil
				}
			}
		}
	}

	src, ok1 = unwrapSqlTypes(src)
	if ok1 {
		return String(src), nil
	}
	jsonStr, err := jsoniterForNil.MarshalToString(src)
	if err == nil {
		//解决返回字符串首位带"的问题
		if len(jsonStr) >= 2 && quotedJSONRegexp.MatchString(jsonStr) {
			jsonStr = jsonStr[1 : len(jsonStr)-1]
		}
		//解决 & 会转换成 \u0026 的问题
		retAll := strFix(jsonStr)
		if ignoreOmitempty {
			v := reflect.ValueOf(src)
			if v.Kind() == reflect.Struct {
				newMapAll := make(map[string]any)
				err = jsoniterForNil.UnmarshalFromString(retAll, &newMapAll)
				if err != nil {
					return retAll, err
				}
				newMap := getStringFromStruct(src, newMapAll)
				if newS, err := jcs.Format(newMap); err == nil {
					return newS, nil
				}
				return jsoniterForNil.MarshalToString(newMap)
			}
		}
		return retAll, nil
	}
	return "", fmt.Errorf(errStrGetStringFromJson, err)
}

// 解决src为struct时，存在json里有 omitempty 时，会隐藏不输出的问题
func getStringFromStruct(obj any, newMap map[string]any) map[string]any {
	if newMap == nil {
		newMap = make(map[string]any)
	}
	if obj == nil {
		return newMap
	}
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Struct {
		if v.Kind() == reflect.Ptr && v.IsValid() && !v.IsNil() {
			return getStringFromStruct(v.Elem().Interface(), newMap)
		}
		return newMap
	}
	t := v.Type()
	// 遍历所有字段，先解决匿名嵌入的字段，因为外层的优先级更高一些，所有后面再次覆盖
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // 非导出字段跳过
			continue
		}
		// 判断是否是匿名嵌入的对象
		if !field.Anonymous {
			continue
		}
		newMapTemp := make(map[string]any)
		newMapTemp = getStringFromStruct(v.Field(i).Interface(), newMapTemp)
		for k, v := range newMapTemp {
			newMap[k] = v
		}
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // 非导出字段跳过
			continue
		}
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		key := jsonTag
		if idx := strings.Index(key, ","); idx >= 0 {
			key = key[:idx] // 去掉 ,omitempty 等后缀
		}
		if key == "" {
			key = field.Name // 无 json name 时回退用字段名
		}
		if _, exists := newMap[key]; exists {
			continue // 已存在则不覆盖
		}
		// 判断是否是匿名嵌入的对象
		if field.Anonymous {
			continue
		}
		fv := v.Field(i)
		newMap[key] = fv.Interface()
	}
	return newMap
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
		newMap := make(map[string]any, len(v))
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
