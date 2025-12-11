package conv

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/magic-lib/go-plat-utils/cond"
	jsoniterForNil "github.com/magic-lib/go-plat-utils/internal/jsoniter/go"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

/*
* Deprecated: 该方法已废弃，请使用 conv.Unmarshal
1、目前不能解决继承为小写的情况
2、已有值了，填充没有的情况
*/
func AssignTo(srcStruct any, dstPoint any) error {
	return toAssignTo(srcStruct, dstPoint)
}
func toAssignTo(srcStruct any, dstPoint any) error {
	srcVal := reflect.ValueOf(srcStruct)
	dstVal := reflect.ValueOf(dstPoint)
	// 检查 dst 是否为指针
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		return fmt.Errorf("dstPoint error nil or not point:%s, %v", reflect.TypeOf(dstPoint).String(), srcStruct)
	}

	dstVal = dstVal.Elem()

	logDebug("UnmarshalByReflect param:", srcVal.Kind().String(), dstVal.Kind().String())
	// 如果是切片
	if srcVal.Kind() == reflect.Slice && dstVal.Kind() == reflect.Slice {
		return assignConvertSlice(srcVal, dstVal)
	}

	//对 srcStruct 和 dstPoint 进行处理
	fill := new(getNewService)
	dstValue, err := fill.GetByDstAll(srcStruct, reflect.TypeOf(dstPoint))
	if err != nil {
		return err
	}
	if !dstValue.IsValid() {
		err = copier.CopyWithOption(dstPoint, srcStruct, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		if err == nil {
			return nil
		}
		return fmt.Errorf("UnmarshalByReflect error: %s, %s", reflect.TypeOf(dstPoint).String(), err.Error())
	}

	logDebug("fill.GetByDstAll:", String(dstValue.Interface()))

	t := new(toolsService)
	dstStruct, _ := t.getNewSrcAndDst(dstValue.Interface(), dstPoint)

	errJson := t.UnmarshalDataFromJson(dstStruct, dstPoint)
	if errJson == nil {
		return nil
	}

	errTemp := t.assignTo(dstValue, dstPoint)
	if errTemp != nil {
		err = copier.CopyWithOption(dstPoint, dstStruct, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		if err == nil {
			return nil
		}
		return errTemp
	}

	return nil
}

func assignConvertSlice(src, dst reflect.Value) error {
	if dst.Kind() != reflect.Slice {
		return fmt.Errorf("dst must be a slice")
	}

	dst.Set(reflect.MakeSlice(dst.Type(), src.Len(), src.Cap()))
	elemType := dst.Type().Elem()
	for i := 0; i < src.Len(); i++ {
		var newData reflect.Value
		if elemType.Kind() == reflect.Ptr {
			newData = reflect.New(elemType.Elem())
		} else {
			newData = reflect.New(elemType)
		}

		if err := toAssignTo(src.Index(i).Interface(), newData.Interface()); err != nil {
			return err
		}

		if dst.Index(i).CanSet() {
			if elemType.Kind() == reflect.Ptr {
				dst.Index(i).Set(newData)
			} else {
				dst.Index(i).Set(newData.Elem())
			}
		}
	}
	return nil
}

func NewPtrByType(dstType reflect.Type) any {
	if dstType.Kind() == reflect.Slice {
		newType := reflect.New(dstType)
		dstSliceValue := reflect.MakeSlice(dstType, 0, 0)
		newType.Elem().Set(dstSliceValue)
		return newType.Interface()
	} else if dstType.Kind() == reflect.Ptr {
		ts := new(toolsService)
		dstDirectType := ts.getDirectTypeByPtr(dstType)
		if dstDirectType.Kind() != reflect.Ptr {
			return NewPtrByType(dstDirectType)
		}
		return reflect.New(dstDirectType).Interface()
	} else if dstType.Kind() == reflect.Struct {
		if dstType == reflect.TypeOf(time.Time{}) {
			return &time.Time{}
		}
		return reflect.New(dstType).Interface()
	} else if dstType.Kind() == reflect.Map {
		newType := reflect.New(dstType)
		keyType := dstType.Key()
		valueType := dstType.Elem()
		datMapValue := reflect.MakeMap(reflect.MapOf(keyType, valueType))
		newType.Elem().Set(datMapValue)
		return newType.Interface()
	} else if dstType.Kind() == reflect.Int64 ||
		dstType.Kind() == reflect.Int || dstType.Kind() == reflect.String { //常用的类型
		ptrVal := reflect.New(dstType)
		return ptrVal.Interface()
	} else if dstType.Kind() == reflect.Bool {
		oneBool := false
		return &oneBool
	} else if dstType.Kind() == reflect.Interface {
		if dstType == reflect.TypeOf((*error)(nil)).Elem() {
			return fmt.Errorf("")
		}
	}

	return reflect.New(dstType).Interface()
}

type getNewService struct {
}

// GetByDstAll 根据Dst的类型，获取srcInterface的值
func (c *getNewService) GetByDstAll(srcInterface any, dstType reflect.Type) (newDstValue reflect.Value, err error) {
	logDebug("GetByDstAll param:", String(srcInterface), dstType.String(), dstType.Kind().String())

	srcType := reflect.TypeOf(srcInterface)
	if srcType == dstType {
		//直接返回
		return reflect.ValueOf(srcInterface), nil
	}

	var newDstList reflect.Value
	var found bool

	if dstType.Kind() == reflect.Slice {
		found = true
		newDstList, err = c.getByDstSlice(srcInterface, dstType)
	} else if dstType.Kind() == reflect.Ptr {
		found = true
		newDstList, err = c.getByDstPtr(srcInterface, dstType)
	} else if dstType.Kind() == reflect.Struct {
		logDebug("GetByDstAll struct:", String(srcInterface), dstType.String())
		dstIns := reflect.New(dstType)
		if srcType.Kind() == reflect.Map {
			found = true
			t := new(toolsService)
			err = t.UnmarshalDataFromJson(srcInterface, dstIns.Interface())
			if err == nil {
				logDebug("GetByDstAll mapToStruct result:", String(dstIns.Interface()))
				newDstList = dstIns.Elem()
			}
		}

		if !found || err != nil {
			found = true
			newDstList, err = c.getByDstStruct(srcInterface, dstType)
			logDebug("GetByDstAll struct result:", newDstList, dstType.String())
		}
	} else if dstType.Kind() == reflect.Map {
		found = true
		newDstList, err = c.getByDstMap(srcInterface, dstType)
	}

	// 完成
	if found {
		if err == nil && newDstList.IsValid() {
			logDebug("GetByDstAll found:", String(newDstList.Interface()))
			return newDstList, nil
		}
		newDstList2, err2 := c.getByDstOther(srcInterface, dstType)
		if err2 == nil && newDstList2.IsValid() {
			return newDstList2, nil
		}
		return newDstList, err
	}

	//未找到的情况用默认的方法
	newDstList2, err2 := c.getByDstOther(srcInterface, dstType)
	if err2 == nil && newDstList2.IsValid() {
		return newDstList2, nil
	}

	return newDstList2, err2
}

// getByDstSlice 根据DstSlice获取列表
func (c *getNewService) getByDstSlice(srcSlice any, dstType reflect.Type) (newDstList reflect.Value, err error) {
	if dstType.Kind() != reflect.Slice {
		//nolint:goerr113
		return reflect.Value{}, fmt.Errorf(errStrNotSlice, dstType.String())
	}

	toPointList := make([]any, 0)

	t := new(toolsService)
	err2 := t.UnmarshalDataFromJson(srcSlice, &toPointList)
	if err2 != nil {
		return reflect.Value{}, err2
	}

	dstSliceValue := reflect.MakeSlice(dstType, 0, 0)
	elemType := dstSliceValue.Type().Elem()

	for m := 0; m < len(toPointList); m++ {
		oneElem := toPointList[m]
		newDataValue, errTemp := c.GetByDstAll(oneElem, elemType)
		if errTemp == nil && newDataValue.IsValid() {
			dstSliceValue = t.appendSliceValue(dstSliceValue, newDataValue)
		}
		if errTemp != nil {
			err = errTemp
		}
	}

	return dstSliceValue, err
}

// getByDstPtr 根据Ptr获得一个指针对象
func (c *getNewService) getByDstPtr(srcInterface any, dstType reflect.Type) (newDstPtr reflect.Value, err error) {
	if dstType.Kind() != reflect.Ptr {
		//nolint:goerr113
		return reflect.Value{}, fmt.Errorf(errStrGetByDstPtr, dstType.String())
	}

	t := new(toolsService)

	dstDataType := t.getDirectTypeByPtr(dstType)

	logDebug("getByDstPtr:", dstDataType.String())

	if dstType == reflect.TypeOf(&timestamppb.Timestamp{}) {
		if oneTime, err1 := toConvert[time.Time](srcInterface); err1 == nil {
			dstData := timestamppb.New(oneTime)
			return reflect.ValueOf(dstData), nil
		}
	}

	logDebug("getByDstPtr2:", dstDataType.String())

	dstDataInterface, err := c.GetByDstAll(srcInterface, dstDataType)
	if err != nil || !dstDataInterface.IsValid() {
		return reflect.Value{}, err
	}

	logDebug("getByDstPtr data:", dstDataInterface.Interface(), dstDataInterface.Type())

	newRetVal := t.getDirectValueByPtr(dstType, dstDataInterface)
	if newRetVal.IsValid() {
		logDebug("getByDstPtr:", newRetVal.Interface())
		logDebug("getByDstPtr:", newRetVal.Type().String())
		return newRetVal, nil
	}
	return reflect.Value{}, nil
}

// getByDstStruct 根据Struct获得一个
func (c *getNewService) getByDstStruct(srcStruct any, dstType reflect.Type) (newDstStruct reflect.Value, err error) {
	if dstType.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf(errStrGetByDstMapNotStruct, dstType.String())
	}

	//屏蔽意外的错误
	defer func() {
		errTemp := recover()
		if !cond.IsNil(errTemp) {
			log.Println("Unmarshal error:", errTemp)
			err = fmt.Errorf(errStrRecover2, errTemp)
		}
	}()

	isSetStruct := false

	t := new(toolsService)

	//查找每一个字段
	dstIns := reflect.New(dstType)

	dstStructValue := dstIns.Elem()
	columnNum := dstType.NumField()
	for i := 0; i < columnNum; i++ {
		dstColumnField := dstType.Field(i)
		dstColumnValue := dstStructValue.Field(i)

		//继承
		logDebug("getByDstStruct Type:", dstColumnField.Name, dstColumnField.Type.Name())
		if dstColumnField.Name == dstColumnField.Type.Name() {
			newDataValue, errTemp := c.GetByDstAll(srcStruct, dstColumnField.Type)

			logDebug("getByDstStruct:", dstColumnField.Name, dstColumnField.Type.String(), newDataValue.Interface())

			if errTemp == nil && newDataValue.IsValid() {
				logDebug("getByDstStruct 312:", dstColumnValue.Type().String(), newDataValue.Interface())
				if dstColumnValue.CanSet() {
					dstColumnValue.Set(newDataValue)
					isSetStruct = true
				} else {
					//如果是继承，则需要递归设置
					//if dstColumnValue.Type() == newDataValue.Type() {
					//	newDst := reflect.New(newDataValue.Type())
					//	newDst.Elem().Set(newDataValue)
					//	dstColumnValue.Set(newDst.Elem())
					//}
				}
			}
			continue
		}

		// 当前字段是否能设置，放后面解决类为小写的情况
		logDebug("canSetStructColumn:", dstColumnField.Name, dstColumnValue.String())
		if canSet := t.canSetStructColumn(dstColumnField.Name, dstColumnValue); !canSet {
			continue
		}

		//从src获取每一个目标的值,src 是一个整体，需要一一读取
		valueTemp := c.GetSrcFromStructField(srcStruct, dstColumnField)

		logDebug("getByDstStruct 344:", String(srcStruct), dstColumnField, valueTemp)

		if cond.IsNil(valueTemp) {
			//源数据为nil，则不用设置
			continue
		}

		newDataValue, errTemp := c.GetByDstAll(valueTemp, dstColumnField.Type)
		if errTemp == nil && newDataValue.IsValid() {
			logDebug("getByDstStruct 346:", dstColumnField.Name, dstColumnValue.Type().String(), newDataValue.Interface())
			dstColumnValue.Set(newDataValue)
			isSetStruct = true
		}

		if errTemp != nil { //如果有错误，需要处理
			err = errTemp
		}
	}

	//正常设置
	if isSetStruct {
		logDebug("getByDstStruct: ", dstStructValue.Interface(), dstStructValue.Type().String())
		return dstStructValue, err
	}

	logDebug("getByDstStruct error:", String(srcStruct), err, dstType.String())

	return reflect.Value{}, err
}

