package httputil_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/magic-lib/go-plat-utils/utils/httputil"
)

// WriteCommFailure 传入 nil error 时不能 panic
func TestWriteCommFailureNilError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	httputil.WriteCommFailure(context.Background(), httptest.NewRecorder(), nil, 1001)
	if resp := httputil.GetErrorResponse(nil, 1001, nil); resp.Code != 1001 {
		t.Fatalf("Code = %d", resp.Code)
	}
}
