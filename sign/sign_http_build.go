package sign

import (
	"bytes"
	"context"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"io"
	"net/http"
	"net/url"
)

type Params struct {
	Method string     `json:"method"`
	Path   string     `json:"path"`
	Query  url.Values `json:"query"`
	Body   []byte     `json:"body"`
}
type signParams struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Query  string `json:"query"`
	Body   string `json:"body"`
}

// 构造排序后的 querystring
func buildSortedQuery(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	queryMap := make(map[string]any)
	for key, val := range values {
		queryMap[key] = val
	}
	return utils.MapToUrlParams(queryMap)
}

func buildAllSortedParams(ctx context.Context, p *Params, bodyEncode func(ctx context.Context, body []byte) (string, error)) map[string]any {
	newSignParams := signParams{
		Method: p.Method,
		Path:   p.Path,
		Query:  buildSortedQuery(p.Query),
	}
	isBodyEncoded := false
	if bodyEncode != nil {
		bodyStr, err := bodyEncode(ctx, p.Body)
		if err == nil {
			isBodyEncoded = true
			newSignParams.Body = bodyStr
		}
	}
	if !isBodyEncoded {
		newSignParams.Body = string(p.Body)
	}
	queryMap := make(map[string]any)
	_ = conv.Unmarshal(newSignParams, &queryMap)

	return queryMap
}

func readAndRestoreBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return []byte{}, nil
	}
	defer func() {
		_ = r.Body.Close()
	}()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bs))
	return bs, nil
}
