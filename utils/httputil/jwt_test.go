package httputil_test

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/magic-lib/go-plat-utils/crypto"
	"github.com/magic-lib/go-plat-utils/utils/httputil"
	"net/http"
	"testing"
	"time"
)

type UserData struct {
	UserName string
	PassWord string
}

func TestJwt(t *testing.T) {
	secretKey := "secret4324343242secret4324343242secret4324343242secret4324343242secret4324343242secret4324343242"
	oldData := UserData{
		UserName: "mmt1234567",
		PassWord: "abcd",
	}
	tokenStr, err := httputil.CreateJwtToken(secretKey, oldData, &crypto.JwtCfg{
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
		EncryptJsonKeyList: []string{"PassWord"},
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(tokenStr)

	head := http.Header{
		httputil.Authorization: []string{tokenStr},
	}

	data, err := httputil.ExtractorJwtToken[UserData](secretKey, head, &crypto.JwtCfg{
		EncryptJsonKeyList: []string{"PassWord"},
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(data)

}
