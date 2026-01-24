package crypto

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"net/url"
	"sort"
	"strings"
)

type SignParams struct {
	Method    string     `json:"method"`
	Path      string     `json:"path"`
	Query     url.Values `json:"query"`
	Body      []byte     `json:"body"`
	Timestamp int64      `json:"timestamp"` // 时间戳，秒
	Nonce     string     `json:"nonce"`
}

// 计算 body 的 SHA256 十六进制
func hashBody(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

// 构造排序后的 querystring
func buildSortedQuery(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	first := true
	for _, k := range keys {
		for _, v := range values[k] {
			if !first {
				buf.WriteByte('&')
			}
			first = false
			// 这里简单拼 key=value，必要时可以再 url encode 一下
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	return buf.String()
}
func (p *SignParams) String() string {
	bodyHash := hashBody(p.Body)
	queryStr := buildSortedQuery(p.Query)

	// 构造签名原串
	var bufList = make([]string, 0)
	if p.Method != "" {
		bufList = append(bufList, p.Method)
	}
	if p.Path != "" {
		bufList = append(bufList, p.Path)
	}
	if queryStr != "" {
		bufList = append(bufList, queryStr)
	}
	if bodyHash != "" {
		bufList = append(bufList, bodyHash)
	}
	if p.Timestamp != 0 {
		bufList = append(bufList, conv.String(p.Timestamp))
	}
	if p.Nonce != "" {
		bufList = append(bufList, p.Nonce)
	}

	return strings.Join(bufList, "\n")
}

func SignatureParamsByHmac(secret string, p *SignParams) (string, string, error) {
	if p == nil || secret == "" {
		return "", "", fmt.Errorf("secret is empty")
	}
	signStr := p.String()
	signature, err := SignatureByHmac(secret, signStr)
	if err != nil {
		return "", "", err
	}
	return signature, signStr, nil
}

// VerifySignatureParamsByHmac 验证签名是否合法
func VerifySignatureParamsByHmac(secret string, p *SignParams, signature string) (bool, error) {
	if p == nil || secret == "" {
		return false, fmt.Errorf("secret is empty")
	}
	signStr := p.String()
	return VerifySignatureByHmac(secret, signStr, signature)
}

func SignatureParamsByRsa(privateKeyStr string, p *SignParams) (string, string, error) {
	if p == nil || privateKeyStr == "" {
		return "", "", fmt.Errorf("secret is empty")
	}
	signStr := p.String()

	privateKey, err := ParsePemPrivateKey(privateKeyStr)
	if err != nil {
		return "", "", err
	}
	signature, err := SignatureByRsa(privateKey, signStr)
	if err != nil {
		return "", "", err
	}
	return signature, signStr, nil
}

func VerifySignatureParamsByRsa(publicKeyStr string, p *SignParams, signature string) (bool, error) {
	if publicKeyStr == "" || p == nil || signature == "" {
		return false, fmt.Errorf("公钥、签名材料、签名值均不能为空")
	}
	signStr := p.String()
	publicKey, err := ParsePemPublicKey(publicKeyStr)
	if err != nil {
		return false, err
	}
	return VerifySignatureByRsa(publicKey, signStr, signature)
}

func SignatureByHmac(secret string, signStr string) (string, error) {
	if signStr == "" || secret == "" {
		return "", fmt.Errorf("secret is empty")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signStr))
	sum := mac.Sum(nil)

	return hex.EncodeToString(sum), nil
}
func VerifySignatureByHmac(secret string, singStr, signature string) (bool, error) {
	genSign, err := SignatureByHmac(secret, singStr)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(genSign, signature), nil
}
func SignatureByRsa11(privateKeyStr string, signStr string) (string, error) {
	if signStr == "" || privateKeyStr == "" {
		return "", fmt.Errorf("secret is empty")
	}
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return "", fmt.Errorf("私钥解析失败")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("非RSA私钥")
	}

	hash := sha256.New()
	hash.Write([]byte(signStr))
	hashBytes := hash.Sum(nil)

	signBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signBytes), nil
}
func SignatureByRsa(rsaPrivateKey *rsa.PrivateKey, signStr string) (string, error) {
	if signStr == "" || rsaPrivateKey == nil {
		return "", fmt.Errorf("secret is empty")
	}

	hash := sha256.New()
	hash.Write([]byte(signStr))
	hashBytes := hash.Sum(nil)

	signBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signBytes), nil
}

func VerifySignatureByRsa111(publicKeyStr string, signStr string, signature string) (bool, error) {
	if publicKeyStr == "" || signStr == "" || signature == "" {
		return false, fmt.Errorf("公钥、签名材料、签名值均不能为空")
	}
	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil || block.Type != "PUBLIC KEY" {
		return false, fmt.Errorf("公钥格式错误（非PEM格式）")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("解析公钥失败：%w", err)
	}
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("公钥非RSA类型")
	}

	signBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("签名值Base64解码失败：%w", err)
	}

	hash := sha256.New()
	hash.Write([]byte(signStr))
	hashBytes := hash.Sum(nil)

	err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashBytes, signBytes)
	if err != nil {
		return false, fmt.Errorf("验签失败：%w", err)
	}
	return true, nil
}
func VerifySignatureByRsa(rsaPubKey *rsa.PublicKey, signStr string, signature string) (bool, error) {
	if rsaPubKey == nil || signStr == "" || signature == "" {
		return false, fmt.Errorf("公钥、签名材料、签名值均不能为空")
	}

	signBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("签名值Base64解码失败：%w", err)
	}

	hash := sha256.New()
	hash.Write([]byte(signStr))
	hashBytes := hash.Sum(nil)

	err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashBytes, signBytes)
	if err != nil {
		return false, fmt.Errorf("验签失败：%w", err)
	}
	return true, nil
}
