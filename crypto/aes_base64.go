package crypto

import (
	"bytes"
	basecrypto "crypto"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// The blockSize argument should be 16, 24, or 32.
// Corresponding AES-128, AES-192, or AES-256.
func pkcs7Padding(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plainText, paddingText...)
}

// pkcs7UnPadding 去除 PKCS7 填充并校验其合法性：空输入（原来访问 [-1] panic）、填充长度越界或填充字节不一致时返回错误
func pkcs7UnPadding(plainText []byte, blockSize int) ([]byte, error) {
	length := len(plainText)
	if length == 0 {
		return nil, fmt.Errorf("pkcs7: empty data")
	}
	unPadding := int(plainText[length-1])
	if unPadding == 0 || unPadding > blockSize || unPadding > length {
		return nil, fmt.Errorf("pkcs7: invalid padding")
	}
	for _, b := range plainText[length-unPadding:] {
		if int(b) != unPadding {
			return nil, fmt.Errorf("pkcs7: invalid padding")
		}
	}
	return plainText[:length-unPadding], nil
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
	// 校验 IV 与密文长度：不合法时 NewCBCDecrypter / CryptBlocks 会直接 panic，不可信密文可使进程崩溃
	if len(ivBytes) != block.BlockSize() {
		return nil, fmt.Errorf("aes: iv length must be %d", block.BlockSize())
	}
	if len(crypt) == 0 || len(crypt)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("aes: ciphertext is not a multiple of the block size")
	}
	blockMode := cipher.NewCBCDecrypter(block, ivBytes)
	origData := make([]byte, len(crypt))

	blockMode.CryptBlocks(origData, crypt)
	return pkcs7UnPadding(origData, block.BlockSize())
}

func getAesKeyFromBase64(keyBase64 string) []byte {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		//表示传的不是base64格式的，则内部兼容处理一下
		key = []byte(keyBase64)
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		//太短或太长，则进行转化
		newKey := SHA(basecrypto.MD5, string(key))
		key = []byte(newKey)
	}
	return key
}
