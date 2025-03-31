package templates

import (
	"github.com/bytedance/go-tagexpr/v2"
)

// TagExpr tag模版格式
func TagExpr(tagName string, structPtrOrReflectValue interface{}) (*tagexpr.TagExpr, error) {
	if tagName == "" {
		tagName = "tagexpr"
	}

	vm := tagexpr.New(tagName)
	tagExpr, err := vm.Run(structPtrOrReflectValue)
	if err != nil {
		return nil, err
	}

	return tagExpr, nil
}
