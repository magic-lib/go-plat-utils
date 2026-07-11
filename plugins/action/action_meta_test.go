package action_test

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/plugins/action"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/samber/lo"
	"log"
	"testing"
)

type Order struct {
	Name  string `json:"name"`
	Group string `json:"group"`
}

type ApiAccountInfoReq struct {
	Time  int64  `json:"time"`
	Key   string `json:"key"`
	Accno string `json:"accno"`
	Mid   int64  `json:"mid"`
	Nid   string `json:"nid"`
	Name  string `json:"name"`
	From  string `json:"from"`
	Bcode string `json:"bcode"`
	Cache int    `json:"cache"`
}

func (o *Order) GetOrderName(ctx context.Context, id int) (string, error) {
	if id == 0 {
		return "", fmt.Errorf("err:no id")
	}
	return fmt.Sprintf("%d", id+1), nil
}

func (o *Order) GetMemberGroup(ctx context.Context, id int) (string, error) {
	if id == 0 {
		return "", fmt.Errorf("err:no id")
	}
	return "M8", nil
}
func (o *Order) GetOrderInfo(ctx context.Context, or *Order) (*Order, error) {
	return or, nil
}

func (o *Order) SetOrderInfo(ctx context.Context, orderId string) (bool, error) {
	fmt.Println(orderId)
	return true, nil
}
func (o *Order) Logger(ctx context.Context, or map[string]any) (bool, error) {
	log.Print("Logger aaaaaaa: ", conv.String(or))
	return true, nil
}

var ns = "order"

func registerAction() {
	orderModel := &Order{
		Name: "tianlin999777",
	}
	getOrderNameInterface, err := action.MethodToActor[int, string](orderModel.GetOrderName, &action.ActMetaData{
		Namespace:    ns,
		Action:       "GetOrderName",
		Desc:         "获取订单名称",
		RequiredArgs: []string{"name", "age"},
		RequiredResp: nil,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	getMemberGroupInterface, err := action.MethodToActor[int, string](orderModel.GetMemberGroup, &action.ActMetaData{
		Namespace:    ns,
		Action:       "GetMemberGroup",
		Desc:         "获取用户客群",
		RequiredArgs: []string{"name", "age"},
		RequiredResp: nil,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	setOrderInfoInterface, err := action.MethodToActor[string, bool](orderModel.SetOrderInfo, &action.ActMetaData{
		Namespace:    ns,
		Action:       "SetOrderInfo",
		Desc:         "设置订单信息",
		RequiredArgs: nil,
		RequiredResp: nil,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	getOrderInfoInterface, err := action.MethodToActor[*Order, *Order](orderModel.GetOrderInfo, &action.ActMetaData{
		Namespace:    ns,
		Action:       "GetOrderInfo",
		Desc:         "获取订单信息",
		RequiredArgs: []string{"name", "group"},
		RequiredResp: nil,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	getLoggerInterface, err := action.MethodToActor[*Order, bool](orderModel.Logger, &action.ActMetaData{
		Action: "Log",
		Desc:   "日志",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	//rkbKlwl9OL0ew2cZ6m
	if err != nil {
		fmt.Println(err)
		return
	}

	actionList := []action.Actor{
		getOrderNameInterface,
		setOrderInfoInterface,
		getMemberGroupInterface,
		getOrderInfoInterface,
		getLoggerInterface,
	}

	lo.ForEach(actionList, func(item action.Actor, _ int) {
		err = action.Register(item)
		if err != nil {
			fmt.Println(err)
			return
		}
	})
}

func TestActionRegister(t *testing.T) {
	registerAction()

	actionFun, err := action.GetAction(ns, "GetOrderName")
	if err != nil {
		fmt.Println(err)
		return
	}
	//functionName := actionFun.Name()
	//fmt.Println(functionName)

	testCases := []*utils.TestStruct{
		{"err", []any{0}, []any{"err:no id"}, func(n int) string {
			_, err := actionFun.Execute(context.Background(), n)
			if err != nil {
				return err.Error()
			}
			return ""
		}},
		{"int", []any{5}, []any{"6"}, func(n int) string {
			intStr, err := actionFun.Execute(context.Background(), n)
			if err != nil {
				return err.Error()
			}
			return conv.String(intStr)
		}},
		{"int2", []any{"7"}, []any{"8"}, func(n int) string {
			fmt.Println("\nexecute:")
			mm, err := actionFun.Execute(context.Background(), "7")
			if err != nil {
				return err.Error()
			}
			return conv.String(mm)
		}},
	}
	utils.TestFunction(t, testCases, nil)
}
