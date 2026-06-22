package httputil

import (
	"bytes"
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"net/url"
)

const (
	jsonContentType = "application/json; charset=utf-8"
)

// CommResponse 接口返回值
type CommResponse struct {
	Code        int64  `json:"code"`
	Message     string `json:"message"`
	InternalMsg string `json:"internal_msg,omitempty"` // 内部消息
	Now         string `json:"now,omitempty"`
	Env         string `json:"env,omitempty"` // 环境
	Time        int64  `json:"time,omitempty"`
	LogId       string `json:"log_id,omitempty"`
	TraceId     string `json:"trace_id,omitempty"`
	Params      any    `json:"params,omitempty"`
	Debug       any    `json:"debug,omitempty"`
	Data        any    `json:"data"`

	// 对原始响应字符串做二次处理的函数
	ProcessResp func(response *CommResponse, resp string) (string, error) `json:"-"`
}

// PageModel 分页结构输出
type PageModel struct {
	Count      int64 `json:"count"`                 // 数据总数
	PageNow    int   `json:"page_now,omitempty"`    // 当前页数
	PageStart  uint  `json:"page_start,omitempty"`  // 当前页第一条数据的全局序号（从1开始)
	PageEnd    uint  `json:"page_end,omitempty"`    // 当前页最后一条数据的全局序号， 前端快速实现「共 XX 条，展示第 XX~XX 条」的文案展示
	PageOffset int   `json:"page_offset,omitempty"` // 偏移下标（切片start索引，从0开始）
	PageSize   int   `json:"page_size,omitempty"`   // 每页显示的数目
	PageTotal  int   `json:"page_total,omitempty"`  // 总页数
	DataList   any   `json:"data_list"`             // 当前页数据
}

type UploadFile struct {
	Filename string
	Size     int
	Buffer   *bytes.Buffer
}

// GetPage 校验并计算分页基础参数
// maxPageSize：单页最大条数限制，防止前端传入超大pageSize拉全量
func (p *PageModel) GetPage(maxPageSize int) *PageModel {
	if maxPageSize <= 0 {
		maxPageSize = 50
	}
	if p.PageNow <= 0 {
		p.PageNow = 1
	}
	// 每页条数默认上限
	if p.PageSize <= 0 || p.PageSize >= maxPageSize {
		p.PageSize = maxPageSize
	}

	total := int(p.Count)
	if total > 0 {
		// 计算整除的结果
		p.PageTotal = total / p.PageSize
		// 计算余数
		if total%p.PageSize > 0 {
			p.PageTotal++
		}

		if p.PageNow > p.PageTotal {
			p.PageNow = p.PageTotal
		}
	}
	p.PageOffset = (p.PageNow - 1) * p.PageSize

	// 计算前端友好的起始、结束序号（从1开始）
	p.PageStart = uint(p.PageOffset) + 1
	endIdx := p.PageOffset + p.PageSize
	if endIdx > total {
		endIdx = total
	}
	p.PageEnd = uint(endIdx)

	return p
}

// SlicePaginate 内存切片通用分页（推荐替换原StaticPageList）
// T 切片元素泛型
// list：内存中已查询/排序完成的全量切片
// page：分页请求参数结构体
// maxPageSize：可选，限制单页最大条数，默认50
func SlicePaginate[T any](list []T, page *PageModel, maxPageSize ...int) *PageModel {
	if page == nil {
		page = new(PageModel)
	}

	maxSize := 0
	if len(maxPageSize) > 0 {
		maxSize = maxPageSize[0]
	}
	page.Count = int64(len(list))
	page = page.GetPage(maxSize)
	if len(list) == 0 {
		page.DataList = []T{}
		return page
	}
	// 计算起始、结束下标
	start := page.PageOffset
	end := start + page.PageSize

	// 边界保护
	if start < 0 {
		start = 0
	}
	total := int(page.Count)
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	page.DataList = list[start:end]
	return page
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
func WriteCommResponse(w http.ResponseWriter, comm *CommResponse, code ...int) error {
	response := comm.withNowTime()

	respStr := conv.String(response)
	if comm.ProcessResp != nil {
		respStrTemp, err := comm.ProcessResp(response, respStr)
		if err == nil {
			respStr = respStrTemp
		}
	}

	{ //处理列表不要返回null的问题
		dataList := gjson.Get(respStr, "data.data_list")
		if dataList.Exists() && !dataList.IsArray() {
			//判断dataList 是否是nil
			if cond.IsNil(dataList.Value()) {
				respStr2, err := sjson.Set(respStr, "data.data_list", []any{})
				if err == nil {
					respStr = respStr2
				}
			}
		}
	}

	w.Header().Set("Content-Type", jsonContentType)
	respByte := []byte(respStr)

	oneStatusCode := http.StatusOK
	if len(code) > 0 {
		oneStatusCode = code[0]
	}
	w.WriteHeader(oneStatusCode)

	_, err := w.Write(respByte)

	return err
}

// TraceId 获取traceId
func TraceId(ctx context.Context) string {
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
func WriteCommFailure(ctx context.Context, w http.ResponseWriter, err error, code int64, statusCode ...int) {
	errResp := GetErrorResponse(nil, code, err)
	writeWithTrace(ctx, errResp, func(span trace.Span) {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
	})
	_ = WriteCommResponse(w, errResp, statusCode...)
}

func writeWithTrace(ctx context.Context, resp *CommResponse, traceFunc func(span trace.Span)) {
	traceId := TraceId(ctx)
	if traceId == "" {
		return
	}
	resp = resp.withTraceId(traceId)
	span := trace.SpanFromContext(ctx)
	if span != nil && traceFunc != nil {
		traceFunc(span)
	}
}

// WriteCommSuccess 系统默认正确返回
func WriteCommSuccess(ctx context.Context, w http.ResponseWriter, data any, msg ...string) {
	retMsg := http.StatusText(http.StatusOK)
	if len(msg) > 0 {
		retMsg = msg[0]
	}
	sucResp := &CommResponse{
		Code:    0,
		Message: retMsg,
		Data:    data,
	}

	if oneData, ok := data.(CommResponse); ok {
		sucResp = &oneData
	} else if oneDataPtr, ok := data.(*CommResponse); ok {
		sucResp = oneDataPtr
	}
	if sucResp.Message == "" {
		sucResp.Message = retMsg
	}

	writeWithTrace(ctx, sucResp, func(span trace.Span) {
		span.SetStatus(codes.Ok, sucResp.Message)
	})

	_ = WriteCommResponse(w, sucResp, http.StatusOK)
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

func WriteFileSuccess(_ context.Context, w http.ResponseWriter, file *UploadFile) error {
	if file == nil || file.Buffer == nil {
		return fmt.Errorf("file is nil")
	}
	if file.Size <= 0 {
		file.Size = file.Buffer.Len()
	}
	if file.Size == 0 {
		log.Println("WriteFileSuccess file is empty")
		return fmt.Errorf("file size is 0")
	}

	escapedFileName := url.PathEscape(file.Filename)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", escapedFileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	_, err := file.Buffer.WriteTo(w)
	if err != nil {
		log.Println("WriteFileSuccess failed to write buffer to response:", err.Error())
		return err
	}
	return nil
}
