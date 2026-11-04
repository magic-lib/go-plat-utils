package crypto_test

import (
	"encoding/base64"
	"testing"

	"github.com/magic-lib/go-plat-utils/crypto"
)

// 不合法的密文返回错误，不能 panic；合法密文正常解密
func TestAesDecryptInvalidInput(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("1234567890abcdef"))
	enc, err := crypto.AesEncrypt("hello", key)
	if err != nil {
		t.Fatalf("AesEncrypt: %v", err)
	}
	if plain, err := crypto.AesDecrypt(enc, key); err != nil || plain != "hello" {
		t.Fatalf("AesDecrypt = %q, %v", plain, err)
	}

	raw, _ := base64.StdEncoding.DecodeString(enc)
	tampered := append([]byte{}, raw...)
	tampered[len(tampered)-1] ^= 0xFF // 破坏最后一个块，填充大概率不合法
	for name, in := range map[string]string{
		"长度非块大小整数倍": base64.StdEncoding.EncodeToString(raw[:5]),
		"填充被篡改":     base64.StdEncoding.EncodeToString(tampered),
	} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: panic %v", name, r)
				}
			}()
			if plain, err := crypto.AesDecrypt(in, key); err == nil {
				t.Errorf("%s: 应返回错误, got %q", name, plain)
			}
		}()
	}
}
