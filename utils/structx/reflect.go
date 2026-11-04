package structx

import "reflect"

// NewInstance 创建泛型 T 的实例。如果 T 是指针类型，则创建其指向类型的实例并返回指针。
func NewInstance[T any]() T {
	var t T
	// 用 TypeFor 取类型：T 为接口时 reflect.TypeOf(t) 是 nil，调用 Kind() 会 panic
	tType := reflect.TypeFor[T]()

	switch tType.Kind() {
	case reflect.Ptr:
		// 创建 T 对应的非指针类型的实例，并取其地址作为新的 T
		t = reflect.New(tType.Elem()).Interface().(T)
	case reflect.Slice:
		// 切片类型创建一个非 nil 的空切片；数组的零值本身可用（原来对数组建切片再断言会 panic）
		t = reflect.MakeSlice(tType, 0, 0).Interface().(T)
	default:
	}
	return t
}
