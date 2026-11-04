package sign

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/conv"
	cryptotool "github.com/magic-lib/go-plat-utils/crypto"
)

// 签名错误时不能消耗 nonce；签名比较不区分大小写
func TestCheckSignatureNonceAfterVerify(t *testing.T) {
	nonceCalls := 0
	svc, err := New(&BaseSignCheck{
		HeaderKeyPrefix: "X-",
		GetAppTokenFunc: func(ctx context.Context, appId string) (string, error) {
			return getAppToken(appId), nil
		},
		GetOrSetNonceFunc: func(ctx context.Context, appId, nonce string, timeout time.Duration) (int, error) {
			nonceCalls++
			return 0, nil
		},
		NoCheckTimestamp: true,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	dto := &BaseSignDto{AppId: "biu-tech", Timestamp: "1783090590", Nonce: "AoFgOQ", SignMethod: "MD5"}
	generated, err := svc.GeneratorSignature(context.Background(), dto, nil)
	if err != nil {
		t.Fatalf("GeneratorSignature: %v", err)
	}
	goodSign := conv.String(generated["signature"])

	header := func(sign string) http.Header {
		return http.Header{
			"X-App-Id": {"biu-tech"}, "X-Timestamp": {"1783090590"}, "X-Nonce": {"AoFgOQ"},
			"X-Signature": {sign}, "X-Sign-Method": {"MD5"},
		}
	}

	ok, err := svc.CheckSignature(context.Background(), header("0000000000000000000000000000000"), nil)
	if err != nil || ok {
		t.Fatalf("伪造签名 = %v, %v; want false", ok, err)
	}
	if nonceCalls != 0 {
		t.Fatalf("签名错误时不应消耗 nonce, 调用了 %d 次", nonceCalls)
	}

	ok, err = svc.CheckSignature(context.Background(), header(strings.ToUpper(goodSign)), nil)
	if err != nil || !ok {
		t.Fatalf("正确签名（大写）= %v, %v; want true", ok, err)
	}
	if nonceCalls != 1 {
		t.Fatalf("验签通过后应记录一次 nonce, 实际 %d 次", nonceCalls)
	}
}

func TestSignatureEqual(t *testing.T) {
	if !cryptotool.SignatureEqual("ABCdef", "abcDEF") || cryptotool.SignatureEqual("abc", "abd") || cryptotool.SignatureEqual("abc", "abcd") {
		t.Fatal("SignatureEqual 结果不正确")
	}
}
