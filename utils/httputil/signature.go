package httputil

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/crypto"
	"github.com/magic-lib/go-plat-utils/utils/httputil/param"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type signatureHeader struct {
	AppID      string `json:"X-App-Id"`
	Timestamp  int64  `json:"X-Timestamp"`
	Nonce      string `json:"X-Nonce"`
	SignMethod string `json:"X-Sign-Method"`
	Signature  string `json:"X-Signature"`
}

func getSignatureHeader(hd http.Header) *signatureHeader {
	sh := new(signatureHeader)
	hd2 := make(map[string]string)
	for k, _ := range hd {
		hd2[param.CanonicalHeaderKey(k)] = hd.Get(k)
	}
	_ = conv.Unmarshal(hd2, sh)
	if sh.SignMethod == "" {
		sh.SignMethod = "HMAC-SHA256"
	}
	return sh
}

func checkTimestamp(sec int64, window time.Duration) bool {
	if sec == 0 {
		return false
	}
	tsTime := time.Unix(sec, 0)
	diff := time.Since(tsTime)
	if diff < 0 {
		diff = -diff
	}
	return diff <= window
}
func checkSignatureHeader(sh *signatureHeader, limitTime time.Duration) (bool, error) {
	if limitTime == 0 {
		limitTime = 5 * time.Minute
	}
	if sh.AppID == "" || sh.Timestamp == 0 || sh.Nonce == "" || sh.Signature == "" {
		return false, fmt.Errorf("missing signature headers")
	}
	if !strings.EqualFold(sh.SignMethod, "HMAC-SHA256") {
		return false, fmt.Errorf("unsupported sign method")
	}
	if !checkTimestamp(sh.Timestamp, limitTime) {
		return false, fmt.Errorf("timestamp expired")
	}
	return true, nil
}

func buildSignParams(r *http.Request) *crypto.SignParams {
	sh := getSignatureHeader(r.Header)
	bodyStr := param.NewParam().GetAllBody(r)
	queryValues, _ := url.ParseQuery(r.URL.RawQuery)
	p := &crypto.SignParams{
		Method:    r.Method,
		Path:      r.URL.Path, // 不带域名、query
		Query:     queryValues,
		Body:      []byte(bodyStr),
		Timestamp: sh.Timestamp,
		Nonce:     sh.Nonce,
	}
	return p
}

func checkNonce(ctx context.Context, cacheNonceFunc func(ctx context.Context, cacheKey string) (int64, error), appId, nonce string, ttl time.Duration) (bool, error) {
	if nonce == "" {
		return false, nil
	}
	if cacheNonceFunc == nil {
		return true, nil
	}
	cacheKey := fmt.Sprintf("%s/%s", appId, nonce)
	cacheVal, err := cacheNonceFunc(ctx, cacheKey)
	if err != nil {
		return false, err
	}
	if cacheVal == 0 {
		return false, fmt.Errorf("cache not set error")
	}
	retTime := checkTimestamp(cacheVal, ttl)
	if retTime {
		return true, nil
	}
	return false, nil
}

func CheckSignature(r *http.Request, ttl time.Duration, cacheNonceFunc func(ctx context.Context, cacheKey string) (int64, error), getSecret func(ctx context.Context, appId string) string) (bool, error) {
	sh := getSignatureHeader(r.Header)
	_, err := checkSignatureHeader(sh, ttl)
	if err != nil {
		return false, err
	}

	// 校验 nonce，防止重放
	if ok, err := checkNonce(r.Context(), cacheNonceFunc, sh.AppID, sh.Nonce, ttl); err != nil || !ok {
		return false, fmt.Errorf("nonce reused")
	}
	secret := getSecret(r.Context(), sh.AppID)
	p := buildSignParams(r)
	checkServerSign, err := crypto.VerifySignatureParamsByHmac(secret, p, sh.Signature)
	if err != nil {
		return false, err
	}
	// 比对签名
	if !checkServerSign {
		return false, fmt.Errorf("invalid signature")
	}
	return true, nil
}
