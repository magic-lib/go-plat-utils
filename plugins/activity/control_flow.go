package activity

type (
	OnExitType  string
	OnErrorType string
)

const (
	OnErrorIgnore  OnErrorType = "ignore"   //遇到错误时，忽略错误，继续执行，比如打日志，或者是发消息，错误就错误了
	OnExitExit     OnExitType  = "exit"     // 直接退出，后续流程不再执行
	OnExitReturn   OnExitType  = "return"   // 直接退出，后续全部异步执行
	OnExitContinue OnExitType  = "continue" // 继续执行
)

type (
	FlowControl struct {
		Timeout       int `yaml:"timeout" json:"timeout,omitempty"`               // 秒，设置超时时间
		DelayDuration int `yaml:"delay_duration" json:"delay_duration,omitempty"` // 延时多长时间执行 ==0 立即执行，> 0 延时执行，<0 异步执行

		OnError OnErrorType `yaml:"on_error" json:"on_error,omitempty"` // 如果出现执行错误了，是直接跳出，还是继续执行
		OnExit  OnExitType  `yaml:"on_exit" json:"on_exit,omitempty"`   // 是否执行完当前的Activity后，就直接返回，后续的则子流程
	}
)

// ShouldExitOnExecute 判断是否在执行后立即退出
func (c *FlowControl) ShouldExitOnExecute() bool {
	if c == nil {
		return false // 默认不退出
	}
	return c.OnExit == OnExitExit
}

// ShouldContinueOnExecute 判断是否在执行后继续执行后续流程
func (c *FlowControl) ShouldContinueOnExecute() bool {
	if c == nil {
		return false // 默认不退出
	}
	return c.OnExit == OnExitContinue
}

// ShouldReturnOnExecute 直接返回，后面流程异步
func (c *FlowControl) ShouldReturnOnExecute() bool {
	if c == nil {
		return false // 默认不退出
	}
	return c.OnExit == OnExitReturn
}
func (c *FlowControl) ShouldIgnoreOnError() bool {
	if c == nil {
		return false
	}
	return c.OnError == OnErrorIgnore
}
