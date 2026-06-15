// Package param 获取参数
package param

import (
	"bytes"
	"github.com/samber/lo"
	"io"
	"net/http"
)

// SafeReadBody 安全读取body内容，不会丢失数据
func SafeReadBody(r *http.Request, nexts ...func(r *http.Request)) []byte {
	var requestBody []byte
	if r == nil {
		return requestBody
	}

	if r.Body != nil {
		var err error
		requestBody, err = io.ReadAll(r.Body)
		if err == nil {
			_ = r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
	}

	if len(nexts) == 0 {
		return requestBody
	}
	nexts = lo.Filter(nexts, func(next func(r *http.Request), _ int) bool {
		return next != nil
	})
	if len(nexts) == 0 {
		return requestBody
	}

	lo.ForEach(nexts, func(next func(r *http.Request), _ int) {
		next(r)
	})

	// 有可能next里有读取以后，并未放入body，需要重新放置一下。
	if len(requestBody) > 0 {
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}
	return requestBody
}
