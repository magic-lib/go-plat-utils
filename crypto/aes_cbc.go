package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/marspere/goencrypt"
	"io"
)

func getAllKeyString(key string) string {
	keyLen := len(key)
	if keyLen >= 32 { //大于32的
		return key[0:32]
	}
	if keyLen >= 24 {
		return key[0:24]
	}
	if keyLen >= 16 {
		return key[0:16]
	}
	if keyLen > 0 && keyLen < 16 {
		for i := keyLen; i < 16; i++ {
			key += " " //不足的后面增加空格字符补齐
		}
		return key
	}
	return "abcdefghijklmnop" //默认的key
}

// CbcEncrypt 加密
// encryptOri encrypt plain text by key
func CbcEncrypt(plainStr string, key string, en ...EnDeCoder) (string, error) {
	var ciphertext []byte
	var err error

	key = getAllKeyString(key)

	keyByte := []byte(key)

	goroutines.GoSync(func(params ...any) {
		var block cipher.Block
		block, err = aes.NewCipher(keyByte)
		if err != nil {
			return
		}

		plaintext := []byte(plainStr)
		if len(plaintext)%aes.BlockSize != 0 {
			plaintext = append(plaintext, bytes.Repeat([]byte{0}, aes.BlockSize-len(plaintext)%aes.BlockSize)...)
		}
		// The IV needs to be unique, but not secure. Therefore it's common to
		// include it at the beginning of the ciphertext.
		ciphertext = make([]byte, aes.BlockSize+len(plaintext))
		iv := ciphertext[:aes.BlockSize]
		if _, err = io.ReadFull(rand.Reader, iv); err != nil {
			return
		}

		err = nil

		cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	})

	if err != nil {
		return "", err
	}

	var enCode EnDeCoder = new(HexCoder)
	if en != nil && len(en) > 0 {
		enCode = en[0]
	}

	return enCode.Encode(ciphertext), nil
}

// CbcDecrypt 解密
// decryptOri decrypt cypher text by key
func CbcDecrypt(cipherStr string, key string, de ...EnDeCoder) (string, error) {
	var ciphertext []byte
	var err error

	key = getAllKeyString(key)
	keyByte := []byte(key)

	goroutines.GoSync(func(params ...any) {
		var block cipher.Block

		var deCode EnDeCoder = new(HexCoder)
		if de != nil && len(de) > 0 {
			deCode = de[0]
		}

		ciphertext, err = deCode.Decode(cipherStr)
		if err != nil {
			return
		}

		block, err = aes.NewCipher(keyByte)
		if err != nil {
			return
		}
		// The IV needs to be unique, but not secure. Therefore it's common to
		// include it at the beginning of the ciphertext.
		if len(ciphertext) < aes.BlockSize {
			err = fmt.Errorf("ciphertext too short")
			return
		}
		iv := ciphertext[:aes.BlockSize]
		ciphertext = ciphertext[aes.BlockSize:]

		stream := cipher.NewCBCDecrypter(block, iv)

		// XORKeyStream can work in-place if the two arguments are the same.
		stream.CryptBlocks(ciphertext, ciphertext)
	})

	if err != nil {
		return "", err
	}
	return string(bytes.TrimRight(ciphertext, string([]byte{0}))), nil
}

// AesCbcEncrypt 加密
// decryptOri decrypt cypher text by key
func AesCbcEncrypt(plainStr string, key string, en ...EnDeCoder) (string, error) {
	key = getAllKeyString(key)

	keyByte := []byte(key)

	cipher, err := goencrypt.NewAESCipher(keyByte, keyByte, goencrypt.CBCMode, goencrypt.PkcsZero, goencrypt.PrintHex)
	if err != nil {
		return "", err
	}
	cipherText, err := cipher.AESEncrypt([]byte(plainStr))
	if err != nil {
		return "", err
	}
	originText, err := hex.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	var enCode EnDeCoder = new(HexCoder)
	if en != nil && len(en) > 0 {
		enCode = en[0]
	}
	return enCode.Encode(originText), nil
}

// AesCbcDecrypt 解密
// decryptOri decrypt cypher text by key
func AesCbcDecrypt(cipherStr string, key string, de ...EnDeCoder) (string, error) {
	key = getAllKeyString(key)

	keyByte := []byte(key)

	var deCode EnDeCoder = new(HexCoder)
	if de != nil && len(de) > 0 {
		deCode = de[0]
	}
	cipherTextTemp, err := deCode.Decode(cipherStr)
	if err != nil {
		return "", err
	}

	cipher, err := goencrypt.NewAESCipher(keyByte, keyByte, goencrypt.CBCMode, goencrypt.PkcsZero, goencrypt.PrintHex)
	if err != nil {
		return "", err
	}
	originStr := hex.EncodeToString(cipherTextTemp)
	return cipher.AESDecrypt(originStr)
}
