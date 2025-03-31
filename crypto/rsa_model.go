package crypto

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
)

type RSAPrivateKeyType int

const (
	PKCS1Type RSAPrivateKeyType = iota
	PKCS8Type

	rSAKeyBit512  = 512
	rSAKeyBit1024 = 1024
	rSAKeyBit2048 = 2048
	rSAKeyBit4096 = 4096
)

type RSASecurity struct {
	pubStr string          //公钥字符串
	priStr string          //私钥字符串
	pubKey *rsa.PublicKey  //公钥
	priKey *rsa.PrivateKey //私钥
}

func getBits(bits int) int {
	if bits < 0 {
		bits = rSAKeyBit512
	} else if bits < 1024 {
		bits = rSAKeyBit1024
	} else if bits < 2048 {
		bits = rSAKeyBit2048
	} else if bits < 4096 {
		bits = rSAKeyBit4096
	} else {
		bits = rSAKeyBit4096
	}
	return bits
}

// CreateRSAKeys 生成RSA私钥和公钥
func (r *RSASecurity) CreateRSAKeys(bits int, keyType RSAPrivateKeyType) (publicKeyString string, privateKeyString string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, getBits(bits))
	if err != nil {
		return "", "", err
	}

	// 将私钥转换为PEM格式
	var privateKeyBlock *pem.Block
	if keyType == PKCS1Type {
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyBlock = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		}
	} else {
		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return "", "", err
		}
		privateKeyBlock = &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateKeyBytes,
		}
	}

	// 将公钥转换为PEM格式
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	privateKeyString = string(pem.EncodeToMemory(privateKeyBlock))
	publicKeyString = string(pem.EncodeToMemory(publicKeyBlock))

	if r.pubStr == "" || r.priStr == "" {
		r.SetPublicAndPrivateKey(publicKeyString, privateKeyString)
	}

	return publicKeyString, privateKeyString, nil
}

// SetPublicAndPrivateKey 设置公钥私钥
func (r *RSASecurity) SetPublicAndPrivateKey(pubStr string, priStr string) (err error) {
	if pubStr != "" {
		r.pubStr = pubStr
		r.pubKey, err = r.getPublicKey()
		if err != nil {
			return err
		}
	}
	if priStr != "" {
		r.priStr = priStr
		r.priKey, err = r.getPrivateKey()
		if err != nil {
			return err
		}
	}
	return nil
}

// getPublicKey *rsa.PrivateKey
func (r *RSASecurity) getPublicKey() (*rsa.PublicKey, error) {
	return getPubKey([]byte(r.pubStr))
}

// getPrivateKey *rsa.PublicKey
func (r *RSASecurity) getPrivateKey() (*rsa.PrivateKey, error) {
	return getPriKey([]byte(r.priStr))
}

// PubKeyEncryptByte 公钥加密
func (r *RSASecurity) PubKeyEncryptByte(input []byte) ([]byte, error) {
	if r.pubKey == nil {
		return []byte(""), fmt.Errorf(`Please set the public key in advance`)
	}
	output := bytes.NewBuffer(nil)
	err := pubKeyIO(r.pubKey, bytes.NewReader(input), output, true)
	if err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}

// PriKeyDecryptByte 私钥解密
func (r *RSASecurity) PriKeyDecryptByte(input []byte) ([]byte, error) {
	if r.priKey == nil {
		return []byte(""), fmt.Errorf(`Please set the private key in advance`)
	}
	output := bytes.NewBuffer(nil)
	err := priKeyIO(r.priKey, bytes.NewReader(input), output, false)
	if err != nil {
		return []byte(""), err
	}

	return io.ReadAll(output)
}

// PubKeyEncrypt 公钥加密
func (r *RSASecurity) PubKeyEncrypt(input string, en ...EnDeCoder) (string, error) {
	oldByte, err := r.PubKeyEncryptByte([]byte(input))
	if err != nil {
		return "", err
	}

	var enCode EnDeCoder = new(Base64Coder)
	if en != nil && len(en) > 0 {
		enCode = en[0]
	}
	return enCode.Encode(oldByte), nil
}

// PriKeyDecrypt 私钥解密
func (r *RSASecurity) PriKeyDecrypt(encodeBase64String string, de ...EnDeCoder) (string, error) {
	var deCode EnDeCoder = new(Base64Coder)
	if de != nil && len(de) > 0 {
		deCode = de[0]
	}

	newByte, err := deCode.Decode(encodeBase64String)
	if err != nil {
		return "", err
	}
	newStr, err := r.PriKeyDecryptByte(newByte)
	if err != nil {
		return "", err
	}
	return string(newStr), nil
}

// PriKeyEncryptByte 私钥加密，用于数字签名
func (r *RSASecurity) PriKeyEncryptByte(input []byte) ([]byte, error) {
	if r.priKey == nil {
		return []byte(""), fmt.Errorf(`Please set the private key in advance`)
	}
	output := bytes.NewBuffer(nil)
	err := priKeyIO(r.priKey, bytes.NewReader(input), output, true)
	if err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}

