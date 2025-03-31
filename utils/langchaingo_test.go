package utils_test

import (
	"context"
	"fmt"
	graphA "github.com/dominikbraun/graph"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langgraphgo/graph"
	"sync"
	"testing"
)

func TestAI(t *testing.T) {
	g := graph.NewMessageGraph()

	g.AddNode("oracle", func(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error) {
		fmt.Println("oracle:", state)

		return append(state,
			llms.TextParts(llms.ChatMessageTypeSystem, "我是系统消息 oracle"),
		), nil

	})
	g.AddNode("redis", func(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error) {
		fmt.Println("redis:", state)

		return append(state,
			llms.TextParts(llms.ChatMessageTypeSystem, "我是系统消息 redis"),
		), nil

	})
	g.AddNode("mysql", func(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error) {
		fmt.Println("mysql:", state)

		return append(state,
			llms.TextParts(llms.ChatMessageTypeSystem, "我是系统消息 mysql"),
		), nil

	})
	g.AddNode("component", func(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error) {
		fmt.Println("component:", state)

		return append(state,
			llms.TextParts(llms.ChatMessageTypeSystem, "我是系统消息 component"),
		), nil

	})
	g.AddNode(graph.END, func(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error) {
		return state, nil
	})

	g.AddEdge("oracle", "redis")
	g.AddEdge("oracle", "mysql")
	g.AddEdge("redis", "component")
	g.AddEdge("mysql", "component")
	g.AddEdge("component", graph.END)
	g.SetEntryPoint("oracle")

	runnable, err := g.Compile()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// Let's run it!
	res, err := runnable.Invoke(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is 1 + 1?"),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res)

}

// 定义一个结构体来存储所需的字段
type ResponseData struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

// 处理单个 URL 并提取所需字段
func processSingleURL(url string) (ResponseData, error) {
	fmt.Println(url)
	return ResponseData{Field1: url, Field2: "aaaa"}, nil
}

// 定义任务图节点
type TaskNode struct {
	URL  string
	Data ResponseData
}

// 获取某个节点的依赖节点
func getDependencies(g graphA.Graph[string, TaskNode], node string) ([]string, error) {
	var dependencies []string
	edges, err := g.Edges()
	if err != nil {
		return nil, err
	}
	for _, edge := range edges {
		if edge.Target == node {
			dependencies = append(dependencies, edge.Source)
		}
	}
	return dependencies, nil
}

// 执行图编排任务
func executeGraph(g graphA.Graph[string, TaskNode]) ([]ResponseData, error) {
	// 拓扑排序获取执行顺序
	order, err := graphA.TopologicalSort(g)
	if err != nil {
		return nil, err
	}

	var results []ResponseData
	var wg sync.WaitGroup

	// 记录每个节点的依赖是否完成
	dependenciesCompleted := make(map[string]bool)
	for _, node := range order {
		dependenciesCompleted[node] = false
	}

	// 遍历每个节点
	for _, node := range order {
		// 获取当前节点的依赖节点
		neighbors, err := getDependencies(g, node)
		if err != nil {
			return nil, err
		}

		// 检查依赖是否完成
		allDependenciesCompleted := true
		for _, neighbor := range neighbors {
			if !dependenciesCompleted[neighbor] {
				allDependenciesCompleted = false
				break
			}
		}

		if allDependenciesCompleted {
			wg.Add(1)
			go func(node string) {
				defer wg.Done()
				task, err := g.Vertex(node)
				if err != nil {
					fmt.Printf("Error getting vertex %s: %v\n", node, err)
					return
				}
				data, err := processSingleURL(task.URL)
				if err != nil {
					fmt.Printf("Error processing URL %s: %v\n", task.URL, err)
					return
				}
				task.Data = data
				results = append(results, data)
				dependenciesCompleted[node] = true
			}(node)
		}
	}

	wg.Wait()
	return results, nil
}

func TestURL(t *testing.T) {
	// 创建一个有向图
	taskHash := func(c TaskNode) string {
		return c.URL
	}
	g := graphA.New(taskHash, graphA.Directed())

	// 添加任务节点
	task1 := TaskNode{URL: "https://example.com/api/endpoint1"}
	task2 := TaskNode{URL: "https://example.com/api/endpoint2"}
	task3 := TaskNode{URL: "https://example.com/api/endpoint3"}

	// 添加节点到图中
	_ = g.AddVertex(task1)
	_ = g.AddVertex(task2)
	_ = g.AddVertex(task3)

	// 定义依赖关系，例如 task2 依赖 task1，task3 依赖 task2
	_ = g.AddEdge(task1.URL, task2.URL)
	_ = g.AddEdge(task2.URL, task3.URL)

	// 执行图编排任务
	results, err := executeGraph(g)
	if err != nil {
		fmt.Printf("Error executing graph: %v\n", err)
		return
	}

	// 输出合成的字段
	for _, result := range results {
		fmt.Printf("Field1: %s, Field2: %s\n", result.Field1, result.Field2)
	}
}
