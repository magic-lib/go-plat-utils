package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// RSAEncrypt RSA加密数据，key必须是成对出现
func RSAEncrypt(oneKeyStr string, message string) (string, error) {
	en := new(Base64Coder)

	rsaModel := new(RSASecurity)
	err := rsaModel.SetPublicAndPrivateKey(oneKeyStr, "")
	if err != nil {
		err = rsaModel.SetPublicAndPrivateKey("", oneKeyStr)
		if err != nil {
			return "", err
		}
		return rsaModel.PriKeyEncrypt(message, en)
	}
	return rsaModel.PubKeyEncrypt(message, en)
}

// RSADecrypt RSA解密数据，key必须是成对出现
func RSADecrypt(otherKeyStr string, cipherText string) (string, error) {
	de := new(Base64Coder)

	rsaModel := new(RSASecurity)
	err := rsaModel.SetPublicAndPrivateKey("", otherKeyStr)
	if err != nil {
		err = rsaModel.SetPublicAndPrivateKey(otherKeyStr, "")
		if err != nil {
			return "", err
		}
		return rsaModel.PubKeyDecrypt(cipherText, de)
	}
	return rsaModel.PriKeyDecrypt(cipherText, de)
}

// OAEPEncrypt 使用公钥+OAEP填充加密（对应EncryptOAEP）
func OAEPEncrypt(publicKey *rsa.PublicKey, plaintext []byte) (string, error) {
	// OAEP填充需要哈希函数（这里用SHA256）
	hash := sha256.New()

	// 执行OAEP加密
	ciphertext, err := rsa.EncryptOAEP(
		hash,        // 哈希函数
		rand.Reader, // 随机数生成器
		publicKey,   // 公钥
		plaintext,   // 明文
		nil,         // 可选的label，加密和解密必须一致，通常设为nil
	)
	if err != nil {
		return "", fmt.Errorf("加密失败: %v", err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// OAEPDecrypt 使用私钥+OAEP填充解密（对应DecryptOAEP）
func OAEPDecrypt(privateKey *rsa.PrivateKey, ciphertext string) ([]byte, error) {
	// 哈希函数必须和加密时一致
	hash := sha256.New()

	cipherByte, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("解密失败: %v", err)
	}

	// 执行OAEP解密
	plaintext, err := rsa.DecryptOAEP(
		hash,        // 哈希函数
		rand.Reader, // 随机数生成器
		privateKey,  // 私钥
		cipherByte,  // 密文
		nil,         // label，必须和加密时一致
	)
	if err != nil {
		return nil, fmt.Errorf("解密失败: %v", err)
	}
	return plaintext, nil
}
