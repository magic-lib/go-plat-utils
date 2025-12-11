package conv

import (
	"fmt"
	"github.com/jinzhu/copier"
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
	srcType := reflect.TypeOf(srcStruct)
	srcVal := reflect.ValueOf(srcStruct)
	if srcType.Kind() == reflect.Ptr {
		if srcVal.IsNil() {
			return nil
		}
	}

	isString := false
	oldString := ""
	ok := false
	if oldString, ok = srcStruct.(string); ok {
		isString = true
	}
	if isString {
		//字符串为空，则直接返回
		if oldString == "" {
			return nil
		}
	}

	if dstPoint == nil {
		return fmt.Errorf("unmarshal DstPoint is nil")
	}

	// 1、首先看能否直接赋值
	dstType := reflect.TypeOf(dstPoint)
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

	// 2、不行则用json方法
	t := new(toolsService)
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
		log.Println("Unmarshal error:", err)
		return errJson
	}
	return nil
}

var OpenLog = false

func logDebug(str ...any) {
	if !OpenLog {
		return
	}
	strArr := make([]any, 0)
	strArr = append(strArr, "[logDebug]")
	strArr = append(strArr, str...)
	fmt.Println(strArr...)
}
