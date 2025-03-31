package httputil_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/utils/httputil"
	"testing"
)

func TestPathNode(t *testing.T) {
	// 创建根节点
	root := new(httputil.PathNode)

	// 插入路由
	root.Insert("aa/user/:id/name/:name", "/")
	root.Insert("aa/static/*filepath", "/")
	root.Insert("aa/user/aaaa/names/:bbbb", "/")

	fmt.Println(root)

	// 测试匹配
	map1, ok1 := root.Match("aa/user/123/name/888", "/")
	fmt.Println(map1, ok1) // 匹配到 /user/:id
	map1, ok1 = root.Match("aa/user/aaaa/names/888", "/")
	fmt.Println(map1, ok1) // 匹配到 /user/:id
	map2, ok2 := root.Match("aa/static/css/style.css", "/")
	fmt.Println(map2, ok2) // 匹配到 /static/*filepat
}
