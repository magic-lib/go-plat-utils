package httputil_test

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/utils/httputil"
)

func TestSign(t *testing.T) {

}

// TestGenFeiShuSign 验证飞书签名：sign = base64(hmac-sha256(timestamp+"\n"+secret, secret))
func TestGenFeiShuSign(t *testing.T) {
	secret := "test-secret"
	ts := time.Now().Unix()
	m := httputil.GenFeiShuSign(ts, secret)

	fmt.Println(m)

	// 独立复现签名算法
	stringToSign := fmt.Sprintf("%d\n%s", ts, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if m != want {
		t.Errorf("sign = %q, want %q", m, want)
	}
}

// TestGenWxPayV2Sign 验证微信支付 V2 签名：key 字典序拼接 + &key=apiKey 后 MD5 十六进制
func TestGenWxPayV2Sign(t *testing.T) {
	params := map[string]string{
		"appid":     "wx1234567890abcdef",
		"mch_id":    "1234567890",
		"nonce_str": "abc123",
		"body":      "测试商品",
		"sign":      "should-be-ignored", // 会被剔除
		"empty":     "",                  // 空值不参与
	}
	apiKey := "0123456789abcdef0123456789abcdef"

	got := httputil.GenWxPayV2Sign(params, apiKey)

	fmt.Println(got)

	// 独立复现签名算法
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var str string
	for _, k := range keys {
		if v := params[k]; v != "" {
			str += fmt.Sprintf("%s=%s&", k, v)
		}
	}
	str = str + "key=" + apiKey
	h := md5.Sum([]byte(str))
	// 微信 V2 签名为大写十六进制
	want := strings.ToUpper(hex.EncodeToString(h[:]))

	if got != want {
		t.Errorf("GenWxPayV2Sign = %q, want %q (stringToSign=%q)", got, want, str)
	}
	if len(got) != 32 {
		t.Errorf("签名长度 = %d, want 32（MD5 十六进制）", len(got))
	}
}

// TestGenWxPayV2SignStable 验证相同入参结果稳定，且剔除 sign/空值字段不影响顺序
func TestGenWxPayV2SignStable(t *testing.T) {
	params := map[string]string{
		"b": "2",
		"a": "1",
		"c": "3",
	}
	apiKey := "key123"
	first := httputil.GenWxPayV2Sign(params, apiKey)
	second := httputil.GenWxPayV2Sign(params, apiKey)
	if first != second {
		t.Errorf("相同入参签名不一致: %q vs %q", first, second)
	}
}