// getByDstMap 根据map获得一个指针对象
func (c *getNewService) getByDstMap(srcStruct any, dstType reflect.Type) (newDstStruct reflect.Value, err error) {
	if dstType.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf(errStrGetByDstMapNotMap, dstType.String())
	}
	keyType := dstType.Key()
	if keyType.Kind() != reflect.String {
		return reflect.Value{}, fmt.Errorf(errStrGetByDstMap, keyType.String())
	}

	logDebug("getByDstMap param:", String(srcStruct), dstType.String())

	toMap := make(map[string]any)

	t := new(toolsService)
	err2 := t.UnmarshalDataFromJson(srcStruct, &toMap)
	if err2 != nil {
		logDebug("getByDstMap err:", err2.Error())
		return reflect.Value{}, err2
	}

	valueType := dstType.Elem()
	datMapValue := reflect.MakeMap(reflect.MapOf(keyType, valueType))

	isSetMap := false
	for key, val := range toMap {
		tempKey, err1 := c.GetByDstAll(key, keyType)
		tempVal, err2 := c.GetByDstAll(val, valueType)
		if err1 == nil &&
			err2 == nil &&
			tempKey.IsValid() &&
			tempVal.IsValid() {
			datMapValue.SetMapIndex(tempKey, tempVal)
			isSetMap = true
		}
	}
	if isSetMap {
		return datMapValue, nil
	}

	logDebug("getByDstMap err2:", err)
	return reflect.Value{}, err
}

