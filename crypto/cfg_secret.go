package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
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

// GenerateRSAKeyPair 生成RSA密钥对（2048位）
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if bits <= 2048 {
		bits = 2048
	}
	// 生成2048位RSA私钥（生产环境推荐4096位）
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("生成私钥失败: %v", err)
	}
	// 从私钥提取公钥
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

func RsaPublicKeyToPEM(pubKey *rsa.PublicKey) (string, error) {
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("MarshalPKIXPublicKey失败：%v", err)
	}

	// 2. 封装为PEM块（标准PUBLIC KEY格式）
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY", // 公钥标准PEM类型
		Bytes: derBytes,
	})

	if pemBytes == nil {
		return "", fmt.Errorf("PEM编码公钥失败")
	}
	return string(pemBytes), nil
}

func RsaPrivateKeyToPEM(privKey *rsa.PrivateKey, format string) (string, error) {
	var pemBytes []byte
	if format == "" {
		format = "pkcs8"
	}

	switch format {
	case "pkcs8":
		// PKCS#8格式（通用，BEGIN PRIVATE KEY）
		derBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
		if err != nil {
			return "", fmt.Errorf("MarshalPKCS8PrivateKey失败：%v", err)
		}
		pemBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: derBytes,
		})
	case "pkcs1":
		// PKCS#1格式（传统RSA，BEGIN RSA PRIVATE KEY）
		derBytes := x509.MarshalPKCS1PrivateKey(privKey)
		pemBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: derBytes,
		})
	default:
		return "", fmt.Errorf("不支持的私钥格式：%s，仅支持 pkcs8/pkcs1", format)
	}

	if pemBytes == nil {
		return "", fmt.Errorf("PEM编码私钥失败")
	}

	return string(pemBytes), nil
}

// ParsePemPublicKey 将PEM格式公钥字符串解析为rsa.PublicKey
func ParsePemPublicKey(pemPublicKey string) (*rsa.PublicKey, error) {
	// 1. 解码PEM块
	block, _ := pem.Decode([]byte(pemPublicKey))
	if block == nil {
		return nil, fmt.Errorf("解析PEM公钥失败：无效的PEM格式")
	}

	// 2. 检查PEM块类型
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("PEM块类型错误，期望 PUBLIC KEY，实际：%s", block.Type)
	}

	// 3. 解析公钥（PKIX格式）
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败：%v", err)
	}

	// 4. 类型断言转换为rsa.PublicKey
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("公钥类型错误，非RSA公钥")
	}

	return rsaPubKey, nil
}

// ParsePemPrivateKey 将PEM格式私钥字符串解析为rsa.PrivateKey
func ParsePemPrivateKey(pemPrivateKey string) (*rsa.PrivateKey, error) {
	// 1. 解码PEM块（从字符串转字节后解析）
	block, _ := pem.Decode([]byte(pemPrivateKey))
	if block == nil {
		return nil, fmt.Errorf("解析PEM私钥失败：无效的PEM格式（可能是字符串格式错误或缺失首尾标记）")
	}

	// 2. 检查PEM块类型（常见的私钥类型有两种：PRIVATE KEY / RSA PRIVATE KEY）
	if block.Type != "PRIVATE KEY" && block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("PEM块类型错误，期望 PRIVATE KEY 或 RSA PRIVATE KEY，实际：%s", block.Type)
	}

	// 3. 解析私钥
	var privKey *rsa.PrivateKey
	var err error

	if block.Type == "PRIVATE KEY" {
		// PKCS#8 格式私钥（通用格式，BEGIN PRIVATE KEY）
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析PKCS#8私钥失败：%v", err)
		}
		// 类型断言转换为RSA私钥
		var ok bool
		privKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("私钥类型错误，非RSA私钥")
		}
	} else {
		// PKCS#1 格式私钥（传统RSA格式，BEGIN RSA PRIVATE KEY）
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析PKCS#1私钥失败：%v", err)
		}
	}

	return privKey, nil
}
