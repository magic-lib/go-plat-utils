package conv

const (
	errStrGetStringFromJson    = "getStringFromJson error:%w"
	errStrUnmarshal1           = "unmarshal DstPoint is %s"
	errStrNotSlice             = "getByDstSlice is not slice:%s"
	errStrGetByDstPtr          = "getByDstPtr is not ptr: %s"
	errStrRecover              = "continueAssignTo error: %v"
	errStrRecover2             = "getByDstStruct error: %v"
	errStrNotPointer           = "continueAssignTo dstPoint is not pointer: %s"
	errStrNotPointer2          = "assignTo dstPoint is not pointer: %s"
	errStrGetByDstOther        = "getByDstOther error: %v"
	errStrGetByDstMap          = "getByDstMap is not string: %s"
	errStrGetByDstMapNotMap    = "getByDstMap is not Map: %s"
	errStrGetByDstMapNotStruct = "getByDstStruct is not Struct: %s"
)

// JS 类型常量（与 ConvertForTypeJs 支持的 jsType 对应，忽略大小写与首尾空格）
const (
	JsTypeString    = "string"
	JsTypeBoolean   = "boolean"
	JsTypeNumber    = "number"
	JsTypeBigInt    = "bigint"
	JsTypeNull      = "null"
	JsTypeUndefined = "undefined"
	JsTypeArray     = "array"
	JsTypeObject    = "object"
	JsTypeDate      = "date"
)

// Go 类型常量（与 ConvertForTypeString 支持的 targetType 对应，忽略大小写与首尾空格）
const (
	GoTypeString  = "string"
	GoTypeBool    = "bool"
	GoTypeInt     = "int"
	GoTypeInt8    = "int8"
	GoTypeInt16   = "int16"
	GoTypeInt32   = "int32"
	GoTypeInt64   = "int64"
	GoTypeUint    = "uint"
	GoTypeUint8   = "uint8"
	GoTypeUint16  = "uint16"
	GoTypeUint32  = "uint32"
	GoTypeUint64  = "uint64"
	GoTypeFloat32 = "float32"
	GoTypeFloat64 = "float64"
	GoTypeDecimal = "decimal"
	GoTypeTime    = "time"
	GoTypeNil     = "nil"
	GoTypeSlice   = "slice"
	GoTypeMap     = "map"
)
