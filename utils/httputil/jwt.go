package httputil

import (
	"fmt"
	jwtRequest "github.com/golang-jwt/jwt/v4/request"
	"github.com/magic-lib/go-plat-utils/crypto"
	"net/http"
)

const (
	Authorization = "Authorization"
	Bearer        = "Bearer"
)

func ExtractorJwtToken[T any](jwtSecret string, header http.Header, jwtCfgList ...*crypto.JwtCfg) (t T, err error) {
	authorizationValue := header.Get(Authorization)
	if authorizationValue == "" {
		return t, fmt.Errorf("no authorization header")
	}
	authorizationValue, err = jwtRequest.AuthorizationHeaderExtractor.Filter(authorizationValue)
	if err != nil {
		return t, err
	}
	jwtData, _, err := crypto.JwtDecrypt[T](jwtSecret, authorizationValue, nil, jwtCfgList...)
	if err != nil {
		return jwtData, err
	}
	return jwtData, nil
}

func CreateJwtToken(jwtSecret string, data any, cfgList ...*crypto.JwtCfg) (string, error) {
	encodeStr, err := crypto.JwtEncrypt(jwtSecret, data, cfgList...)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s", Bearer, encodeStr), nil
}
