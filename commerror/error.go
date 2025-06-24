package commerror

import (
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/magic-lib/go-plat-utils/commerror/errpb"
	"github.com/magic-lib/go-plat-utils/conf"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

// CommError 通用错误信息
type CommError interface {
	error
	Code() int
	Cause() error
	Details() map[string]any
}

type commErr struct {
	code    int
	message string
	cause   error
	details map[string]any
}

// Error 错误信息返回，实现error接口
func (err *commErr) Error() string {
	if err == nil {
		return conf.EmptyString
	}
	causeMsg := conf.EmptyString
	if err.cause != nil {
		causeMsg = err.cause.Error()
	}

	formattedErr := struct {
		Code    int            `json:"code"`
		Message string         `json:"message"`
		Cause   string         `json:"cause"`
		Details map[string]any `json:"details"`
	}{
		Code:    err.Code(),
		Message: err.message,
		Cause:   causeMsg,
		Details: err.details,
	}

	errByte, errTemp := jsoniter.MarshalToString(formattedErr)
	if errTemp != nil {
		if causeMsg != "" {
			return fmt.Sprintf("code=%d, message=%s, cause=%s", err.code, err.message, causeMsg)
		}
		return fmt.Sprintf("code=%d, message=%s", err.code, err.message)
	}
	return errByte
}
func (err *commErr) Code() int {
	if err == nil {
		return conf.DefaultErrorCode
	}
	return err.code
}

func (err *commErr) Cause() error {
	return err.cause
}

func (err *commErr) Details() map[string]any {
	if len(err.details) == 0 {
		return make(map[string]any)
	}
	return err.details
}

// New 新建错误对象
func New(msg string, opts ...Option) CommError {
	err := &commErr{
		code:    conf.DefaultErrorCode,
		message: msg,
	}
	for _, opt := range opts {
		opt(err)
	}
	return err
}

// Option 错误构造选项
type Option func(*commErr)

// WithCode 设置错误码
func WithCode(code int) Option {
	return func(err *commErr) {
		err.code = code
	}
}

// WithCause 设置错误原因
func WithCause(cause error) Option {
	return func(err *commErr) {
		err.cause = cause
	}
}

// WithDetail 添加错误详情
func WithDetail(key string, value any) Option {
	return func(err *commErr) {
		if len(err.details) == 0 {
			err.details = make(map[string]any)
		}
		err.details[key] = value
	}
}

// Wrap 包装现有错误
func Wrap(err error, opts ...Option) CommError {
	if err == nil {
		return nil
	}

	// 如果已经是CommError，直接添加选项
	var commError CommError
	if errors.As(err, &commError) {
		newErr := &commErr{
			code:    commError.Code(),
			message: commError.Error(),
			cause:   commError.Cause(),
			details: commError.Details(),
		}

		for _, opt := range opts {
			opt(newErr)
		}

		return newErr
	}

	// 否则创建新的CommError
	newErr := &commErr{
		code:    conf.DefaultErrorCode,
		message: err.Error(),
		cause:   err,
		details: make(map[string]any),
	}

	for _, opt := range opts {
		opt(newErr)
	}

	return newErr
}

// WrapCode 兼容原有Wrap功能的快捷方式
func WrapCode(err error, code int) CommError {
	return Wrap(err, WithCode(code))
}

// GrpcError grpc错误信息
func GrpcError(code codes.Code, msg string, details ...CommError) error {
	codesCode := codes.Internal
	codesMsg := codesCode.String()
	//系统默认的值，msg不能进行改变
	if code <= codes.Unauthenticated {
		codesCode = code
		codesMsg = code.String()
	} else {
		// 自定义的值
		oneError := New(msg, WithCode(int(code)))
		if len(details) == 0 {
			details = []CommError{}
		}
		details = append(details, oneError)
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
	grpcStatus := codes.Unknown
	s, ok := status.FromError(err)
	if ok && s != nil {
		grpcStatus = s.Code()
		for _, d := range s.Details() {
			if detail, ok := d.(*errpb.ErrorDetail); ok {
				if codes.Code(detail.Code) == codes.OK { //如果是正确的，则直接返回nil
					return nil, codes.OK
				}
				return New(detail.Message, WithCode(int(detail.Code))), grpcStatus
			}
		}
		return New(s.Message(), WithCode(int(s.Code()))), grpcStatus
	}

	var commErr CommError
	if errors.As(err, &commErr) {
		return commErr, grpcStatus
	}
	return New(err.Error()), grpcStatus
}
