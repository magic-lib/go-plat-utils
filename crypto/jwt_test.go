package crypto_test

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/magic-lib/go-plat-utils/crypto"
	"github.com/magic-lib/go-plat-utils/utils"
	"log"
	"testing"
	"time"
)

func ExampleMd5() {
	str := "hello"

	res := crypto.SHA(crypto.MD5, str)
	fmt.Println(res)

	// Output:
	// 5d41402abc4b2a76b9719d911017c592
}

//func ExampleMd5_Base64() {
//	str := "hello"
//
//	res := crypto.SHA(crypto.MD5, str, goencrypt.PrintBase64)
//	fmt.Println(res)
//
//	// Output:
//	// XUFAKrxLKna5cZ2REBfFkg==
//}
//func ExampleSha1() {
//	str := "hello"
//	res := crypto.SHA(crypto.SHA1, str)
//	fmt.Println(res)
//
//	// Output:
//	// aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
//}
//func ExampleSha256() {
//	str := "hello"
//	res := crypto.SHA(crypto.SHA256, str)
//	fmt.Println(res)
//
//	// Output:
//	// 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
//}
//func ExampleSha256_Base64() {
//	str := "hello"
//	res := crypto.SHA(crypto.SHA256, str)
//	fmt.Println(res)
//
//	// Output:
//	// 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
//}
//
//func ExampleMd51() {
//	s := "hello world"
//
//	value, err := goencrypt.MD5(s)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	uuid := func(s string) string {
//		d := md5.Sum([]byte(s))
//		return hex.EncodeToString(d[:])
//	}(s)
//
//	if uuid == string(value) {
//		fmt.Println("ok")
//	}
//
//	fmt.Println(uuid)
//	// Output:
//	// 5eb63bbbe01eeed093cb22bb8f5acdc3
//}

type UserData struct {
	UserName string
	PassWord string
}

func TestJwtCase(t *testing.T) {
	secretKey := "secret4324343242secret4324343242secret4324343242secret4324343242secret4324343242secret4324343242"
	oldData := UserData{
		UserName: "mmt1234567",
		PassWord: "abcd",
	}
	testCases := []*utils.TestStruct{
		{"encode string", []any{}, []any{"abcd"}, func() string {
			encodeString, err := crypto.JwtEncrypt(secretKey, oldData, &crypto.JwtCfg{
				StandardClaims: &jwt.StandardClaims{
					ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
				},
				EncryptJsonKeyList: []string{"PassWord"},
			})
			if err != nil {
				t.Error(err)
				return ""
			}
			newData, dataStr, err := crypto.JwtDecrypt[UserData](secretKey, encodeString, &crypto.JwtCfg{
				EncryptJsonKeyList: []string{"PassWord"},
			})
			log.Println(encodeString)
			log.Println(dataStr)
			if err != nil {
				t.Error(err)
				return ""
			}
			return newData.PassWord
		}},
		{"encode all", []any{"aaaa"}, []any{"aaaa"}, func(oldStr string) string {
			kkk, err := crypto.JwtEncrypt(secretKey, oldStr, &crypto.JwtCfg{
				EncryptJsonKeyList: []string{"."},
			})
			if err != nil {
				t.Error(err)
				return ""
			}
			newData, dataStr, err := crypto.JwtDecrypt[string](secretKey, kkk, &crypto.JwtCfg{
				EncryptJsonKeyList: []string{"."},
			})
			log.Println(kkk)
			log.Println(dataStr)
			if err != nil {
				t.Error(err)
				return ""
			}
			return newData
		}},
	}
	utils.TestFunction(t, testCases, nil)
}
