package crypto

import (
	"testing"
)

func TestHybridEncryptKeyLenAndCompat(t *testing.T) {
	priKey, pubKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair: %v", err)
	}
	plain := []byte("hybrid plaintext")

	// 新格式：加解密往返，且 RSA 加密的随机 AES 密钥为 hybridAesKeyLen 个字符
	enc, err := HybridEncrypt(pubKey, plain)
	if err != nil {
		t.Fatalf("HybridEncrypt: %v", err)
	}
	if got, err := HybridDecrypt(priKey, enc); err != nil || string(got) != string(plain) {
		t.Fatalf("HybridDecrypt = %q, %v", got, err)
	}
	_, rsaKeyB64, _, err := decodeJwtStyle(enc)
	if err != nil {
		t.Fatalf("decodeJwtStyle: %v", err)
	}
	aesKey, err := OAEPDecrypt(priKey, rsaKeyB64)
	if err != nil || len(aesKey) != hybridAesKeyLen {
		t.Fatalf("AES 密钥长度 = %d, err=%v; want %d", len(aesKey), err, hybridAesKeyLen)
	}

	// 兼容旧格式：8 位密钥生成的密文仍能解密
	oldKey, _ := generateReadableByteNonce(8, charset)
	dataB64, nonce, err := GCMEncrypt(string(oldKey), plain)
	if err != nil {
		t.Fatalf("GCMEncrypt: %v", err)
	}
	oldRsaKey, err := OAEPEncrypt(pubKey, oldKey)
	if err != nil {
		t.Fatalf("OAEPEncrypt: %v", err)
	}
	oldEnc, err := createJwtStyle(dataB64, oldRsaKey, nonce)
	if err != nil {
		t.Fatalf("createJwtStyle: %v", err)
	}
	if got, err := HybridDecrypt(priKey, oldEnc); err != nil || string(got) != string(plain) {
		t.Fatalf("旧格式 HybridDecrypt = %q, %v", got, err)
	}
}
