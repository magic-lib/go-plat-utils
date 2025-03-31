package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

// The blockSize argument should be 16, 24, or 32.
// Corresponding AES-128, AES-192, or AES-256.
func pkcs7Padding(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plainText, paddingText...)
}

func pkcs7UnPadding(plainText []byte) []byte {
	length := len(plainText)
	unPadding := int(plainText[length-1])
	if length-unPadding < 0 { // 不能<0
		return []byte{}
	}
	return plainText[:length-unPadding]
}

func zeroPadding(plainText []byte, blockSize int) []byte {
	if len(plainText)%blockSize != 0 {
		paddingSize := blockSize - len(plainText)%blockSize
		paddingText := bytes.Repeat([]byte{byte(0)}, paddingSize)
		plainText = append(plainText, paddingText...)
	}
	return plainText
}

func unZeroPadding(plainText []byte) []byte {
	length := len(plainText)
	count := 1
	for i := length - 1; i > 0; i-- {
		if plainText[i] == 0 && plainText[i-1] == plainText[i] {
			count++
		}
	}
	return plainText[:length-count]
}

func aesEncrypt(origData, key []byte, iv string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs7Padding(origData, blockSize)

	var ivBytes []byte
	if iv == "" {
		ivBytes = key[:blockSize]
	} else {
		ivBytes = []byte(iv)
	}

	blockMode := cipher.NewCBCEncrypter(block, ivBytes)
	crypt := make([]byte, len(origData))
	blockMode.CryptBlocks(crypt, origData)
	return crypt, nil
}

func aesDecrypt(crypt, key []byte, iv string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	var ivBytes []byte
	if iv == "" {
		blockSize := block.BlockSize()
		ivBytes = key[:blockSize]
	} else {
		ivBytes = []byte(iv)
	}
	blockMode := cipher.NewCBCDecrypter(block, ivBytes)
	origData := make([]byte, len(crypt))

	blockMode.CryptBlocks(origData, crypt)
	origData = pkcs7UnPadding(origData)
	return origData, nil
}

func getAesKeyFromBase64(keyBase64 string) []byte {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		//表示传的不是base64格式的，则内部兼容处理一下
		key = []byte(keyBase64)
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		//太短或太长，则进行转化
		newKey := SHA(MD5, string(key))
		key = []byte(newKey)
	}
	return key
}
