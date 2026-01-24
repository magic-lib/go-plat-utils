package crypto_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/crypto"
	"net/http"
	"testing"
)

func TestSignHmac(t *testing.T) {
	secret := "$apr1$CppP/Mm.$SSDya7Irg8zZOhSJHqAfc/"
	enCode := "572414e361bbb61862106b398fa664eeda5f3993bdd113887ae063803db24594"

	pData := crypto.SignParams{
		Method: http.MethodPost,
	}
	aa, bb, err := crypto.SignatureParamsByHmac(secret, &pData)

	fmt.Println(aa, bb, err)

	ret, err := crypto.VerifySignatureParamsByHmac(secret, &pData, enCode)
	fmt.Println(ret, err)
}

var pubKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC6866oJusYb8I/NuFBIZnvls0w
VVQZqg2shem1JCOVD59HVh7NarQdzEImR3lq4BtNLv7Ggh4T7ap0Z4YvGD6Qcz6m
jLHo1ilMJOwc3B6rUL3N+b7hBNQHd82FX8MnzlVzU/NfeEGMg11LfYMQxNTP0q+T
kU9OLx7u9zVUScBIJwIDAQAB
-----END PUBLIC KEY-----`
var privateKey = `-----BEGIN PRIVATE KEY-----
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBALrzrqgm6xhvwj82
4UEhme+WzTBVVBmqDayF6bUkI5UPn0dWHs1qtB3MQiZHeWrgG00u/saCHhPtqnRn
hi8YPpBzPqaMsejWKUwk7BzcHqtQvc35vuEE1Ad3zYVfwyfOVXNT8194QYyDXUt9
gxDE1M/Sr5ORT04vHu73NVRJwEgnAgMBAAECgYATm9+j74EVLRO4wa7awAV/Zdfb
y/doQbfxcpJS15mL1vmj59qPPTPrNDN0BGct2JfEfrtmYtt4x6LrVrhyVB6rpiOr
4RCnR617QPnUSHteBRqnjOy+rGVhGBuDvsgfVyTyOQeOG8bDahwDdBq4SZpryo4f
aT0UPhb5FZgsTRUb8QJBAN9xUNs8np6x9TmewGCvB6uMYHhVXgtMHYtX7WfhTdmj
2qAjZ3fakb8uGydUFMlmNSp957hZmLE50plQi2oKq6MCQQDWMTds6dgFYoYf6/bg
AH23MWdDupCQD2sY50LSkiZBGmsejjtcDx58Z0qM7Chv0s6cNJPQJwlRFceCRPJj
kTmtAkAzhsUXmZYWkIE1ZWeFpDdHlxqUBVOnlUjm3kLwBqPWQZPkA+YTXILprG80
lY4pl3lBMEGkYHz2uZfYJRvRO16zAkBcj341KcS5Zv8xEkZoPK4XGVlXsmrAZnlQ
lLeSyaeQYLtDxBEw0jPJbNWRmohK8p1ocWwi+ouTJ8dEq0jX8C0tAkEAwwXVaFaO
yik/1rcqlO+yrj1U0bArkK8g3UGPHNgcppuS6zK13WZu+hRvlLkCHulNAdn4seM3
StqabUv6lS/ArA==
-----END PRIVATE KEY-----`

func TestSignRsa(t *testing.T) {

	signature := `JG9/0Vwwvwdfx2fL/Js0Zla8evFS/+vIX7pWxe/dq2BA6a5RcZEVDIHv4qAydPCoenXlrsQJFjmVsnHDQMRdNoTz4Hk5ILFBMuxnfvxPg+Je0fO3KWNH2jjLXGmaE0dN3ySSJSEmZYICY1AQeRYvN7g1pBgOTrOvF/lerTmOTAQ=`

	pData := crypto.SignParams{
		Method: http.MethodPost,
	}
	aa, bb, err := crypto.SignatureParamsByRsa(privateKey, &pData)

	fmt.Println(aa, bb, err)

	ret, err := crypto.VerifySignatureParamsByRsa(pubKey, &pData, signature)
	fmt.Println(ret, err)
}

func TestEncryptOAEP(t *testing.T) {
	signature := `hello`
	pubKey, _ := crypto.ParsePemPublicKey(pubKey)
	priKey1, _ := crypto.ParsePemPrivateKey(privateKey)
	encodeStr, err := crypto.OAEPEncrypt(pubKey, []byte(signature))
	fmt.Println(encodeStr, err)
	decodeStr, err := crypto.OAEPDecrypt(priKey1, encodeStr)
	fmt.Println(string(decodeStr), err)
}

func TestGenerateRsa(t *testing.T) {
	priKey1, pubKey, err := crypto.GenerateRSAKeyPair(0)
	//fmt.Println(priKey1, pubKey, err)
	priStr, _ := crypto.RsaPrivateKeyToPEM(priKey1, "")
	pubStr, _ := crypto.RsaPublicKeyToPEM(pubKey)
	fmt.Println(pubStr)
	fmt.Println(priStr)

	pubKey2, _ := crypto.ParsePemPublicKey(pubStr)
	priKey2, _ := crypto.ParsePemPrivateKey(priStr)

	signature := `world`
	encodeStr, err := crypto.OAEPEncrypt(pubKey2, []byte(signature))
	fmt.Println(encodeStr, err)
	decodeStr, err := crypto.OAEPDecrypt(priKey2, encodeStr)
	fmt.Println(string(decodeStr), err)
}

func TestHybrid(t *testing.T) {
	priKey1, pubKey, err := crypto.GenerateRSAKeyPair(0)
	//fmt.Println(priKey1, pubKey, err)
	priStr, _ := crypto.RsaPrivateKeyToPEM(priKey1, "")
	pubStr, _ := crypto.RsaPublicKeyToPEM(pubKey)

	pubKey2, _ := crypto.ParsePemPublicKey(pubStr)
	priKey2, _ := crypto.ParsePemPrivateKey(priStr)

	signature := `world`
	encodeStr, err := crypto.HybridEncrypt(pubKey2, []byte(signature))
	fmt.Println(encodeStr, err)
	decodeStr, err := crypto.HybridDecrypt(priKey2, encodeStr)
	fmt.Println(string(decodeStr), err)
}
