package goroutines

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"reflect"
)

// callFun 动态调用map中保存的方法
type callFun struct {
	result []reflect.Value
	params []interface{}
	fun    interface{}
}

func (c *callFun) checkParam() error {
	f := reflect.ValueOf(c.fun)
	fType := f.Type()
	if fType.Kind() != reflect.Func {
		return fmt.Errorf("fun not function")
	}
	if len(c.params) != fType.NumIn() {
		return fmt.Errorf("the number of params is not adapted: %d, %d", len(c.params), fType.NumIn())
	}
	return nil
}
func (c *callFun) execute() (*callFun, error) {
	if err := c.checkParam(); err != nil {
		return nil, err
	}
	in := make([]reflect.Value, len(c.params))
	for k, param := range c.params {
		in[k] = reflect.ValueOf(param)
	}
	f := reflect.ValueOf(c.fun)
	fType := f.Type()
	result := f.Call(in)
	if len(result) != fType.NumOut() {
		return nil, fmt.Errorf("the number of return is not adapted: %d, %d", len(result), fType.NumOut())
	}
	c.result = result
	return c, nil
}

func (c *callFun) SetFunc(fun interface{}) *callFun {
	f := reflect.ValueOf(fun)
	fType := f.Type()
	if fType.Kind() == reflect.Func {
		c.fun = fun
		c.result = nil
	}
	return c
}
func (c *callFun) SetParams(params ...interface{}) *callFun {
	c.params = make([]interface{}, 0)
	c.params = append(c.params, params...)
	c.result = nil
	return c
}

func (c *callFun) Get(i int, result interface{}) error {
	if c.result == nil {
		_, err := c.execute()
		if err != nil {
			return err
		}
	}
	if i >= len(c.result) || i < 0 {
		max := len(c.result)
		if max > 0 {
			max--
		}
		return fmt.Errorf("i out fun return Number get i: %d, max: %d", i, max)
	}

	oneRet := c.result[i]
	return conv.Unmarshal(oneRet.Interface(), result)
}

func (c *callFun) GetAll(result ...interface{}) error {
	if c.result == nil {
		_, err := c.execute()
		if err != nil {
			return err
		}
	}
	for i, one := range result {
		// 如果长度超过了，则直接退出执行
		if i >= len(c.result) {
			break
		}
		err := c.Get(i, one)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewCallFunc() *callFun {
	return new(callFun)
}
