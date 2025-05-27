package templates_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/templates"
	"testing"
)

// RuleExpr 字符串规则引擎，也是模版的一种
func TestRuleExpr(t *testing.T) {
	aaaa, err := templates.Template("{{.name}} dffff,,", map[string]any{
		"name": "jinjin",
	})
	fmt.Println(aaaa, err)
}
