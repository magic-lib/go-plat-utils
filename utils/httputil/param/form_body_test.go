package param

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// 解析 form 后 body 仍然可以读取
func TestGetParamFromFormKeepsBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("a=1&b=2"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	forms := getParamFromForm(r)
	if forms.Get("a") != "1" {
		t.Fatalf("form a = %q", forms.Get("a"))
	}
	if body := string(SafeReadBody(r)); body != "a=1&b=2" {
		t.Fatalf("body = %q, want a=1&b=2", body)
	}
}

// getPathConfig 不能改写调用方传入的配置
func TestGetPathConfigNoMutate(t *testing.T) {
	pc := &PathConfig{}
	got := getPathConfig(pc)
	if got.SplitPath != "/" || got.VarPrefix != ":" {
		t.Fatalf("defaults not applied: %#v", got)
	}
	if pc.SplitPath != "" || pc.VarPrefix != "" {
		t.Fatalf("caller config mutated: %#v", pc)
	}
}
