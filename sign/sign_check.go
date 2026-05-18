package sign

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	cryptotool "github.com/magic-lib/go-plat-utils/crypto"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/magic-lib/go-plat-utils/utils/httputil/param"
	"github.com/samber/lo"
	"net/http"
	"strings"
	"time"
)

type SingChecker interface {
	GeneratorSignature(ctx context.Context, baseSign *BaseSignDto, params map[string]any) (map[string]any, error)
	BaseSignDtoFromHeader(header http.Header) (*BaseSignDto, error)
	CheckSignature(ctx context.Context, header http.Header, params map[string]any) (bool, error)
	CheckHttpSignature(r *http.Request) (bool, error)
}

// BaseSignCheck 基础签名参数结构体
type BaseSignCheck struct {
	GetAppTokenFunc           func(ctx context.Context, appId string) (string, error)           //根据appId获取GetAppTokenFunc
	GetNonceCreateTimeFunc    func(ctx context.Context, appId, nonce string) (time.Time, error) //从缓存或数据库中获取nonce的创建时间
	HttpRequestBodyEncodeFunc func(ctx context.Context, body []byte) (string, error)            //httpBody编码函数，避免内容太长影响效率
	HeaderKeyPrefix           string                                                            //header中key的前缀名称
	NonceTimeout              time.Duration                                                     // nonce缓存的过期时间
	TimestampTimeout          time.Duration                                                     // 时间戳的过期时间，超过就不让访问了，默认为5分钟
}

func New(bs *BaseSignCheck) (SingChecker, error) {
	if bs == nil {
		bs = &BaseSignCheck{}
	}
	if bs.HeaderKeyPrefix == "" {
		bs.HeaderKeyPrefix = "X-"
	}
	if bs.NonceTimeout == 0 {
		bs.NonceTimeout = 2 * time.Minute
	}
	if bs.TimestampTimeout == 0 {
		bs.TimestampTimeout = 5 * time.Minute
	}
	if bs.GetAppTokenFunc == nil {
		return nil, fmt.Errorf("%s", "GetAppTokenFunc not set")
	}
	return bs, nil
}

func (bs *BaseSignCheck) GeneratorSignature(ctx context.Context, baseSign *BaseSignDto, params map[string]any) (map[string]any, error) {
	if bs.GetAppTokenFunc == nil {
		return nil, fmt.Errorf("AppSecretFunc is empty")
	}
	appSecret, err := bs.GetAppTokenFunc(ctx, baseSign.AppId)
	if err != nil {
		return nil, err
	}
	if appSecret == "" {
		return nil, fmt.Errorf("secret is empty")
	}
	baseSign.AppToken = appSecret

	encodeType, err := cryptotool.StringToHash(baseSign.SignMethod)
	if err != nil {
		return nil, err
	}

	newParams := make(map[string]any)
	_ = conv.Unmarshal(baseSign, &newParams)
	paramNew := lo.Assign(params, newParams)
	delete(paramNew, "signature")

	oldStr := utils.MapToUrlParams(paramNew)

	sign := cryptotool.SHA(encodeType, oldStr)
	if sign == "" {
		return paramNew, fmt.Errorf("sign method is empty: %s", encodeType)
	}
	paramNew["signature"] = sign
	return paramNew, nil
}

func (bs *BaseSignCheck) BaseSignDtoFromHeader(header http.Header) (*BaseSignDto, error) {
	baseSign, err := baseSignDtoFromHeader(bs.HeaderKeyPrefix, header)
	if err != nil {
		return nil, err
	}
	return baseSign, nil
}

// checkSignatureByBaseSignDto 从header中进行签名验证
func (bs *BaseSignCheck) checkSignatureByBaseSignDto(ctx context.Context, baseSign *BaseSignDto, params map[string]any) (bool, error) {
	if bs.GetAppTokenFunc == nil {
		return false, fmt.Errorf("AppSecretFunc is empty")
	}
	if params == nil {
		params = make(map[string]any)
	}
	resultParams, err := bs.GeneratorSignature(ctx, baseSign, params)
	if err != nil {
		return false, err
	}

	generatedSignature, ok := resultParams["signature"]
	if !ok {
		return false, fmt.Errorf("generated signature is empty")
	}

	return strings.EqualFold(conv.String(generatedSignature), baseSign.Signature), nil
}

func (bs *BaseSignCheck) CheckSignature(ctx context.Context, header http.Header, params map[string]any) (bool, error) {
	baseSign, err := bs.BaseSignDtoFromHeader(header)
	if err != nil {
		return false, err
	}
	err = checkBaseSignDto(baseSign)
	if err != nil {
		return false, err
	}
	// 首先检查timestamp
	checkTime := checkTimestamp(baseSign.Timestamp, bs.TimestampTimeout)
	if !checkTime {
		return false, nil
	}
	// 检查nonce
	checkNonceTemp, err := checkNonce(ctx, baseSign.AppId, baseSign.Nonce, bs.GetNonceCreateTimeFunc, bs.NonceTimeout)
	if err != nil {
		return false, err
	}
	if !checkNonceTemp {
		return false, nil
	}
	checkSignature, err := bs.checkSignatureByBaseSignDto(ctx, baseSign, params)
	if err != nil {
		return false, err
	}
	if !checkSignature {
		return false, nil
	}
	return true, nil
}
func (bs *BaseSignCheck) checkHttpSignature(ctx context.Context, header http.Header, httpParams *Params) (bool, error) {
	newHttpParam := buildAllSortedParams(ctx, httpParams, bs.HttpRequestBodyEncodeFunc)
	return bs.CheckSignature(ctx, header, newHttpParam)
}

func (bs *BaseSignCheck) CheckHttpSignature(r *http.Request) (bool, error) {
	p := new(Params)
	p.Method = r.Method
	p.Path = r.URL.Path
	p.Query = r.URL.Query()
	p.Body = param.SafeReadBody(r, nil)
	return bs.checkHttpSignature(r.Context(), r.Header, p)
}