// getByDstOther 根据map获得一个指针对象
func (c *getNewService) getByDstOther(srcOther any, dstType reflect.Type) (newDstOther reflect.Value, err error) {
	newPtr := reflect.New(dstType)
	srcValue := reflect.ValueOf(srcOther)

	defer func() {
		errTemp := recover()
		if !cond.IsNil(errTemp) {
			err = fmt.Errorf(errStrGetByDstOther, errTemp)
		}
	}()

	hasSet := false
	if newPtr.Elem().Type() == srcValue.Type() {
		newPtr.Elem().Set(srcValue)
		hasSet = true
	} else {
		if srcValue.Kind() == reflect.Interface || srcValue.Kind() == reflect.Ptr {
			if newPtr.Elem().Type() == srcValue.Elem().Type() {
				newPtr.Elem().Set(srcValue.Elem())
				hasSet = true
			}
		}
	}
	if !hasSet {
		if srcValue.CanConvert(dstType) {
			newV := srcValue.Convert(dstType)
			logDebug("newV", String(String(newV.Interface())), String(srcValue.Interface()))
			//这里有可能将55数字转换为7字符的问题，不是想要的效果，所以需要对值进行判断
			if newPtr.Elem().Type() == newV.Type() && String(newV.Interface()) == String(srcValue.Interface()) {
				newPtr.Elem().Set(newV)
				hasSet = true
			}
		}
		if !hasSet {
			//自定义转换
			newDst, err := c.getByDstDefault(srcOther, dstType)
			if err == nil {
				logDebug("getByDstOther type:", newPtr.Elem().Type().String(), newDst.Type().String())

				if newPtr.Elem().Type() == newDst.Type() {
					newPtr.Elem().Set(newDst)
					hasSet = true
				} else if newPtr.Elem().Kind() == reflect.Interface {
					// 目标为any，则可以设置任意值
					newPtr.Elem().Set(newDst)
					hasSet = true
				}
			}
		}
	}
	if hasSet {
		return newPtr.Elem(), err
	}
	if err == nil {
		err = fmt.Errorf(errStrGetByDstOther, "no set")
	}
	return reflect.Value{}, err
}

