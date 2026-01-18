package state_test

import (
	"context"
	"fmt"
	"github.com/looplab/fsm"
	"github.com/magic-lib/go-plat-utils/state"
	flow "github.com/s8sg/goflow/flow/v1"
	goflow "github.com/s8sg/goflow/v1"
	"testing"
)

type AA struct {
	Name string
}

func TestCacheMap(t *testing.T) {

	aa := state.NewBaseFsm[string, string]()
	err1 := aa.Register("a", "b", "c")
	err2 := aa.Register("c", "b", "d")
	err3 := aa.Register("a", "b", "d")

	fmt.Println(aa.StateMap(), err1, err2, err3)
}

// 任务函数：处理数据
func processData(data []byte, option map[string][]string) ([]byte, error) {
	return []byte(fmt.Sprintf("Processed: %s", string(data))), nil
}

// 定义工作流
func DefineWorkflow(workflow *flow.Workflow, ctx *flow.Context) error {
	dag := workflow.Dag()
	dag.Node("test", processData)
	return nil
}

func TestTask1(t *testing.T) {
	// 配置服务
	fs := &goflow.FlowService{
		Port:     8080,
		RedisURL: "localhost:6379",
		//OpenTraceUrl:      "localhost:5775",
		WorkerConcurrency: 5,
		EnableMonitoring:  true,
	}

	// 注册工作流
	fs.Register("myflow", DefineWorkflow)

	// 启动服务（同时作为服务器和工作节点）
	err := fs.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("end")
	select {}
}

//// 定义加法节点
//func addNode(ctx flow.Context, input int) (int, error) {
//	return input + 5, nil
//}
//
//// 定义乘法节点
//func multiplyNode(ctx flow.Context, input int) (int, error) {
//	return input * 3, nil
//}
//
//func TestTask(t *testing.T) {
//	// 创建工作流定义
//	flowDef := flow.NewFlow("myFlow")
//
//	// 添加加法节点
//	addTask := flowDef.AddTask("addTask", addNode)
//
//	// 添加乘法节点，依赖于加法节点的输出
//	multiplyTask := flowDef.AddTask("multiplyTask", multiplyNode).DependsOn(addTask)
//
//	// 注册工作流
//	if err := gorunway.RegisterFlow(flowDef); err != nil {
//		fmt.Printf("注册工作流错误: %v\n", err)
//		return
//	}
//
//	// 启动工作流引擎
//	if err := gorunway.StartEngine(); err != nil {
//		fmt.Printf("启动工作流引擎错误: %v\n", err)
//		return
//	}
//
//	// 触发工作流执行，传入初始输入值为 10
//	result, err := gorunway.RunFlow("myFlow", 10)
//	if err != nil {
//		fmt.Printf("执行工作流错误: %v\n", err)
//		return
//	}
//
//	// 输出结果
//	fmt.Printf("工作流执行结果: %v\n", result)
//
//	// 关闭工作流引擎
//	gorunway.StopEngine()
//}

type Door struct {
	To  string
	FSM *fsm.FSM
}

func NewDoor(to string) *Door {
	d := &Door{
		To: to,
	}

	d.FSM = fsm.NewFSM(
		"closed",
		fsm.Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		fsm.Callbacks{
			"enter_state": func(_ context.Context, e *fsm.Event) { d.enterState(e) },
		},
	)

	return d
}

func (d *Door) enterState(e *fsm.Event) {
	fmt.Printf("The door to %s is %s\n", d.To, e.Dst)
}

//func main() {
//	door := NewDoor("heaven")
//
//	err := door.FSM.Event(context.Background(), "open")
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	err = door.FSM.Event(context.Background(), "close")
//	if err != nil {
//		fmt.Println(err)
//	}
//}
