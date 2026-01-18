package utils

import (
	"fmt"
	"reflect"
)

// GetGenericTypeId 获取泛型类型的唯一标识符
func GetGenericTypeId[T any]() (string, error) {
	typ := reflect.TypeFor[T]()
	if typ == nil {
		return "", fmt.Errorf("GetGenericTypeId: type is not a generic type")
	}
	if typ.PkgPath() == "" {
		return typ.String(), nil
	}
	return fmt.Sprintf("%s:%s", typ.PkgPath(), typ.String()), nil
}
