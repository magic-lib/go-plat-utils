package httputil

import (
	"github.com/magic-lib/go-plat-utils/conv"
	"net/http"
)

// CommResponse 接口返回值
type CommResponse struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Now     string      `json:"now,omitempty"`
	Env     string      `json:"env,omitempty"` //环境
	Time    int64       `json:"time,omitempty"`
	LogId   string      `json:"logId,omitempty"`
	TraceId string      `json:"traceId,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Debug   interface{} `json:"debug,omitempty"`
	Data    interface{} `json:"data"`
}

// PageModel 分页结构输出
type PageModel struct {
	Count      int64       `json:"count"`                // 数据总数
	PageNow    int         `json:"pageNow,omitempty"`    // 当前页数
	PageOffset int         `json:"pageOffset,omitempty"` // 当前页面的偏移量
	PageSize   int         `json:"pageSize,omitempty"`   // 每页显示的数目
	PageTotal  int         `json:"pageTotal,omitempty"`  // 总页数
	DataList   interface{} `json:"dataList"`             // 数据列表
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

// WriteCommResponse 将通用返回设置到response，输出到客户端
func WriteCommResponse(respWriter http.ResponseWriter, comm *CommResponse, statusCode ...int) error {
	response := comm.withNowTime()

	contentType := "Content-Type"
	respWriter.Header().Set(contentType, "application/json; charset=utf-8")

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

// GetErrorResponse 系统获取错误码和错误信息
func GetErrorResponse(allErrorMap map[int64]string, errorCode int64, err ...error) *CommResponse {
	respError := &CommResponse{}

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

	if respError.Message == "" {
		respError.Message = "系统错误"
	}

	return respError
}
