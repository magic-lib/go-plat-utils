package param

import (
	"github.com/magic-lib/go-plat-utils/conv"
	"net/url"
	"sort"
)

// HttpBuildQuery 将map转换为a=1&b=2
func HttpBuildQuery(paramData map[string]any) string {
	params := url.Values{}
	keyList := make([]string, 0)
	for key := range paramData {
		keyList = append(keyList, key)
	}
	sort.Strings(keyList)
	for _, key := range keyList {
		params.Set(key, conv.String(paramData[key]))
	}
	return params.Encode()
}
