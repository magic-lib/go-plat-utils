package param_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils/httputil/param"
	"testing"
)

func TestMergeMap(t *testing.T) {
	t.Run("merge map", func(t *testing.T) {
		m := param.NewDynamicConfigManager(map[string]param.KeySourcePolicy{
			"name": param.KeyPolicyFrontendPriority,
			"mm":   param.KeyPolicyDefaultOnly,
		})
		frontend := map[string]any{
			"name":   "frontend",
			"age":    18,
			"number": 0,
			"mm":     50,
		}
		backend := map[string]any{
			"name": "backend",
			"kk":   56,
			"mm":   20,
		}
		defaults := map[string]any{
			"name": "default",
			"mm":   0,
		}
		result := m.MergeMap(frontend, backend, defaults)
		if result["name"] != "frontend" {
			t.Errorf("expected %s, but got %s", "frontend", result["name"])
		}
		fmt.Println(conv.String(result))
	})
}
