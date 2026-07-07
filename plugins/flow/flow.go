package flow

// FlowNode 节点基础通用结构，所有流程节点统一基类
type FlowNode struct {
	NodeId     string       `json:"nodeId"`     // 节点唯一ID
	NodeType   string       `json:"nodeType"`   // 固定值：start, switch，decision, end,
	NodeName   string       `json:"nodeName"`   // 展示名称：客群判断
	Desc       string       `json:"desc"`       // 描述，对应图中 No description
	OutLinks   []NodeLink   `json:"outLinks"`   // 输出连线：label -> 下游节点ID
	SwitchConf SwitchConfig `json:"switchConf"` // 切换专属配置
}

// 连线映射：每个分支标签对应下游节点
type NodeLink struct {
	CaseLabel string `json:"caseLabel"` // PMEC / ZANACO / Default
	TargetId  string `json:"targetId"`  // 该分支跳转的下一个节点ID
}

// 客群判断切换核心配置（业务语义层，无JS硬编码）
type SwitchConfig struct {
	Mode         string       `json:"mode"`         // MatchFirst / MatchAll
	VarName      string       `json:"varName"`      // 匹配变量：customerType
	MatchType    string       `json:"matchType"`    // EqualsIgnoreCase 忽略大小写相等
	Cases        []SwitchCase `json:"cases"`        // 业务分支列表
	DefaultLabel string       `json:"defaultLabel"` // 兜底分支固定标签：Default
}

// 单条业务分支（只存业务值，不存脚本）
type SwitchCase struct {
	Label string `json:"label"` // 展示标签：PMEC
	Value string `json:"value"` // 匹配值：PMEC（统一大写存储）
}

// InputField 活动输入参数元数据定义
type InputField struct {
	IsWrapped    bool   `json:"isWrapped"`
	UiHint       string `json:"uiHint"`
	IsSensitive  bool   `json:"isSensitive"`
	AutoEvaluate bool   `json:"autoEvaluate"`
	Name         string `json:"name"`
	TypeName     string `json:"typeName"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
	Order        int    `json:"order"`
	IsBrowsable  bool   `json:"isBrowsable"`
	IsSynthetic  bool   `json:"isSynthetic"`
}

// OutputField 活动输出参数元数据定义
type OutputField struct {
	Name        string `json:"name"`
	TypeName    string `json:"typeName"`
	DisplayName string `json:"displayName"`
	Order       int    `json:"order"`
	IsBrowsable bool   `json:"isBrowsable"`
	IsSynthetic bool   `json:"isSynthetic"`
}

// ActivityCustomProperties 活动自定义扩展属性
type ActivityCustomProperties struct {
	Type     string `json:"Type"`
	RootType string `json:"RootType"`
}

// WorkflowActivity Elsa 流程活动顶层通用结构体
// 兼容 IntegrationServiceActivity / FlowSwitch / FlowDecision 所有活动元数据
type WorkflowActivity struct {
	TypeName               string                   `json:"typeName"`
	ClrType                string                   `json:"clrType"`
	Namespace              string                   `json:"namespace"`
	Name                   string                   `json:"name"`
	Version                int                      `json:"version"`
	Category               string                   `json:"category"`
	DisplayName            string                   `json:"displayName"`
	Description            string                   `json:"description"`
	Inputs                 []InputField             `json:"inputs"`
	Outputs                []OutputField            `json:"outputs"`
	Kind                   string                   `json:"kind"`
	RunAsynchronously      bool                     `json:"runAsynchronously"`
	Ports                  []any                    `json:"ports"`
	CustomProperties       ActivityCustomProperties `json:"customProperties"`
	ConstructionProperties map[string]any           `json:"constructionProperties"`
	IsContainer            bool                     `json:"isContainer"`
	IsBrowsable            bool                     `json:"isBrowsable"`
	IsStart                bool                     `json:"isStart"`
	IsTerminal             bool                     `json:"isTerminal"`
}
