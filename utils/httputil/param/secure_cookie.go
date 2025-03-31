package param

import (
	"fmt"
	"github.com/gorilla/securecookie"
	"net/http"
)

type secureCookie struct {
	hashKey  []byte
	blockKey []byte
	sc       *securecookie.SecureCookie
}

// NewSecureCookie 新建一个安全的cookie
func NewSecureCookie(hashKey, blockKey []byte) *secureCookie {
	if hashKey == nil || len(hashKey) == 0 {
		hashKey = securecookie.GenerateRandomKey(64)
	}
	if blockKey == nil || len(blockKey) == 0 {
		blockKey = securecookie.GenerateRandomKey(32)
	}
	return &secureCookie{
		hashKey:  hashKey,
		blockKey: blockKey,
		sc:       securecookie.New(hashKey, blockKey),
	}
}

// Encode 加密一个cookieValue
func (s *secureCookie) Encode(cookieName, cookieValue string) (string, error) {
	if cookieName == "" || cookieValue == "" {
		return "", fmt.Errorf("cookie name or value empty: %s, %s", cookieName, cookieValue)
	}
	encoded, err := s.sc.Encode(cookieName, cookieValue)
	if err != nil {
		return "", err
	}
	return encoded, nil
}

// Decode 解密一个cookieValue
func (s *secureCookie) Decode(cookieName, cookieValue string) (string, error) {
	var dstValue string
	err := s.sc.Decode(cookieName, cookieValue, &dstValue)
	if err != nil {
		return "", err
	}
	return dstValue, nil
}
func (s *secureCookie) EncodeCookieToResponse(w http.ResponseWriter, cookie *http.Cookie) error {
	if cookie == nil {
		return nil
	}
	encoded, err := s.Encode(cookie.Name, cookie.Value)
	if err != nil {
		http.SetCookie(w, cookie)
		return err
	}
	cookie.Value = encoded
	cookie.Secure = true
	cookie.HttpOnly = true

	http.SetCookie(w, cookie)
	return nil
}
func (s *secureCookie) DecodeCookieFromRequest(r *http.Request, cookieName string) (string, error) {
	var cookie *http.Cookie
	var err error
	if cookie, err = r.Cookie(cookieName); err != nil {
		return "", err
	}
	if cookie == nil {
		return "", fmt.Errorf("cookie not found: %s", cookieName)
	}
	return s.Decode(cookieName, cookie.Value)
}