// PubKeyDecryptByte 公钥解密，用于数字签名
func (r *RSASecurity) PubKeyDecryptByte(input []byte) ([]byte, error) {
	if r.pubKey == nil {
		return []byte(""), fmt.Errorf(`Please set the public key in advance`)
	}
	output := bytes.NewBuffer(nil)
	err := pubKeyIO(r.pubKey, bytes.NewReader(input), output, false)
	if err != nil {
		return []byte(""), err
	}
	return io.ReadAll(output)
}

// PriKeyEncrypt 私钥加密
func (r *RSASecurity) PriKeyEncrypt(input string, en ...EnDeCoder) (string, error) {
	oldByte, err := r.PriKeyEncryptByte([]byte(input))
	if err != nil {
		return "", err
	}

	var enCode EnDeCoder = new(Base64Coder)
	if en != nil && len(en) > 0 {
		enCode = en[0]
	}
	return enCode.Encode(oldByte), nil
}

// PubKeyDecrypt 公钥解密
func (r *RSASecurity) PubKeyDecrypt(encodeBase64String string, de ...EnDeCoder) (string, error) {
	var deCode EnDeCoder = new(Base64Coder)
	if de != nil && len(de) > 0 {
		deCode = de[0]
	}
	newByte, err := deCode.Decode(encodeBase64String)
	if err != nil {
		return "", err
	}
	newStr, err := r.PubKeyDecryptByte(newByte)
	if err != nil {
		return "", err
	}
	return string(newStr), nil
}

// SignMd5 使用RSAWithMD5算法签名
func (r *RSASecurity) SignMd5(data string, en ...EnDeCoder) (string, error) {
	md5Hash := md5.New()
	sData := []byte(data)
	md5Hash.Write(sData)
	hashed := md5Hash.Sum(nil)

	signByte, err := rsa.SignPKCS1v15(rand.Reader, r.priKey, crypto.MD5, hashed)

	var enCode EnDeCoder = new(Base64Coder)
	if en != nil && len(en) > 0 {
		enCode = en[0]
	}

	return enCode.Encode(signByte), err
}

// SignSha1 使用RSAWithSHA1算法签名
func (r *RSASecurity) SignSha1(data string, en ...EnDeCoder) (string, error) {
	sha1Hash := sha1.New()
	sData := []byte(data)
	sha1Hash.Write(sData)
	hashed := sha1Hash.Sum(nil)

	signByte, err := rsa.SignPKCS1v15(rand.Reader, r.priKey, crypto.SHA1, hashed)

	var enCode EnDeCoder = new(Base64Coder)
	if en != nil && len(en) > 0 {
		enCode = en[0]
	}

	return enCode.Encode(signByte), err
}

// SignSha256 使用RSAWithSHA256算法签名
func (r *RSASecurity) SignSha256(data string, en ...EnDeCoder) (string, error) {
	sha256Hash := sha256.New()
	sData := []byte(data)
	sha256Hash.Write(sData)
	hashed := sha256Hash.Sum(nil)

	signByte, err := rsa.SignPKCS1v15(rand.Reader, r.priKey, crypto.SHA256, hashed)

	var enCode EnDeCoder = new(Base64Coder)
	if en != nil && len(en) > 0 {
		enCode = en[0]
	}
	return enCode.Encode(signByte), err
}

// VerifySignMd5 使用RSAWithMD5验证签名
func (r *RSASecurity) VerifySignMd5(data string, signData string, de ...EnDeCoder) error {
	var deCode EnDeCoder = new(Base64Coder)
	if de != nil && len(de) > 0 {
		deCode = de[0]
	}

	sign, err := deCode.Decode(signData)
	if err != nil {
		return err
	}
	hash := md5.New()
	hash.Write([]byte(data))
	return rsa.VerifyPKCS1v15(r.pubKey, crypto.MD5, hash.Sum(nil), sign)
}

// VerifySignSha1 使用RSAWithSHA1验证签名
func (r *RSASecurity) VerifySignSha1(data string, signData string, de ...EnDeCoder) error {
	var deCode EnDeCoder = new(Base64Coder)
	if de != nil && len(de) > 0 {
		deCode = de[0]
	}

	sign, err := deCode.Decode(signData)
	if err != nil {
		return err
	}
	hash := sha1.New()
	hash.Write([]byte(data))
	return rsa.VerifyPKCS1v15(r.pubKey, crypto.SHA1, hash.Sum(nil), sign)
}

// VerifySignSha256 使用RSAWithSHA256验证签名
func (r *RSASecurity) VerifySignSha256(data string, signData string, de ...EnDeCoder) error {
	var deCode EnDeCoder = new(Base64Coder)
	if de != nil && len(de) > 0 {
		deCode = de[0]
	}

	sign, err := deCode.Decode(signData)
	if err != nil {
		return err
	}
	hash := sha256.New()
	hash.Write([]byte(data))

	return rsa.VerifyPKCS1v15(r.pubKey, crypto.SHA256, hash.Sum(nil), sign)
}
