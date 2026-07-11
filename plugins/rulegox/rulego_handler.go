package rulegox

type RuleChainModel struct {
	ChainID     string `json:"chain_id"`
	Name        string `json:"name"`
	JSONContent string `json:"json_content"` // 存储前端生成的完整定义 JSON
}

//
//func GetEngineInstance(chainID string, jsonContent string) (types.RuleEngine, error) {
//	// a. 先去 RuleGo 全局内存管理器里找，如果已经加载过了，直接返回
//	if engine, exists := rulego.Get(chainID); exists {
//		return engine, nil
//	}
//
//	// c. 动态反序列化并注入到 RuleGo 管理器中（这步执行完，它就在内存里了）
//	engine, err := rulego.New(chainID, []byte(jsonContent))
//	if err != nil {
//		return nil, err
//	}
//
//	return engine, nil
//}
//
//func SaveWorkflowHandler(w http.ResponseWriter, r *http.Request) {
//	var req RuleChainModel
//	json.NewDecoder(r.Body).Decode(&req)
//
//	// Step 1: 持久化到数据库
//	dbMock[req.ChainID] = req.JSONContent
//	fmt.Printf("【数据库】成功保存/更新了工作流: %s\n", req.ChainID)
//
//	// Step 2: 核心热更新！
//	// 如果这个工作流在内存中已经存在，Reload 会在不重启系统、不影响其他并发流量的情况下，
//	// 直接用最新的 JSON 配置把老流程无缝替换掉。如果不存在，则直接创建。
//	err := rulego.Registry.Reload(req.ChainID, []byte(req.JSONContent))
//
//	if err != nil {
//		w.Write([]byte(`{"status":"error","msg":"热更新失败"}`))
//		return
//	}
//	w.Write([]byte(`{"status":"success"}`))
//}
//
//func ExecuteWorkflowHandler(w http.ResponseWriter, r *http.Request) {
//	// 假设请求传来了想执行的流程ID：user_register_flow
//	targetChainID := r.URL.Query().Get("chainId")
//
//	// 动态获取/加载引擎（如果数据库更新了，这里拿到的就是最新版的流程）
//	engine, err := GetEngineInstance(targetChainID)
//	if err != nil {
//		_, _ = w.Write([]byte(err.Error()))
//		return
//	}
//
//	// 准备业务数据
//	inputData := `{"username":"Tom", "age":18}`
//	msg := types.NewMsg(0, "BIZ_START", types.JSON, types.NewMetadata(), inputData)
//
//	// 触发执行并等待结果
//	engine.OnMsgAndWait(msg, types.WithOnEnd(func(ctx types.RuleContext, msg types.RuleMsg, err error, relationType string) {
//		_, _ = w.Write(msg.Data.GetBytes())
//	}))
//}
