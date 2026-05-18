// Package param 获取参数
package param

import (
	"bytes"
	"io"
	"net/http"
)

// SafeReadBody 安全读取body内容，不会丢失数据
func SafeReadBody(r *http.Request, next func(r *http.Request)) []byte {
	var requestBody []byte
	if r == nil {
		return requestBody
	}

	if r.Body != nil {
		var err error
		requestBody, err = io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
	}

	if next == nil {
		return requestBody
	}

	next(r)

	// 有可能next里有读取以后，并未放入body，需要重新放置一下。
	if len(requestBody) > 0 {
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}
	return requestBody
}
