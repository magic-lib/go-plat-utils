package cond

import (
	"database/sql"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"strings"
	"sync"
)

// IsNotFoundErr 统一判断"记录不存在（sql: no rows in result set）"错误。
//
// 为什么不能只用 errors.Is(err, sqlx.ErrNotFound)：
//   - 本进程内直接查库时，err 就是 sqlx.ErrNotFound（= sql.ErrNoRows）本身，errors.Is 能命中；
//   - 但跨 gRPC 调用时，服务端返回的错误只会被序列化成 (code, message, details) 传到客户端，
//     客户端据此新建出一个 *status.statusError，它与 sqlx.ErrNotFound 既不是同一个对象、
//     也没有任何 Unwrap 关系，所以 errors.Is 一定返回 false，只能从 status 的 Code/Message 里认。
//
// 判断顺序：本地 errors.Is → gRPC code → gRPC message → 字符串兜底。
func IsNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	// 1) 本进程内的错误（直接查库 / 单测里构造的哨兵错误）
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}

	//errStr := err.Error() // rpc error: code = Unknown desc = sql: no rows in result set
	//fmt.Println(errStr)
	// 2) gRPC 错误：优先认 Code（服务端改造后走这条），再认 Message（当前服务端直接 return err 走这条）
	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.NotFound {
			return true
		}
		if msg := st.Message(); msg == sql.ErrNoRows.Error() {
			return true
		}
	}
	// 3) 兜底：错误被中间件又包了一层、status.FromError 认不出来（ok=false）时，退化为字符串匹配
	return strings.Contains(err.Error(), "no rows in result set")
}

// errorTypeCache 缓存「类型 → IsError 判定结果」。
// NumMethod/Method(i) 走反射且 Method(i) 会构造 reflect.Value（有分配），
// 同一类型的判定结果恒定不变，按 reflect.Type 缓存可避免重复计算。
var errorTypeCache sync.Map // map[reflect.Type]bool

// IsError 判断 val 是否是一个「纯粹的」错误对象，需要同时满足：
//  1. val 实现了 error 接口（方法集内含签名为 Error() string 的方法）；
//  2. 该类型的**导出**方法只允许是 Error，以及可选的 Unwrap；
//  3. 该类型没有**可导出**的字段（即序列化后不会有别的内容输出）。
//
// 之所以加第 2 条：Go 的接口是隐式实现，业务结构体只要碰巧定义了
// Error() string 就会被 val.(error) 命中；而这类结构体通常还带有
// GetXxx/ToXxx/MarshalJSON 等其它导出方法，据此可把它们与真正的错误区分开。
//
// Unwrap 之所以放行：它是 errors 包约定的错误链方法（Unwrap() error），
// 带它说明该类型仍是一条纯粹的错误链的成员，而非某个碰巧有 Error 方法的业务对象。
// errors.New/fmt.Errorf（含 %w 包装）/errors.Join 的结果都会被判为 true。
//
// 之所以加第 3 条：第 2 条只看方法，一个只定义了 Error 方法、却带着可导出字段的结构体
// 仍会被判为纯错误，于是调用方只能拿到 Error() 的文本而丢掉了字段内容。
// 这类对象序列化后是有内容的（有别的输出元素），应当被当作业务对象输出，故返回 false。
// 典型如 `type BizErr struct{ Code int; Msg string }`，只有 Error 方法但有可导出字段 → false。
//
// 未导出方法不计入统计；嵌入结构体提升上来的导出方法计入。
// 字段侧：未导出字段忽略，json:"-" 的字段视为不输出，同样忽略；指针会逐层解引用。
//
// 注意：判定基于 val 的动态类型本身。若 Error 使用指针接收者，则传值 T{}
// 连条件 1 都不满足（返回 false），需传 &T{}；这与 Go 自身的方法集规则一致。
func IsError(val any) bool {
	if val == nil {
		return false
	}
	if _, ok := val.(error); !ok {
		return false
	}

	typ := reflect.TypeOf(val)
	if cached, ok := errorTypeCache.Load(typ); ok {
		return cached.(bool)
	}
	ret := hasOnlyErrorMethods(typ) && !hasExportableField(typ)
	errorTypeCache.Store(typ, ret)
	return ret
}

// hasOnlyErrorMethods 判断 typ 的导出方法是否只由 Error 和 Unwrap 组成，且 Error 必须存在
func hasOnlyErrorMethods(typ reflect.Type) bool {
	hasError := false
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if !method.IsExported() { // 未导出方法忽略
			continue
		}
		switch method.Name {
		case "Error":
			hasError = true
		case "Unwrap":
			// 错误链约定方法，放行
		default:
			return false // 出现第 3 方导出方法，立即结束
		}
	}
	return hasError
}

// hasExportableField 判断 typ（逐层解引用后）是否存在「可导出且会被输出」的字段。
// 只要存在这样的字段，就说明该错误对象序列化后还有别的内容，不是纯粹的错误。
// json:"-" 表示该字段不参与序列化，不计入；非结构体类型一律视为没有字段。
func hasExportableField(typ reflect.Type) bool {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() { // 未导出字段不会被输出
			continue
		}
		if field.Tag.Get("json") == "-" { // 显式声明不输出
			continue
		}
		return true
	}
	return false
}