// getByDstDefault 自定义的格式转换
func (c *getNewService) getByDstDefault(srcDefault any, dstType reflect.Type) (newDstOther reflect.Value, err error) {
	//fmt.Println("getByDstDefault:", srcDefault, dstType.String())

	srcValue := reflect.ValueOf(srcDefault)
	if !srcValue.IsValid() {
		return reflect.Value{}, fmt.Errorf("src inValid")
	}
	if dstType == srcValue.Type() {
		return srcValue, nil
	}

	if dstType.Kind() == reflect.Bool {
		retBool, err1 := toConvert[bool](srcDefault)
		if err1 == nil {
			return reflect.ValueOf(retBool), nil
		}
	}

	if retData, ok := c.changeValueToDstByDstType(srcValue, dstType); ok {
		if retData != nil {
			return reflect.ValueOf(retData), nil
		}
		return reflect.Value{}, fmt.Errorf("changeValueToDstByDstType error")
	}

	if retData, ok := c.changeValueToDstBySrcType(srcValue, dstType); ok {
		if retData != nil {
			return reflect.ValueOf(retData), nil
		}
		return reflect.Value{}, fmt.Errorf("changeValueToDstBySrcType error")
	}

	return reflect.Value{}, fmt.Errorf("no change")
}

func (c *getNewService) changeValueToDstByDstType(srcValue reflect.Value, dstType reflect.Type) (any, bool) {
	dstTypeName := dstType.Name()
	dstTypeString := dstType.String()

	//fmt.Println("changeValueToDstByDstType:", dstTypeName, dstTypeString, srcValue.Type().String())

	if dstType.Kind() == reflect.String {
		tempTime, ok := c.changeValueToString(srcValue)
		if ok {
			return tempTime, ok
		}
		return nil, true
	}

	if dstType.Kind() == reflect.Int8 {
		tempTime, ok := c.changeValueToInt8(srcValue)
		if ok {
			return tempTime, ok
		}
		return nil, true
	}

	if dstType.Kind() == reflect.Uint64 {
		tempTime, ok := c.changeValueToInt64(srcValue)
		if ok {
			return uint64(tempTime), ok
		}
		return nil, true
	}

	if dstType.Kind() == reflect.Int64 {
		tempTime, ok := c.changeValueToInt64(srcValue)
		if ok {
			return tempTime, ok
		}
		return nil, true
	}

	srcInterface := srcValue.Interface()
	switch dstType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if srcInterface == nil {
			return nil, true
		}
		s := asString(srcInterface)
		i64, err := strconv.ParseInt(s, 10, dstType.Bits())
		if err != nil {
			return nil, false
		}
		return reflect.ValueOf(i64).Convert(dstType).Interface(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if srcInterface == nil {
			return nil, true
		}
		s := asString(srcInterface)
		u64, err := strconv.ParseUint(s, 10, dstType.Bits())
		if err != nil {
			return nil, false
		}
		return reflect.ValueOf(u64).Convert(dstType).Interface(), true
	case reflect.Float32, reflect.Float64:
		if srcInterface == nil {
			return nil, true
		}
		s := asString(srcInterface)
		f64, err := strconv.ParseFloat(s, dstType.Bits())
		if err != nil {
			return nil, false
		}
		return reflect.ValueOf(f64).Convert(dstType).Interface(), true
	default:

	}

	if dstType.Kind() == reflect.Bool {
		tempTime, ok := c.changeValueToBool(srcValue)
		if ok {
			return tempTime, ok
		}
		return nil, true
	}

	if dstType.Kind() == reflect.Interface {
		if dstTypeName == "error" {
			tempTime, ok := c.changeValueToError(srcValue)
			if ok {
				return tempTime, ok
			}
			return nil, true
		}
		//如果为any类型，则看是否可以直接进行赋值
		return srcValue.Interface(), true
	}
	if dstType.Kind() == reflect.Struct {
		if dstType == reflect.TypeOf(time.Time{}) {
			tempTime, ok := c.changeValueToTime(srcValue)
			if ok {
				return tempTime, ok
			}
			return nil, true
		}
		if dstType == reflect.TypeOf(decimal.Decimal{}) {
			sStr := fmt.Sprintf("%v", srcValue.Interface())
			//如果数字里含有","，西方数字分段字符，则会报错，所以需要替换为空
			sStr = strings.ReplaceAll(sStr, ",", "")
			decimalData, err := decimal.NewFromString(sStr)
			if err == nil {
				return decimalData, true
			}
			return nil, true
		}
	}
	if dstType.Kind() == reflect.Slice {
		if dstTypeString == "[]string" {
			tempTime, ok := c.changeValueStringToStringList(srcValue)
			if ok {
				return tempTime, ok
			}
			return nil, true
		}
	}

	//log.Println("changeValueToDstByDstType error:", dstType.String())

	return nil, false
}

