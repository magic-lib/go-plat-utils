package httputil

import (
	"context"
	"github.com/magic-lib/go-plat-utils/conv"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

// CommResponse 接口返回值
type CommResponse struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Now     string `json:"now,omitempty"`
	Env     string `json:"env,omitempty"` //环境
	Time    int64  `json:"time,omitempty"`
	LogId   string `json:"log_id,omitempty"`
	TraceId string `json:"trace_id,omitempty"`
	Params  any    `json:"params,omitempty"`
	Debug   any    `json:"debug,omitempty"`
	Data    any    `json:"data"`
}

// PageModel 分页结构输出
type PageModel struct {
	Count      int64 `json:"count"`                 // 数据总数
	PageNow    int   `json:"page_now,omitempty"`    // 当前页数
	PageStart  uint  `json:"page_start,omitempty"`  // 当前开始页数
	PageEnd    uint  `json:"page_end,omitempty"`    // 当前结束页数
	PageOffset int   `json:"page_offset,omitempty"` // 当前页面的偏移量
	PageSize   int   `json:"page_size,omitempty"`   // 每页显示的数目
	PageTotal  int   `json:"page_total,omitempty"`  // 总页数
	DataList   any   `json:"data_list"`             // 数据列表
}

func (p *PageModel) GetPage(maxPageSize int) *PageModel {
	if maxPageSize <= 0 {
		maxPageSize = 50
	}
	if p.PageNow <= 0 {
		p.PageNow = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = maxPageSize
	}
	if p.PageSize >= maxPageSize {
		p.PageSize = maxPageSize
	}

	if p.Count > 0 {
		// 计算整除的结果
		quotient := int(p.Count) / p.PageSize
		// 计算余数
		remainder := int(p.Count) % p.PageSize
		if remainder == 0 {
			p.PageTotal = quotient
		} else {
			p.PageTotal = quotient + 1
		}
		if p.PageNow > p.PageTotal {
			p.PageNow = p.PageTotal
		}
	}
	p.PageOffset = (p.PageNow - 1) * p.PageSize

	return p
}

// WithNowTime 获取通用的返回格式
func (comm *CommResponse) withNowTime() *CommResponse {
	comm.Now = conv.FormatFromUnixTime() //当前时间
	return comm
}

func (comm *CommResponse) withTraceId(traceId string) *CommResponse {
	if traceId != "" {
		comm.TraceId = traceId
	}
	return comm
}

// WriteCommResponse 将通用返回设置到response，输出到客户端
func WriteCommResponse(respWriter http.ResponseWriter, comm *CommResponse, statusCode ...int) error {
	response := comm.withNowTime()

	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	respStr := conv.String(response)
	respByte := []byte(respStr)

	oneStatusCode := http.StatusOK
	if len(statusCode) > 0 {
		oneStatusCode = statusCode[0]
	}
	respWriter.WriteHeader(oneStatusCode)

	_, err := respWriter.Write(respByte)

	return err
}

func getTraceId(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span != nil && span.IsRecording() {
		spanContext := span.SpanContext()
		if spanContext.IsValid() {
			traceID := spanContext.TraceID()
			return traceID.String()
		}
	}
	return ""
}

// WriteCommFailure 系统默认错误返回，需要加入ctx信息
func WriteCommFailure(ctx context.Context, respWriter http.ResponseWriter, err error, code int64, statusCode ...int) {
	errResp := GetErrorResponse(nil, code, err)
	traceId := getTraceId(ctx)
	if traceId != "" {
		errResp = errResp.withTraceId(traceId)
		span := trace.SpanFromContext(ctx)
		if span != nil && err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
	}
	_ = WriteCommResponse(respWriter, errResp, statusCode...)
}

// WriteCommSuccess 系统默认正确返回
func WriteCommSuccess(ctx context.Context, respWriter http.ResponseWriter, data any) {
	sucResp := &CommResponse{
		Code:    0,
		Message: http.StatusText(http.StatusOK),
		Data:    data,
	}
	traceId := getTraceId(ctx)
	if traceId != "" {
		sucResp = sucResp.withTraceId(traceId)
		span := trace.SpanFromContext(ctx)
		if span != nil {
			span.SetStatus(codes.Ok, sucResp.Message)
		}
	}
	_ = WriteCommResponse(respWriter, sucResp, http.StatusOK)
}

// GetErrorResponse 系统获取错误码和错误信息
func GetErrorResponse(allErrorMap map[int64]string, errorCode int64, err ...error) *CommResponse {
	respError := &CommResponse{}

	if errorCode == 0 {
		errorCode = http.StatusInternalServerError
	}

	respError.Code = errorCode

	if len(err) > 0 {
		respError.Message = err[0].Error()
	}

	if allErrorMap != nil {
		if errorMsg, ok := allErrorMap[errorCode]; ok {
			if respError.Message == "" {
				respError.Message = conv.String(errorMsg)
			}
			return respError
		}
	}

	//根据errorCode 获取错误信息
	if respError.Message == "" {
		respError.Message = http.StatusText(int(errorCode))
	}

	if respError.Message == "" {
		respError.Message = http.StatusText(http.StatusInternalServerError)
	}

	return respError
}
