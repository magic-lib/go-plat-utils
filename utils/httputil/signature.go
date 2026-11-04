package httputil

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/crypto"
	"github.com/magic-lib/go-plat-utils/utils/httputil/param"
	"net/http"
	"net/url"
	"sort"
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

// checkNonce 校验 nonce 是否首次出现。cacheNonceFunc 需记录 cacheKey 并返回其被记录的次数（含本次），
// 如 redis INCR 后对新 key 设置 EXPIRE ttl：返回 1 表示首次出现，大于 1 表示重放。
// 原实现把返回值当作首次出现的时间戳、仍在 ttl 内即放行，ttl 内的重放请求同样能通过。
func checkNonce(ctx context.Context, cacheNonceFunc func(ctx context.Context, cacheKey string) (int64, error), appId, nonce string) (bool, error) {
	if nonce == "" {
		return false, fmt.Errorf("nonce is empty")
	}
	if cacheNonceFunc == nil {
		return true, nil
	}
	cacheKey := fmt.Sprintf("%s/%s", appId, nonce)
	seenTimes, err := cacheNonceFunc(ctx, cacheKey)
	if err != nil {
		return false, fmt.Errorf("check nonce: %w", err)
	}
	if seenTimes <= 0 {
		return false, fmt.Errorf("check nonce: cache not set")
	}
	if seenTimes > 1 {
		return false, fmt.Errorf("nonce reused")
	}
	return true, nil
}

// CheckSignature 校验请求签名。cacheNonceFunc 返回 nonce 被记录的次数（含本次），见 checkNonce；为 nil 时不做防重放。
func CheckSignature(r *http.Request, ttl time.Duration, cacheNonceFunc func(ctx context.Context, cacheKey string) (int64, error), getSecret func(ctx context.Context, appId string) string) (bool, error) {
	sh := getSignatureHeader(r.Header)
	_, err := checkSignatureHeader(sh, ttl)
	if err != nil {
		return false, err
	}
	if getSecret == nil {
		return false, fmt.Errorf("getSecret is nil")
	}
	secret := getSecret(r.Context(), sh.AppID)
	if secret == "" {
		return false, fmt.Errorf("secret is empty")
	}
	p := buildSignParams(r)
	checkServerSign, err := crypto.VerifySignatureParamsByHmac(secret, p, sh.Signature)
	if err != nil {
		return false, err
	}
	// 比对签名
	if !checkServerSign {
		return false, fmt.Errorf("invalid signature")
	}
	// 验签通过后再记录 nonce：否则伪造签名的请求也会消耗掉合法请求的 nonce
	if ok, err := checkNonce(r.Context(), cacheNonceFunc, sh.AppID, sh.Nonce); err != nil || !ok {
		if err == nil {
			err = fmt.Errorf("nonce reused")
		}
		return false, err
	}
	return true, nil
}

// GenFeiShuSign 生成飞书开放平台的签名参数。
// 签名算法：stringToSign = timestamp + "\n" + secret，sign = base64(hmac-sha256(stringToSign, secret))
func GenFeiShuSign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return sign
}

// GenWxPayV2Sign 微信支付V2 MD5签名
// params: 业务参数map，内部会自动剔除 sign 字段
// apiKey: 商户32位API密钥
func GenWxPayV2Sign(params map[string]string, apiKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	// key 字典序升序排序
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		v := params[k]
		if v == "" {
			continue // 空值不参与拼接
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
		b.WriteString("&")
	}
	// 拼接 &key=xxx
	b.WriteString("key=")
	b.WriteString(apiKey)
	str := b.String()

	h := md5.New()
	h.Write([]byte(str))
	signBytes := h.Sum(nil)
	// V2要求转大写
	return strings.ToUpper(hex.EncodeToString(signBytes))
}
