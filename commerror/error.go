package commerror

import (
	"errors"
	"github.com/magic-lib/go-plat-utils/commerror/errpb"
	"github.com/magic-lib/go-plat-utils/conf"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

// CommError 通用错误信息
type CommError interface {
	Error() string
	Code() int
}

type commErr struct {
	code    int    `json:"code"`
	message string `json:"message"`
}

// Error 错误信息返回，实现error接口
func (err *commErr) Error() string {
	if err == nil {
		return conf.EmptyString
	}
	return err.message
}
func (err *commErr) Code() int {
	if err == nil {
		return conf.DefaultErrorCode
	}
	return err.code
}

// New 新建错误对象
func New(msg string, code ...int) *commErr {
	err := &commErr{
		code:    conf.DefaultErrorCode,
		message: msg,
	}
	if len(code) > 0 {
		err.code = code[0]
	}
	return err
}

// Wrap 新增error Code
func Wrap(err error, code ...int) error {
	if err == nil {
		return nil
	}
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	var tempCode int = conf.DefaultErrorCode
	var errTemp CommError
	if errors.As(err, &errTemp) {
		tempCode = errTemp.Code()
	} else {
		if len(code) > 0 {
			tempCode = code[0]
		}
	}
	return New(errStr, tempCode)
}

func GrpcError(code codes.Code, msg string, details ...CommError) error {
	codesCode := codes.Internal
	codesMsg := codesCode.String()
	//系统默认的值，msg不能进行改变
	if code >= codes.OK &&
		code <= codes.Unauthenticated {
		codesCode = code
		codesMsg = codesCode.String()
	} else {
		// 自定义的值
		oneError := New(msg, int(code))
		if len(details) == 0 {
			details = []CommError{oneError}
		} else {
			details = append(details, oneError)
		}
	}
	if len(details) == 0 {
		return status.Error(codesCode, codesMsg)
	}
	s := status.New(codesCode, codesMsg)
	errDetails := make([]protoadapt.MessageV1, 0)
	lo.ForEach(details, func(item CommError, index int) {
		errDetails = append(errDetails, &errpb.ErrorDetail{
			Code:    int32(item.Code()),
			Message: item.Error(),
		})
	})
	s, _ = s.WithDetails(errDetails...)
	return s.Err()
}
func FromGrpcError(err error) (CommError, codes.Code) {
	if err == nil {
		return nil, codes.OK
	}
	s, ok := status.FromError(err)
	if ok && s != nil {
		for _, d := range s.Details() {
			if detail, ok := d.(*errpb.ErrorDetail); ok {
				if codes.Code(detail.Code) == codes.OK { //如果是正确的，则直接返回nil
					return nil, codes.OK
				}
				return New(detail.Message, int(detail.Code)), s.Code()
			}
		}
		return New(s.Message(), int(s.Code())), s.Code()
	}

	var commErr CommError
	if errors.As(err, &commErr) {
		return commErr, codes.Unknown
	}
	return New(err.Error()), codes.Unknown
}
