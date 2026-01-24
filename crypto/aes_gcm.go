package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// stringToCipherBlock 将任意字符串转为 cipher.Block
// algorithm: 可选 "aes-128"、"aes-192"、"aes-256"
// keyStr: 任意字符串密钥
func stringToCipherBlock(algorithm, keyStr string) (cipher.Block, error) {
	hash := sha256.Sum256([]byte(keyStr))
	var key []byte

	switch algorithm {
	case "aes-128":
		key = hash[:16] // 128 位 = 16 字节 md5
	case "aes-192":
		key = hash[:24] // 192 位 = 24 字节
	case "aes-256":
		key = hash[:32] // 256 位 = 32 字节
	default:
		return nil, fmt.Errorf("不支持的算法: %s，仅支持 aes-128/aes-192/aes-256", algorithm)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("初始化 Block 失败: %v", err)
	}

	return block, nil
}

// stringTo12BytePadTrunc 填充/截断法：将字符串转为12字节byte数组（适合普通场景）
// 原理：长度>12则截断，长度<12则补0，长度=12则直接转换
func stringToBytePadTrunc(s string, copyLen int) []byte {
	if copyLen <= 0 {
		copyLen = 12
	}
	var result = make([]byte, copyLen, copyLen)
	// 将字符串转为字节切片
	sBytes := []byte(s)
	// 计算实际要复制的长度（取12和字符串字节长度的较小值）
	if len(sBytes) < copyLen {
		copyLen = len(sBytes)
	}
	// 复制字节到结果数组（不足部分自动为0）
	copy(result[:], sBytes[:copyLen])
	return result
}

func generateReadableByteNonce(bits int, charset string) ([]byte, error) {
	if bits <= 0 || len(charset) <= 1 {
		return nil, fmt.Errorf("参数错误")
	}

	charsetLen := byte(len(charset))

	// 生成12字节随机数据（用于逐个生成字符）
	bytes := make([]byte, bits)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return nil, err
	}

	nonce := make([]byte, bits)
	for i := 0; i < bits; i++ {
		nonce[i] = charset[bytes[i]%charsetLen]
	}
	return nonce, nil
}

const (
	charset  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	nonceLen = 12
)

// GCMEncrypt AES-GCM模式加密（带认证，更安全）
func GCMEncrypt(key string, plaintext []byte) (ciphertextStr string, nonceStr string, err error) {
	if key == "" || len(plaintext) == 0 {
		return "", "", fmt.Errorf("参数错误")
	}
	// 创建AES区块
	block, err := stringToCipherBlock("aes-256", key)
	if err != nil {
		return "", "", err
	}

	nonce, err := generateReadableByteNonce(nonceLen, charset)
	if err != nil {
		return "", "", err
	}

	// 创建GCM加密器
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), string(nonce), nil
}

// GCMDecrypt AES-GCM模式解密
func GCMDecrypt(key string, nonce string, ciphertext string) (plaintext []byte, err error) {
	block, err := stringToCipherBlock("aes-256", key)
	if err != nil {
		return nil, err
	}

	// 创建GCM解密器
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceByte := stringToBytePadTrunc(nonce, nonceLen)
	cipherByte, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("解密失败: %v", err)
	}

	// 解密（验证认证标签，防止数据篡改）
	plaintext, err = gcm.Open(nil, nonceByte, cipherByte, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
