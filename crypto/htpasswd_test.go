package crypto_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/crypto"
	"testing"
)

func TestDecodeMd5(t *testing.T) {
	oldPwd := "1"
	salt := "CppP/Mm."
	hash := "$apr1$CppP/Mm.$SSDya7Irg8zZOhSJHqAfc/"

	htp := new(crypto.Htpasswd)
	hash2 := htp.Apr1Md5Password(oldPwd, salt)
	fmt.Println(hash2)
	fmt.Println(htp.Apr1Md5PasswordCompare(hash2, oldPwd))
	fmt.Println(htp.Apr1Md5PasswordCompare(hash, oldPwd))

	hash3 := htp.Sha1Password(oldPwd)
	fmt.Println(hash3)
	fmt.Println(htp.Sha1PasswordCompare(hash3, oldPwd))

	hash4 := htp.BcryptPassword(oldPwd)
	fmt.Println(hash4)
	fmt.Println(htp.BcryptPasswordCompare(hash4, oldPwd))
	//fmt.Println(htp.CryptPasswordCompare("p00mtXNcbnJ3U", oldPwd))
}
