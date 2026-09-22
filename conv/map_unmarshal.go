package conv

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/magic-lib/go-plat-utils/cond"
	"log"
	"reflect"
)

var Marshal = String

// Unmarshal 将前一个的对象填充到后一个对象中，字段名相同的覆盖值，
// 返回 interface 的作用是如果toPoint为nil的时候，也能正常返回对象.
func Unmarshal(srcStruct any, dstPoint any) error {
	if srcStruct == nil {
		return nil
	}
	oldStruct := srcStruct
	oldString, isString := checkIsString(srcStruct)
	if isString {
		if oldString == "" {
			return nil
		}
		srcStruct = oldString
	}

	srcType := reflect.TypeOf(srcStruct)
	srcVal := reflect.ValueOf(srcStruct)
	if srcType.Kind() == reflect.Ptr {
		if srcVal.IsNil() {
			return nil
		}
	}

	if dstPoint == nil {
		return fmt.Errorf("unmarshal DstPoint is nil")
	}

	dstType := reflect.TypeOf(dstPoint)
	if (srcType.Kind() == reflect.Ptr || srcType.Kind() == reflect.Struct) &&
		dstType.Kind() == reflect.Ptr {
		if dstType.Elem().Kind() == reflect.Map {
			return Unmarshal(String(srcStruct), dstPoint)
		}
	}

	// 解决 map[string]any 相互转换会存在any类型丢失后，类型不同的问题
	if srcType.Kind() == reflect.Map {
		if srcType.String() == dstType.Elem().String() {
			if newMap, ok := cloneSrcAnyMap(srcStruct); ok {
				if dstPtr, ok2 := dstPoint.(*map[string]any); ok2 {
					if newMap != nil {
						// 直接把克隆出来的 map 赋给 *dstPoint 指向的变量
						*dstPtr = newMap
					}
					return nil
				}
			}
		}
	}

	// 1、首先看能否直接赋值
	if srcType == dstType {
		if srcType.Kind() != reflect.Ptr &&
			srcType.Kind() != reflect.Struct &&
			srcType.Kind() != reflect.Map {

			err := copier.CopyWithOption(dstPoint, srcStruct, copier.Option{IgnoreEmpty: true, DeepCopy: true})
			if err == nil {
				return nil
			}

			t := new(toolsService)
			err = t.AssignTo(reflect.ValueOf(srcStruct), dstPoint)
			if err == nil {
				return nil
			}

		}
	}

	t := new(toolsService)
	// 2、不行则用json方法
	if cond.IsBytes(oldStruct) {
		errJson := t.UnmarshalDataFromJson(oldStruct, dstPoint)
		if errJson == nil {
			return nil
		}
	}

	srcStruct, dstPoint = t.getNewSrcAndDst(srcStruct, dstPoint)

	//先用对象进行替换，因为转换为json串以后，会丢失类型
	err := toAssignTo(srcStruct, dstPoint)
	if err == nil {
		return nil
	}

	//2.2 Unmarshal 会丢失类型
	errJson := t.UnmarshalDataFromJson(srcStruct, dstPoint)
	if errJson == nil {
		return nil
	}

	// 3、用转换一一覆盖
	//表示有格式不能兼容，出现错误，所以需要进行特殊处理
	srcType = reflect.TypeOf(srcStruct)
	if srcType.Kind() != reflect.Struct &&
		srcType.Kind() != reflect.Map &&
		srcType.Kind() != reflect.Slice &&
		srcType.Kind() != reflect.Array {
		//普通类型能否相互转换，由 string 转换为 *int64
		err = toAssignTo(srcStruct, dstPoint)
		if err == nil {
			return nil
		}

		//如果是字符串，则需要保证是json格式的
		if isString {
			//nolint:goerr113
			return fmt.Errorf(errStrUnmarshal1, oldString)
		}
		return errJson
	}

	err = toAssignTo(srcStruct, dstPoint)
	if err != nil {
		log.Println("Unmarshal toAssignTo error:", err, "src:", String(srcStruct))
		return errJson
	}
	return nil
}

func checkIsString(srcStruct any) (string, bool) {
	isString := false
	oldString := ""

	// string 直接取值；底层是 []byte 的（含 json.RawMessage 这类具名类型）统一转成 string
	if str, ok := srcStruct.(string); ok {
		oldString, isString = str, true
	} else {
		srcType := reflect.TypeOf(srcStruct)
		if srcType.Kind() == reflect.Slice && srcType.Elem().Kind() == reflect.Uint8 {
			srcVal := reflect.ValueOf(srcStruct)
			oldString, isString = string(srcVal.Bytes()), true
		}
	}
	return oldString, isString
}

func logDebug(str ...any) {
	if !cond.IsOpenLog() {
		return
	}
	strArr := make([]any, 0)
	strArr = append(strArr, "[logDebug]")
	strArr = append(strArr, str...)
	fmt.Println(strArr...)
}
