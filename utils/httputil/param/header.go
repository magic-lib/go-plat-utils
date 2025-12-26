package param

import (
	"net/http"
	"net/textproto"
	"strings"
)

// getAllHeaders 取得所有请求的header列表
func getAllHeaders(r *http.Request) http.Header {
	headers := http.Header{}
	if r == nil {
		return headers
	}
	if r.Header != nil && len(r.Header) > 0 {
		headers = r.Header.Clone()
	}
	return headers
}

// CanonicalHeaderKey canonical header key
func CanonicalHeaderKey(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSpace(s)
	return textproto.CanonicalMIMEHeaderKey(s)
}
