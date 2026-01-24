package crypto

import (
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/magic-lib/go-plat-utils/conv"
)

// HybridEncrypt 混合加密：AES加密数据 + RSA加密AES密钥
// 返回格式：base64(rsa加密的aes密钥) + ":" + base64(aes加密的数据)
func HybridEncrypt(rsaPubKey *rsa.PublicKey, plaintext []byte) (string, error) {
	aesKey, err := generateReadableByteNonce(8, charset)
	if err != nil {
		return "", fmt.Errorf("生成AES密钥失败：%v", err)
	}

	// 2. 用AES-GCM加密原始数据（对称加密，速度极快）
	dataB64, nonce, err := GCMEncrypt(string(aesKey), plaintext)
	if err != nil {
		return "", fmt.Errorf("AES加密数据失败：%v", err)
	}

	// 3. 用RSA-OAEP加密AES密钥
	rsaKeyB64, err := OAEPEncrypt(rsaPubKey, aesKey)
	if err != nil {
		return "", fmt.Errorf("RSA加密AES密钥失败：%v", err)
	}
	return createJwtStyle(dataB64, rsaKeyB64, nonce)
}

// HybridDecrypt 混合解密：RSA解密AES密钥 + AES解密数据
func HybridDecrypt(rsaPrivKey *rsa.PrivateKey, encryptedStr string) ([]byte, error) {
	dataB64, rsaKeyB64, nonce, err := decodeJwtStyle(encryptedStr)
	if err != nil {
		return nil, fmt.Errorf("解析HybridDecrypt失败: %v", err)
	}

	aesKey, err := OAEPDecrypt(rsaPrivKey, rsaKeyB64)
	if err != nil {
		return nil, fmt.Errorf("OAEPDecrypt解密AES密钥失败：%v", err)
	}

	plaintext, err := GCMDecrypt(string(aesKey), nonce, dataB64)
	if err != nil {
		return nil, fmt.Errorf("AES解密数据失败：%v", err)
	}

	return plaintext, nil
}

type hybridClaims struct {
	Data  string `json:"data"`
	Key   string `json:"key"`
	Nonce string `json:"nonce"`
}

func createJwtStyle(dataB64, rsaKeyB64, nonce string) (string, error) {
	mapData := hybridClaims{
		Data:  dataB64,
		Key:   rsaKeyB64,
		Nonce: nonce,
	}
	mapClaims := make(jwt.MapClaims)
	err := conv.Unmarshal(mapData, &mapClaims)
	if err != nil {
		return "", fmt.Errorf("转换数据失败: %v", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	jwtString, err := token.SignedString([]byte(nonce))
	if err != nil {
		return "", fmt.Errorf("生成JWT失败: %v", err)
	}
	return jwtString, nil
}
func decodeJwtStyle(encryptedStr string) (dataB64, rsaKeyB64, nonce string, err error) {
	text, err := jwtPayload(encryptedStr)
	if err != nil {
		return "", "", "", err
	}
	dataClaims := new(hybridClaims)
	err = conv.Unmarshal(text, dataClaims)
	if err != nil {
		return "", "", "", fmt.Errorf("解析JWT失败: %v", err)
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(dataClaims.Nonce), nil
	}
	_, err = jwt.ParseWithClaims(
		encryptedStr,
		&jwt.MapClaims{},
		keyFunc,
	)
	if err != nil {
		return "", "", "", fmt.Errorf("解析令牌失败: %v", err)
	}
	return dataClaims.Data, dataClaims.Key, dataClaims.Nonce, nil
}
