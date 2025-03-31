package crypto

import (
	"encoding/hex"
	"fmt"
	"github.com/forgoer/openssl"
)

// ConfigEncryptSecret 对配置文件中密钥进行加密
func ConfigEncryptSecret(secret string, encryptionKey string) (string, error) {
	encryptedBytes, err := openssl.AesECBEncrypt([]byte(secret), []byte(encryptionKey), openssl.PKCS7_PADDING)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt value for key %s: %w", secret, err)
	}
	return hex.EncodeToString(encryptedBytes), nil
}

// ConfigDecryptSecret 对配置文件中密钥进行解密
func ConfigDecryptSecret(encryptSecret string, encryptionKey string) (string, error) {
	decodedValue, err := hex.DecodeString(encryptSecret)
	if err != nil {
		return encryptSecret, fmt.Errorf("failed to decode hex string: %w", err)
	}
	encryptedBytes, err := openssl.AesECBDecrypt(decodedValue, []byte(encryptionKey), openssl.PKCS7_PADDING)
	if err != nil {
		return encryptSecret, fmt.Errorf("failed to encrypt value: %w", err)
	}
	return string(encryptedBytes), nil
}
