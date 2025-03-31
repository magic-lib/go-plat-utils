package pool_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/pool"
	"testing"
)

// 1、原始对象

type SimpleObject struct {
	Name string
}

// 2、定义Creator

type SimpleObjectCreator struct {
	pool.BaseCreator[*SimpleObject]
}

func (s *SimpleObjectCreator) New() *SimpleObject {
	fmt.Println("new create....")
	return new(SimpleObject)
}
func (s *SimpleObjectCreator) Reset(v *SimpleObject) *SimpleObject {
	v.Name = ""
	return v
}

var _ pool.ObjectBuilder[*SimpleObject] = (*SimpleObjectCreator)(nil)

// 3、注册structPool
var sp = new(pool.StructPool[*SimpleObjectCreator, *SimpleObject])

// 4、使用

func TestPool(t *testing.T) {
	err := sp.InitPool(new(SimpleObjectCreator))
	if err != nil {
		return
	}
	defer sp.ClosePool()

	err = sp.GetWithFunc(func(v *SimpleObject) {
		v.Name = "aaaa"
		fmt.Println(v)
	})
	if err != nil {
		return
	}
	err = sp.GetWithFunc(func(v *SimpleObject) {
		v.Name = "bbb"
		fmt.Println(v)
	})
	if err != nil {
		return
	}

}
