package httputil_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/utils/httputil"
)

// 回溯失败的分支写入的通配参数不能残留在结果中
func TestPathNodeBacktrackParams(t *testing.T) {
	root := &httputil.PathNode{}
	root.Insert("/api/:project/detail", "/")
	root.Insert("/api/list/:id", "/")

	params, ok := root.Match("/api/list/7", "/")
	if !ok || params["id"] != "7" {
		t.Fatalf("Match = %v, %v", params, ok)
	}
	if _, has := params["project"]; has {
		t.Fatalf("失败分支的参数不应残留: %v", params)
	}
}
