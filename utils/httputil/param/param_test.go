package param

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"net/http"
	"net/url"
	"testing"
)

func TestToString(t *testing.T) {
	a := "a=1&b=2&a=7&a=9"
	q, err := url.ParseQuery(a)

	fmt.Println(conv.String(q), err)
}
func TestToParam(t *testing.T) {
	q, err := url.QueryUnescape("%2B260967714983")
	fmt.Println(conv.String(q), err)
}
func TestConvString(t *testing.T) {
	II := getIns()
	if oneParamMapTemp, ok := II.(map[string]any); ok {
		for key, val := range oneParamMapTemp {
			fmt.Println(key, val)
		}
	}

	//na := conv.ChangeVariableName("aaa_bbb_ccc", "camel")
	//fmt.Println(na)
}

func getIns() any {
	return map[string]any{
		"aaa": 1,
	}
}

func TestContext(t *testing.T) {
	rawQuery := "pp=1&mm=2&pp=2"
	ret, err := url.ParseQuery(rawQuery)

	fmt.Println(ret, err)

	//ctx := context.Background()
	//setContext(ctx)
	//
	//fmt.Println(ctx.Value("aaaa"))

	//na := conv.ChangeVariableName("aaa_bbb_ccc", "camel")
	//fmt.Println(na)
}
func TestParam(t *testing.T) {
	req := new(http.Request)
	req.Method = http.MethodGet
	req.URL = new(url.URL)
	req.URL.RawQuery = "/v1/auth/auth-check?gpid=&exCluster=&paas_name=gdp-appserver-go"
	data := NewParam().GetAll(req)
	fmt.Println(data)
}
func TestPath(t *testing.T) {
	req := new(http.Request)
	req.Method = http.MethodGet
	req.URL = new(url.URL)
	req.URL.RawQuery = "/v1/auth/auth-check?gpid=&exCluster=&paas_name=gdp-appserver-go"
	data := NewParam().SetParsePathFunc(func(r *http.Request) map[string]string {
		return map[string]string{
			"name": "zhangsan",
		}
	}).GetAllQuery(req)
	fmt.Println(data)
}

type AAA struct {
	Name   string `json:"name" validate:"len=10"`
	Name2  string `json:"name2" validate:"required"`
	Number int    `json:"number" validate:"required"`
}

func TestParse(t *testing.T) {
	req := new(http.Request)
	req.Method = http.MethodGet
	req.URL = new(url.URL)
	req.URL.RawQuery = "/v1/auth/auth-check?gpid=&exCluster=&paas_name=gdp-appserver-go"
	aaa := new(AAA)
	err := NewParam().SetParsePathFunc(func(r *http.Request) map[string]string {
		return map[string]string{
			"name":   "1111111111",
			"name2":  "33333",
			"number": "0",
		}
	}).SetValidatorCustomErrorMessages(map[string]string{
		"aaaa":               "dddd",
		"AAA.Name2.required": "aaaaaajkjkjk",
	}, nil).Parse(req, aaa)

	fmt.Println(aaa, err)
}

func setContext(ctx context.Context) context.Context {
	newCtx := context.WithValue(ctx, "aaaa", "bbbb")
	ctx = newCtx
	return newCtx
}

func TestURL1(t *testing.T) {
	var TraceIdKey = CanonicalHeaderKey("x - trace - id")
	fmt.Println(TraceIdKey)
}

func TestPathParam(t *testing.T) {
	testCheckCases := []*utils.TestStruct{
		{"path Check", []any{"/aa/bb/cc", "/aa/:bb/cc"}, []any{true}, nil},
		{"path Check", []any{"/aa/bb/cc", "/aa/:bb/cc/mm"}, []any{false}, nil},
		{"path Check", []any{"/aa/bb/cc", "/aa/:bb/:cc"}, []any{true}, nil},
	}
	utils.TestFunction(t, testCheckCases, PathMatch)

	testParamCases := []*utils.TestStruct{
		{"path Param", []any{"/aa/bb/cc", "/aa/:aa/cc"}, []any{"bb"}, nil},
		{"path Param", []any{"/aa/bb/cc", "/aa/:bb/cc/mm"}, []any{""}, nil},
		{"path Param", []any{"/aa/bb/cc", "/aa/:bb/:aa"}, []any{"cc"}, nil},
	}
	utils.TestFunction(t, testParamCases, func(requestPath, routePath string) string {
		paramTemp := PathParam(requestPath, routePath)
		fmt.Println(paramTemp)
		return paramTemp["aa"]
	})
}
