package param_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/magic-lib/go-plat-utils/utils/httputil/param"
)

// 加密失败（值超出 securecookie 长度上限）时不能下发明文 cookie
func TestEncodeCookieFailNoPlaintext(t *testing.T) {
	sc := param.NewSecureCookie(nil, nil)
	rec := httptest.NewRecorder()
	err := sc.EncodeCookieToResponse(rec, &http.Cookie{Name: "token", Value: strings.Repeat("x", 10000)})
	if err == nil {
		t.Fatal("超长值应加密失败")
	}
	if got := rec.Header().Get("Set-Cookie"); got != "" {
		t.Fatalf("加密失败时不应写 cookie, got %.60s", got)
	}
}
