package httputil_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/crypto"
	"github.com/magic-lib/go-plat-utils/utils/httputil"
)

func signedRequest(t *testing.T, secret, nonce, signature string) *http.Request {
	ts := time.Now().Unix()
	r := httptest.NewRequest(http.MethodGet, "/api/x?a=1", nil)
	if signature == "" {
		sign, _, err := crypto.SignatureParamsByHmac(secret, &crypto.SignParams{
			Method: http.MethodGet, Path: "/api/x", Query: r.URL.Query(), Body: []byte(`{"a":"1"}`), Timestamp: ts, Nonce: nonce,
		})
		if err != nil {
			t.Fatalf("SignatureParamsByHmac: %v", err)
		}
		signature = sign
	}
	r.Header.Set("X-App-Id", "app1")
	r.Header.Set("X-Timestamp", strconv.FormatInt(ts, 10))
	r.Header.Set("X-Nonce", nonce)
	r.Header.Set("X-Signature", signature)
	return r
}

// 同一 nonce 只能使用一次；伪造签名不消耗 nonce；getSecret 为 nil 报错
func TestCheckSignatureNonce(t *testing.T) {
	var mu sync.Mutex
	seen := map[string]int64{}
	cacheNonce := func(ctx context.Context, key string) (int64, error) {
		mu.Lock()
		defer mu.Unlock()
		seen[key]++
		return seen[key], nil
	}
	getSecret := func(ctx context.Context, appId string) string { return "secret" }

	if ok, err := httputil.CheckSignature(signedRequest(t, "secret", "n1", "bad"), time.Minute, cacheNonce, getSecret); ok || err == nil {
		t.Fatalf("伪造签名应失败: %v %v", ok, err)
	}
	if ok, err := httputil.CheckSignature(signedRequest(t, "secret", "n1", ""), time.Minute, cacheNonce, getSecret); !ok || err != nil {
		t.Fatalf("首次合法请求应通过（伪造签名不应消耗 nonce）: %v %v", ok, err)
	}
	if ok, err := httputil.CheckSignature(signedRequest(t, "secret", "n1", ""), time.Minute, cacheNonce, getSecret); ok || err == nil {
		t.Fatalf("重放请求应被拒绝: %v %v", ok, err)
	}
	if ok, err := httputil.CheckSignature(signedRequest(t, "secret", "n2", ""), time.Minute, cacheNonce, nil); ok || err == nil {
		t.Fatalf("getSecret 为 nil 应报错: %v %v", ok, err)
	}
}
