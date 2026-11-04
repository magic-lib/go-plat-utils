package crypto

import (
	"testing"

	"github.com/golang-jwt/jwt"
)

// JwtEncrypt 不能改写调用方传入的 StandardClaims
func TestJwtEncryptNoMutateClaims(t *testing.T) {
	cfg := &JwtCfg{StandardClaims: &jwt.StandardClaims{}}
	if _, err := JwtEncrypt("secret", map[string]any{"a": 1}, cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.StandardClaims.IssuedAt != 0 || cfg.StandardClaims.ExpiresAt != 0 {
		t.Fatalf("claims mutated: %#v", cfg.StandardClaims)
	}
}