func asString(src any) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	default:
		return fmt.Sprintf("%v", src)
	}
}

func (c *getNewService) changeValueToString(srcValue reflect.Value) (string, bool) {
	srcTypeName := srcValue.Type().Name()
	sStr := ""
	if srcTypeName == "int64" {
		sStr = strconv.FormatInt(srcValue.Int(), 10)
		return sStr, true
	} else if srcTypeName == "int" {
		int64Num := srcValue.Int()
		intNum := *(*int)(unsafe.Pointer(&int64Num))
		sStr = strconv.Itoa(intNum)
		return sStr, true
	} else if srcTypeName == "Time" {
		temp := srcValue.Interface().(time.Time)
		return String(temp), true
	} else if strings.Contains(srcTypeName, "byte") {
		temp := srcValue.Interface().(string)
		sStr = String(temp)
		return sStr, true
	} else if srcValue.Interface() != nil {
		sStr = String(srcValue.Interface())
		return sStr, true
	}
	return sStr, false
}
func (c *getNewService) changeValueToTime(srcValue reflect.Value) (time.Time, bool) {
	tempTime, err1 := toConvert[time.Time](srcValue.Interface())
	if err1 == nil {
		return tempTime, true
	}
	return time.Time{}, false
}

func (c *getNewService) changeValueToInt8(srcValue reflect.Value) (int8, bool) {
	srcInterface := srcValue.Interface()
	sStr := fmt.Sprintf("%v", srcInterface)
	sStrInt, err := strconv.ParseInt(sStr, 10, 8)
	if err == nil {
		return int8(sStrInt), true
	}
	return 0, false
}
func (c *getNewService) changeValueToInt64(srcValue reflect.Value) (int64, bool) {
	srcValueTypeName := srcValue.Type().Name()
	if srcValueTypeName == "int" {
		return srcValue.Int(), true
	}

	srcInterface := srcValue.Interface()
	intTemp, _ := Int64(srcInterface)
	if intTemp != 0 {
		return intTemp, true
	}

	sStr := fmt.Sprintf("%v", srcValue)
	sStrInt, err := strconv.ParseInt(sStr, 10, 64)
	if err == nil {
		return sStrInt, true
	}
	return 0, false
}

