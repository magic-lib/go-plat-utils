package crypto_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/crypto"
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
