package activity_test

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/plugins/action"
	"github.com/magic-lib/go-plat-utils/plugins/activity"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/magic-lib/go-plat-utils/utils/httputil/param"
	"testing"
)

func TestActivityExecute(t *testing.T) {
	oneAct := activity.Activity{
		Id:           "",
		ActNamespace: "",
		ActName:      "",
		Arguments: []*param.BindConfig{
			{
				Key:   "name",
				Value: "{{ffff.responses.id}}",
			},
		},
		ArgTemplate: "{{aaaa.responses.id}}+{{return_bbbb.responses.mm}}+{{cccc.aa}}",
		Responses:   nil,
		Hooks:       nil,
		Control:     activity.ActivityControl{},
	}

	testCases := []*utils.TestStruct{
		{"err", []any{0}, []any{"err:no id"}, func(n int) string {
			//task.WithReturnKeyPrefix("return_")

			fmt.Println(oneAct)

			return ""
		}},
	}
	utils.TestFunction(t, testCases, nil)
}

// 1. 加法：输入两个数字，返回和
type AddReq struct {
	A int `json:"a"`
	B int `json:"b"`
}

func AddMethod(_ context.Context, req *AddReq) (int, error) {
	return req.A + req.B, nil
}

// 2. 拼接字符串：输入两个字符串，返回拼接结果
type ConcatReq struct {
	S1 string `json:"s1"`
	S2 string `json:"s2"`
}

func ConcatMethod(_ context.Context, req *ConcatReq) (string, error) {
	return req.S1 + req.S2, nil
}

// 3. 打印日志：输入任意 map，返回 echo
func LoggerMethod(_ context.Context, req map[string]any) (bool, error) {
	fmt.Println("[LoggerMethod] received:", conv.String(req))
	return true, nil
}

// ============== 注册 Action ==============

func registerTestActions() {
	// action 1: add
	addActor, err := action.MethodToActor[*AddReq, int](AddMethod, &action.ActMetaData{
		Namespace:    "test",
		Action:       "add",
		Desc:         "加法计算",
		RequiredArgs: []string{"a", "b"},
	})
	if err != nil {
		panic(err)
	}
	// action 2: concat
	concatActor, err := action.MethodToActor[*ConcatReq, string](ConcatMethod, &action.ActMetaData{
		Namespace:    "test",
		Action:       "concat",
		Desc:         "字符串拼接",
		RequiredArgs: []string{"s1", "s2"},
	})
	if err != nil {
		panic(err)
	}
	// action 3: logger
	loggerActor, err := action.MethodToActor[map[string]any, bool](LoggerMethod, &action.ActMetaData{
		Namespace: "test",
		Action:    "logger",
		Desc:      "日志打印",
	})
	if err != nil {
		panic(err)
	}

	for _, a := range []action.Actor{addActor, concatActor, loggerActor} {
		if err := action.Register(a); err != nil {
			panic(err)
		}
	}
}

func TestActivityExecuteDemo(t *testing.T) {
	registerTestActions()
	ctx := context.Background()

	// ---------- Activity 1: add（使用 Arguments 绑定参数）----------
	addAct := &activity.Activity{
		Id:           "add1",
		ActNamespace: "test",
		ActName:      "add",
		Arguments: []*param.BindConfig{
			{Key: "a", Value: 3, Policy: param.KeyPolicyDefaultOnly},
			{Key: "b", Value: 5, Policy: param.KeyPolicyDefaultOnly},
		},
		Control: activity.ActivityControl{
			CtxCacheable: true, // 启用流程级缓存
		},
	}
	ret, err := addAct.Execute(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("addAct execute error: %v", err)
	}
	t.Logf("add1 result: %v", conv.String(ret))
	//if conv.String(ret["add1"]["responses"]) != "8" {
	//	t.Fatalf("add1 expect sum=8, got %v", ret["sum"])
	//}

	// ---------- Activity 2: concat（使用 ArgTemplate 动态取依赖结果）----------
	concatAct := &activity.Activity{
		Id:           "concat1",
		ActNamespace: "test",
		ActName:      "concat",
		// 从 add1 的返回值中取 sum，转成字符串作为 s1
		ArgTemplate: `{"s1":"{{add1.responses}}","s2":"!"}`,
		DependsOn: []*activity.Activity{
			addAct,
		},
	}
	ret2, err := concatAct.Execute(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("concatAct execute error: %v", err)
	}
	t.Logf("concat1 result: %v", conv.String(ret2))

	// ---------- Activity 3: logger（使用 Hooks + When 条件）----------
	loggerAct := &activity.Activity{
		Id:           "log1",
		ActNamespace: "test",
		ActName:      "logger",
		// 仅当 add1 的 sum > 5 时才执行
		Control: activity.ActivityControl{
			When: `{{add1.responses}} > 5`,
		},
		DependsOn: []*activity.Activity{addAct},
	}
	ret3, err := loggerAct.Execute(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("loggerAct execute error: %v", err)
	}
	t.Logf("log1 result: %v", conv.String(ret3))

	// ---------- Activity 4: 使用 Hooks（OnStart 钩子，自身不绑定 Action）----------
	hookAct := &activity.Activity{
		Id:           "hook_log",
		ActNamespace: "test",
		ActName:      "concat",
		Hooks: activity.LifecycleHooks{
			activity.LifecycleEventOnStart: &activity.Activity{
				ActNamespace: "test",
				ActName:      "logger",
				Arguments: []*param.BindConfig{
					{Key: "msg", Value: "on start hook triggered", Policy: param.KeyPolicyDefaultOnly},
				},
			},
		},
	}
	_, err = hookAct.Execute(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("hookAct execute error: %v", err)
	}

	// ---------- Activity 5: DelayDuration 异步执行（不阻塞）----------
	asyncAct := &activity.Activity{
		Id:           "async1",
		ActNamespace: "test",
		ActName:      "logger",
		Control: activity.ActivityControl{
			DelayDuration: -1, // <0 异步执行
		},
	}
	_, err = asyncAct.Execute(ctx, map[string]any{"msg": "async message"})
	if err != nil {
		t.Fatalf("asyncAct execute error: %v", err)
	}

	t.Log("TestActivityExecute all passed")
}