func (c *getNewService) changeValueToBool(srcValue reflect.Value) (bool, bool) {
	srcTypeName := reflect.TypeOf(srcValue.Interface()).String()
	if srcTypeName == "string" {
		srcColumnValueString := srcValue.String()
		newColumnString := strings.ToLower(srcColumnValueString)
		if newColumnString == "true" {
			return true, true
		} else if newColumnString == "false" {
			return false, true
		} else {
			sInt, err := strconv.Atoi(srcColumnValueString)
			if err == nil {
				if sInt == 1 {
					return true, true
				} else if sInt == 0 {
					return false, true
				}
			}
		}
	}

	if srcTypeName == "int" ||
		srcTypeName == "float64" ||
		srcTypeName == "int64" {
		srcInterface := srcValue.Interface()
		boolRet, _ := Int64(srcInterface)
		if boolRet == 1 {
			return true, true
		} else if boolRet == 0 {
			return false, true
		}
	}
	return false, false
}

func (c *getNewService) changeValueToError(srcValue reflect.Value) (error, bool) {
	srcIns := srcValue.Interface()
	if srcInsErr, ok := srcIns.(error); ok {
		return srcInsErr, true
	}
	return nil, false
}

func (c *getNewService) changeValueStringToStringList(srcValue reflect.Value) ([]string, bool) {
	srcTypeName := srcValue.Type().Name()
	if srcTypeName == "string" {
		srcColumnValueString := srcValue.String()
		arrList := make([]string, 0)
		err := jsoniterForNil.UnmarshalFromString(srcColumnValueString, &arrList)
		if err == nil {
			return arrList, true
		} else {
			t := new(toolsService)
			arrList = t.split(srcColumnValueString, []string{"|", ";", ","})
			return arrList, true
		}
	}

	return []string{}, false
}

func (c *getNewService) changeFromString(srcValue reflect.Value, dstTypeName string) (any, bool) {
	srcColumnValueString := srcValue.String()
	if dstTypeName == "float32" {
		sFloat, err := strconv.ParseFloat(srcColumnValueString, 32)
		if err == nil {
			return float32(sFloat), true
		}
		return nil, true
	}
	if dstTypeName == "float64" {
		sFloat, err := strconv.ParseFloat(srcColumnValueString, 64)
		if err == nil {
			return sFloat, true
		}
		return nil, true
	}
	if dstTypeName == "int" {
		sInt, err := strconv.Atoi(srcColumnValueString)
		if err == nil {
			return sInt, true
		}
		return nil, true
	}
	if dstTypeName == "Time" {
		sTime, err1 := toConvert[time.Time](srcColumnValueString)
		if err1 == nil {
			return sTime, true
		}
		return nil, true
	}

	return nil, false
}
func (c *getNewService) changeFromBool(srcValue reflect.Value, dstType reflect.Type) (any, bool) {
	srcBool := srcValue.Bool()
	switch dstType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		disIns := reflect.New(dstType)
		intNum := 0
		if srcBool {
			intNum = 1
		}
		disIns.Elem().Set(reflect.ValueOf(intNum))
		return disIns.Elem().Interface(), true
	case reflect.String:
		if srcBool {
			return "true", true
		}
		return "false", true
	default:
		return nil, false
	}
}

