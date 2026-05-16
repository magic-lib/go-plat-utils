package sign

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"net/http"
	"time"
)

// BaseSignDto 基础签名参数结构体
// 前端必须传这三个参数：timestamp, nonce, sign
// app_id=&app_token=&nonce=&sign_method=SHA%2D256&timestamp=  //按这个顺序进行签名
type BaseSignDto struct {
	AppId      string `json:"app_id" form:"app_id" binding:"required"`           // 客户端
	AppToken   string `json:"app_token"`                                         // 客户端秘钥
	Timestamp  string `json:"timestamp" form:"timestamp" binding:"required"`     // 10位秒级时间戳
	Nonce      string `json:"nonce" form:"nonce" binding:"required"`             // 随机字符串
	Signature  string `json:"signature" form:"signature" binding:"required"`     // 签名
	SignMethod string `json:"sign_method" form:"sign_method" binding:"required"` // 签名方法
}

func checkNonce(ctx context.Context, appID, nonce string, getStartTimeFunc func(ctx context.Context, appId, nonce string) (time.Time, error), ttl time.Duration) (pass bool, err error) {
	if nonce == "" || appID == "" {
		return false, fmt.Errorf("%s", "app_id or nonce is empty")
	}
	if getStartTimeFunc == nil {
		return false, nil
	}
	startTime, err := getStartTimeFunc(ctx, appID, nonce)
	if err != nil {
		return false, err
	}

	diff := time.Since(startTime)
	if diff < ttl {
		return true, nil
	}
	return false, nil
}

func checkTimestamp(ts string, window time.Duration) bool {
	if ts == "" {
		return false
	}
	sec, err := conv.Convert[int64](ts)
	if err != nil {
		return false
	}
	t := time.Unix(sec, 0)
	diff := time.Since(t)
	if diff < 0 {
		diff = -diff
	}
	return diff <= window
}

func getBaseSignDtoFromHeader(headerPrefix string, header http.Header) (*BaseSignDto, error) {
	if header == nil {
		return nil, fmt.Errorf("header is nil")
	}
	if headerPrefix == "" {
		headerPrefix = "X-"
	}

	baseSign := &BaseSignDto{}

	clientIdKey := http.CanonicalHeaderKey(headerPrefix + "App-Id")
	timestampKey := http.CanonicalHeaderKey(headerPrefix + "Timestamp")
	nonceKey := http.CanonicalHeaderKey(headerPrefix + "Nonce")
	signatureKey := http.CanonicalHeaderKey(headerPrefix + "Signature")
	signMethodKey := http.CanonicalHeaderKey(headerPrefix + "Sign-Method")

	baseSign.AppId = header.Get(clientIdKey)
	if baseSign.AppId == "" {
		return nil, fmt.Errorf("app_id is empty in header")
	}

	baseSign.Timestamp = header.Get(timestampKey)
	if baseSign.Timestamp == "" {
		return nil, fmt.Errorf("timestamp is empty in header")
	}

	baseSign.Nonce = header.Get(nonceKey)
	if baseSign.Nonce == "" {
		return nil, fmt.Errorf("nonce is empty in header")
	}

	baseSign.SignMethod = header.Get(signMethodKey)
	if baseSign.SignMethod == "" {
		return nil, fmt.Errorf("sign_method is empty in header")
	}
	baseSign.Signature = header.Get(signatureKey)
	if baseSign.Signature == "" {
		return nil, fmt.Errorf("signature is empty in header")
	}

	return baseSign, nil
}
