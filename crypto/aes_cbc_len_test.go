package crypto_test

import (
	"encoding/hex"
	"testing"

	"github.com/magic-lib/go-plat-utils/crypto"
)

func TestCbcDecryptInvalidLength(t *testing.T) {
	enc, err := crypto.CbcEncrypt("hello world", "1234567890abcdef")
	if err != nil {
		t.Fatalf("CbcEncrypt err: %v", err)
	}
	if plain, err := crypto.CbcDecrypt(enc, "1234567890abcdef"); err != nil || plain != "hello world" {
		t.Fatalf("CbcDecrypt = %q, %v", plain, err)
	}

	// IV 之后的密文长度不是 16 的倍数
	raw, _ := hex.DecodeString(enc)
	bad := hex.EncodeToString(raw[:len(raw)-3])
	if plain, err := crypto.CbcDecrypt(bad, "1234567890abcdef"); err == nil {
		t.Fatalf("密文长度非法时应返回错误, got %q", plain)
	}
}