func (c *getNewService) changeFromByte(srcValue reflect.Value, dstTypeName string) (any, bool) {
	srcInterface := srcValue.Interface()
	if byteTemp, ok := srcInterface.([]byte); ok {
		srcColumnValueString := string(byteTemp)
		if dstTypeName == "string" {
			return srcColumnValueString, true
		} else if dstTypeName == "int64" {
			sStr, _ := Int64(srcColumnValueString)
			if sStr != 0 {
				return sStr, true
			}
		} else if dstTypeName == "int" {
			sInt, err := strconv.Atoi(srcColumnValueString)
			if err == nil {
				return sInt, true
			}
		} else if dstTypeName == "float64" {
			sFloat, err := strconv.ParseFloat(srcColumnValueString, 64)
			if err == nil {
				return sFloat, true
			}
		} else if dstTypeName == "Time" {
			STime, err1 := toConvert[time.Time](srcColumnValueString)
			if err1 == nil {
				return STime, true
			}
		}
	}
	return nil, false
}
func (c *getNewService) changeFromUint8(srcValue reflect.Value, dstTypeName string) (any, bool) {
	srcInterface := srcValue.Interface()
	srcString := String(srcInterface)
	if dstTypeName == "int" {
		one, _ := Int64(srcString)
		return int(one), true
	}
	if dstTypeName == "int64" {
		return Int64(srcString)
	}
	if dstTypeName == "Time" {
		timeTemp, err1 := toConvert[time.Time](srcString)
		if err1 == nil {
			return timeTemp, true
		}
	}

	return srcString, true
}
func (c *getNewService) changeFromFloat64(srcValue reflect.Value, dstTypeName string) (any, bool) {
	float64Num := srcValue.Float()
	if dstTypeName == "int" {
		intNum := int(float64Num)
		return intNum, true
	}
	if dstTypeName == "int64" {
		intNum := int64(float64Num)
		return intNum, true
	}
	if dstTypeName == "string" {
		intStr := strconv.FormatFloat(float64Num, 'E', -1, 64)
		return intStr, true
	}
	return nil, false
}

func (c *getNewService) changeValueToDstBySrcType(srcValue reflect.Value, dstType reflect.Type) (any, bool) {
	dstTypeName := dstType.Name()
	srcTypeString := srcValue.Type().String()

	srcTypeKind := srcValue.Type().Kind()

	//fmt.Println("changeValueToDstBySrcType:")
	//fmt.Println(srcValue.Interface())
	//fmt.Println(srcValue.String())
	//fmt.Println(dstType.String())

	if srcTypeKind == reflect.String {
		if retData, found := c.changeFromString(srcValue, dstTypeName); found {
			return retData, found
		}
	}
	if srcTypeKind == reflect.Bool {

		if retData, found := c.changeFromBool(srcValue, dstType); found {
			return retData, found
		}
	}

	if srcTypeKind == reflect.Int64 {
		if dstType.Kind() == reflect.Int {
			int64Num := srcValue.Int()
			intNum := *(*int)(unsafe.Pointer(&int64Num))
			return intNum, true
		}
	}

	if srcTypeKind == reflect.Float64 {
		if retData, found := c.changeFromFloat64(srcValue, dstTypeName); found {
			return retData, found
		}
	}

	if srcTypeKind == reflect.Slice {
		sonType := srcValue.Type().Elem()
		if sonType.Kind() == reflect.Uint8 {

			if srcTypeString == "[]byte" { //不是普通类型，是复合类型，看看能否再次调用了
				if retData, found := c.changeFromByte(srcValue, dstTypeName); found {
					return retData, found
				}
			}

			if srcTypeString == "[]uint8" {
				if retData, found := c.changeFromUint8(srcValue, dstTypeName); found {
					return retData, found
				}
			}
		}
	}

	if srcTypeString == "map[string]interface {}" {
		srcInterface := srcValue.Interface()
		if _, ok := srcInterface.(map[string]any); ok {
			//for kkk, vvv := range mapTemp {
			//
			//}
			//fmt.Println(mapTemp)
		}

	}

	return nil, false
}

