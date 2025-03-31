package crypto_test

import (
	"fmt"
	jwtA "github.com/golang-jwt/jwt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/crypto"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestAesCbc(t *testing.T) {
	key := "tianlin020250214"
	testCases := []*utils.TestStruct{
		{"hello world", []any{"hello world"}, []any{"5eb63bbbe01eeed093cb22bb8f5acdc3"}, crypto.Md5},
		{"空字符串", []any{""}, []any{"d41d8cd98f00b204e9800998ecf8427e"}, crypto.Md5},
		{"字符串", []any{"test"}, []any{"098f6bcd4621d373cade4e832627b4f6"}, crypto.Md5},
		{"带空格的字符串", []any{"hello world"}, []any{"5eb63bbbe01eeed093cb22bb8f5acdc3"}, crypto.Md5},
		{"CBCEncrypt", []any{"tianlin0"}, []any{true}, func(input string) bool {
			//每次都会不一样
			enStr, err := crypto.CbcEncrypt(input, key, nil)
			if err != nil {
				return false
			}

			oldStr, err := crypto.CbcDecrypt(enStr, key, nil)
			if err != nil {
				return false
			}
			if oldStr == input {
				return true
			}
			return false
		}},
		{"AesCbcEncrypt", []any{"tianlin0"}, []any{true}, func(input string) bool {
			enStr, err := crypto.AesCbcEncrypt(input, key, nil)
			if err != nil {
				return false
			}

			oldStr, err := crypto.AesCbcDecrypt(enStr, key, nil)
			if err != nil {
				return false
			}

			fmt.Println(input, enStr, oldStr)

			if oldStr == input {
				return true
			}
			return false
		}},
		{"AesEncrypt", []any{"tianlin0"}, []any{true}, func(input string) bool {

			appSecret := "IgkibX71IEf382PT"

			enStr, err := crypto.AesEncrypt(input, appSecret, nil)
			if err != nil {
				return false
			}
			oldStr, err := crypto.AesDecrypt(enStr, appSecret, nil)
			if err != nil {
				return false
			}
			if oldStr == input {
				return true
			}
			return false
		}},
		{"Argon2PasswordEncoder", []any{"tianlin0"}, []any{true}, func(input string) bool {
			//每次都会不一样
			enStr, err := crypto.Argon2PasswordEncoder(input)
			if err != nil {
				return false
			}
			fmt.Println("Argon2PasswordEncoder:", input, enStr)

			has, err := crypto.Argon2CompareHashAndPassword(enStr, input)
			if err != nil {
				return false
			}

			return has
		}},
		{"BCryptPasswordEncoder", []any{"tianlin0"}, []any{true}, func(input string) bool {
			//每次都会不一样
			enStr, err := crypto.BCryptPasswordEncoder(input)
			if err != nil {
				return false
			}
			fmt.Println("BCryptPasswordEncoder:", input, enStr)

			has, err := crypto.BCryptCompareHashAndPassword(enStr, input)
			if err != nil {
				return false
			}

			return has
		}},
		{"GobEncode", []any{"tianlin0"}, []any{"tianlin0"}, func(input string) string {
			//每次都会不一样
			enStr, err := crypto.GobEncode(input, nil)
			if err != nil {
				return ""
			}
			fmt.Println("GobEncode:", input, enStr)

			aa := ""
			err = crypto.GobDecode(enStr, &aa, nil)
			if err != nil {
				return ""
			}

			return aa
		}},
		{"XorEncode", []any{"tianlin0"}, []any{"tianlin0"}, func(input string) string {
			keyInt := 5
			enStr := crypto.XorEncode(input, keyInt, nil)

			fmt.Println("XorEncode:", input, enStr)

			deStr, _ := crypto.XorDecode(enStr, keyInt, nil)

			return deStr
		}},
	}
	utils.TestFunction(t, testCases, nil)
}

func TestCreateRSAKeys(t *testing.T) {
	rsa := new(crypto.RSASecurity)
	pub, pri, err := rsa.CreateRSAKeys(0, crypto.PKCS8Type)

	fmt.Println(pub)
	fmt.Println(pri)
	fmt.Println(err)

	pubStr := `-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANQjZeh61+XSP/XgQLv0EeWJylEsTQiV
r/wUSB2pY7XZyEIJ7czGbl3xe5nXxEMHjlgkWH14VzS3+cPIejrraBECAwEAAQ==
-----END PUBLIC KEY-----`
	priStr := `-----BEGIN PRIVATE KEY-----
MIIBVAIBADANBgkqhkiG9w0BAQEFAASCAT4wggE6AgEAAkEA1CNl6HrX5dI/9eBA
u/QR5YnKUSxNCJWv/BRIHaljtdnIQgntzMZuXfF7mdfEQweOWCRYfXhXNLf5w8h6
OutoEQIDAQABAkAhPxb2k2IIq6XIhAfBLSQs5CZoCFheUw9Mo2UV+Pkeg6UW2DHU
TB1N4DFKlCpGpsYHaPgmpRqEilBtdslcNzJNAiEA7ZSBDicBSwChAdCHBKF6mTSh
fWgCfcBmItqNLrNrJhMCIQDkleyOFQjeOVxcTm0RA6cC/f/wJ5lc4++c16eL9Y3N
ywIgUmVDkO30I9f2/xMcEH4Ub9fx/fU5j/VPNt1HQ6AUFCMCIQDXhR/Pare8xrp1
caBV3Wq3YILSjJOFyIdgCti3FmOH9wIgMT5Fzig0Qsp47jwJ0ICRZCaXFYA0XKPI
TjWDiQ4P6p8=
-----END PRIVATE KEY-----`
	rsa.SetPublicAndPrivateKey(pubStr, priStr)
	kk, err := rsa.PubKeyEncrypt("tiantian", nil)
	fmt.Println(kk, err)

	mm, err := rsa.PriKeyDecrypt(kk, nil)
	fmt.Println(mm, err)

	nn, err := rsa.SignMd5("abc", nil)
	fmt.Println(nn, err)
	err = rsa.VerifySignMd5("abc", nn, nil)
	fmt.Println(err)

	qq, err := crypto.EncryptRSA(priStr, "tiantian")
	fmt.Println(qq, err)
	zz, err := crypto.DecryptRSA(pubStr, qq)
	fmt.Println(zz, err)

	qq, err = crypto.EncryptRSA(pubStr, "tiantian")
	fmt.Println(qq, err)
	zz, err = crypto.DecryptRSA(priStr, qq)
	fmt.Println(zz, err)

}

func parseJWT(tokenString string, secret []byte) (*jwtA.Token, error) {
	// 解析 JWT
	token, err := jwtA.Parse(tokenString, func(token *jwtA.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwtA.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	return token, err
}

func TestJwt(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDIzNTExNDAsImlhdCI6MTc0MjI2NDc0MCwidXNlcl9pZCI6MSwidXNlcl9uYW1lIjoiYWRtaW4ifQ.flHmeEgfHSMTmiCx-JO7PCsP7ymHM5hxXDYG1Q_VWi0"
	secret := "sdsafgtrytyuty5656743gtrertert"

	aa, err := parseJWT(token, []byte(secret))
	fmt.Println(err)
	fmt.Println(conv.String(aa))
}
