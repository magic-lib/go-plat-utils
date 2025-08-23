package crypto_test

import (
	"crypto/rand"
	"fmt"
	"github.com/forgoer/openssl"
	"github.com/magic-lib/go-plat-utils/crypto"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestSHAForHmac(t *testing.T) {
	openssl.AesECBEncrypt(nil, nil, "")

	key := "tianlin020250214"
	testCases := []*utils.TestStruct{
		{"SHAWithHmac", []any{crypto.SHA256, "hello world", key}, []any{"b390cd4fcd9864133e838efa76ee3e0b0e0b4774dc04a646edc956ba34b8072c"}, crypto.SHAWithHmac},
	}
	utils.TestFunction(t, testCases, nil)
}

func generateRandomBytes22(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
func TestSHAForHmac1(t *testing.T) {
	mm, _ := generateRandomBytes22(2)

	fmt.Printf("%x", mm)

	fmt.Println("")

}
