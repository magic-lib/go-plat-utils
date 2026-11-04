package utils_test

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/magic-lib/go-plat-utils/utils/funcname"
)

type structInfoT struct {
	Name string `json:"name"`
}

type respIface interface{ Get() string }

func TestGetStructInfo(t *testing.T) {
	for _, obj := range []any{structInfoT{}, &structInfoT{}} {
		info := utils.GetStructInfo(obj)
		if info.PackageName != "utils_test" || info.TypeName != "structInfoT" || len(info.Fields) != 1 {
			t.Errorf("GetStructInfo(%T) = %+v", obj, info)
		}
	}
	if info := utils.GetStructInfo(5); len(info.Fields) != 0 {
		t.Errorf("非结构体不应有字段: %+v", info)
	}
}

func TestFuncNameNonFunc(t *testing.T) {
	if name, isMethod := utils.GetFuncName(structInfoT{}); name != "" || isMethod {
		t.Errorf("GetFuncName(struct) = %q, %v", name, isMethod)
	}
	if got := funcname.Of(structInfoT{}); got != funcname.Unknown {
		t.Errorf("funcname.Of(struct) = %q, want %q", got, funcname.Unknown)
	}
}

// TResp 为接口时，方法返回 (nil, err) 应透传真实错误
func TestContextMethodToAnyHandlerKeepsError(t *testing.T) {
	wantErr := errors.New("real error")
	h, err := utils.ContextMethodToAnyHandler[string, respIface](func(ctx context.Context, s string) (respIface, error) {
		return nil, wantErr
	})
	if err != nil {
		t.Fatalf("ContextMethodToAnyHandler: %v", err)
	}
	if _, err = h(context.Background(), "x"); !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestRandomIntInRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		if v := utils.RandomIntInRange(3, 5); v < 3 || v > 5 {
			t.Fatalf("RandomIntInRange(3,5) = %d", v)
		}
	}
	_ = utils.RandomIntInRange(math.MinInt64, math.MaxInt64) // 原来区间长度溢出会 panic
}