// GetSrcFromStructField 取下一级的数据
func (c *getNewService) GetSrcFromStructField(srcInterface any, dstColumn reflect.StructField) any {

	logDebug("GetSrcFromStructField", String(srcInterface), dstColumn.Name)

	srcVal := reflect.ValueOf(srcInterface)
	for srcVal.Kind() == reflect.Ptr || srcVal.Kind() == reflect.Interface {
		srcVal = srcVal.Elem()
	}

	srcType := srcVal.Type()

	//1、如果是struct，则首先从struct中进行匹配，名称完全一样的进行匹配
	if srcType.Kind() == reflect.Struct {
		return c.getColumnValueFromStruct(srcInterface, dstColumn)
	}

	if srcType.Kind() == reflect.Map {
		return c.getColumnValueFromMap(srcInterface, dstColumn)
	}

	return c.getColumnValueFromType(srcInterface, dstColumn)
}

func (c *getNewService) getColumnValueFromStruct(srcStruct any, dstColumn reflect.StructField) any {
	logDebug("getColumnValueFromStruct:", String(srcStruct), dstColumn.Name)

	//1、如果是struct，则首先从struct中进行匹配，名称完全一样的进行匹配
	srcType := reflect.TypeOf(srcStruct)
	srcValue := reflect.ValueOf(srcStruct)

	var srcColumnField reflect.StructField
	var srcColumnValue reflect.Value

	findFromSrc := false //名字一模一样，

	t := new(toolsService)
	allSrcTypeList, allSrcValueList := t.getAllStructColumn(srcType, srcValue)

	for j := 0; j < len(allSrcTypeList); j++ {
		s := allSrcTypeList[j]
		logDebug("getColumnValueFromStruct:", dstColumn.Name, s.Name)
		if s.Name == dstColumn.Name {
			srcColumnField = s
			srcColumnValue = allSrcValueList[j]
			findFromSrc = true
			break
		}
	}

	//没有找到
	if !findFromSrc {
		return nil
	}
	if !srcColumnValue.IsValid() {
		return nil
	}

	if srcColumnField.Type == dstColumn.Type {
		return srcColumnValue.Interface()
	}

	retVal, err := c.GetByDstAll(srcColumnValue.Interface(), dstColumn.Type)
	if err != nil || !retVal.IsValid() {
		return nil
	}
	if !retVal.IsValid() {
		return nil
	}

	return retVal.Interface()
}

func (c *getNewService) getColumnValueFromMap(srcMap any, dstColumn reflect.StructField) any {
	//1、如果是map，则先用原来的名字，再用json的名字
	srcValue := reflect.ValueOf(srcMap)

	t := new(toolsService)
	dstColumnJsonNameList := t.getAllMapNameByField(dstColumn)

	var srcColumnKey reflect.Value
	findFromSrc := false
	for _, oneName := range dstColumnJsonNameList {
		for _, k := range srcValue.MapKeys() {
			if k.String() == oneName {
				srcColumnKey = k
				findFromSrc = true
				break
			}
		}
	}

	if !findFromSrc {
		return nil
	}

	srcColumnValue := srcValue.MapIndex(srcColumnKey)

	if !srcColumnValue.IsValid() {
		return nil
	}

	if srcColumnValue.Type() == dstColumn.Type {
		return srcColumnValue.Interface()
	}

	retVal, err := c.GetByDstAll(srcColumnValue.Interface(), dstColumn.Type)
	if err != nil || !retVal.IsValid() {
		return nil
	}
	if !retVal.IsValid() {
		return nil
	}

	//fmt.Println("getColumnValueFromMap return:", retVal.Interface())

	return retVal.Interface()
}

// getColumnValueFromType 从里面拿一个值，而不是取本身
func (c *getNewService) getColumnValueFromType(srcInterface any, dstColumn reflect.StructField) any {
	t := new(toolsService)
	toMap := make(map[string]any)
	err2 := t.UnmarshalDataFromJson(srcInterface, &toMap)
	if err2 != nil {
		return nil
	}

	if len(toMap) == 0 {
		return nil
	}

	return c.getColumnValueFromMap(toMap, dstColumn)
}

// newPtrType 接收任意类型（包括指针），若为指针则创建同类型新指针
func newPtrType(valType reflect.Type) any {
	if valType.Kind() != reflect.Ptr {
		valType = reflect.PointerTo(valType)
	}
	// 获取指针指向的元素类型（如 *int -> int）
	elemType := valType.Elem()
	newPtr := reflect.New(elemType)
	return newPtr.Interface()
}
