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
