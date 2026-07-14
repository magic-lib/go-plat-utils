package action

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/utils"
	"reflect"
)

// MethodToActor 转换为Actor
func MethodToActor[I, O any](method any, ac *ActMetaData) (Actor, error) {
	if ac == nil || method == nil {
		return nil, fmt.Errorf("data or method is nil")
	}
	if ac.Name() == "" {
		return nil, fmt.Errorf("action name is empty")
	}

	methodType := reflect.TypeOf(method)
	if methodType == nil || methodType.Kind() != reflect.Func {
		return nil, fmt.Errorf("method must be a callable function/method, got %T", method)
	}

	newMethod, err := utils.ContextMethodToAnyHandler[I, O](method)
	if err != nil {
		return nil, err
	}
	if ac.ArgumentType == nil {
		var zero I
		ac.ArgumentType = reflect.TypeOf(zero)
	}
	ac.actionMethod = newMethod
	return ac, nil
}

// RegisterActor 初始化完直接注册，这样方便，避免重复操作。
func RegisterActor[I, O any](method any, ac *ActMetaData) error {
	actor, err := MethodToActor[I, O](method, ac)
	if err != nil {
		return err
	}
	return Register(actor)
}
