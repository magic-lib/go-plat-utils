package conf

import "strings"

// EnvCode 环境变量
type EnvCode string

const (
	EnvLoc     EnvCode = "loc"
	EnvDev     EnvCode = "dev"
	EnvPre     EnvCode = "pre"
	EnvRelease EnvCode = "release"
)

// String 转换为小写字符串
func (m EnvCode) String() string {
	temp := string(m)
	return strings.ToLower(temp)
}

// EnvStruct 环境变量结构体
type EnvStruct interface {
	SetEnv(env EnvCode) bool //设置环境变量
	GetEnv() EnvCode         //获取环境变量
	IsTestEnv() bool         //是否是测试环境
}

type commEnvCode struct {
	currentEnv EnvCode
}

// SetEnv 一些地址需要根据环境访问地址不同
func (e *commEnvCode) SetEnv(env EnvCode) bool {
	if env != "" {
		e.currentEnv = env
		return true
	}
	return false
}

// GetEnv 获取当前的环境
func (e *commEnvCode) GetEnv() EnvCode {
	return e.currentEnv
}

// IsTestEnv 是不是测试环境
func (e *commEnvCode) IsTestEnv() bool {
	return e.currentEnv == EnvDev || e.currentEnv == EnvLoc
}

var currentEnvInstance EnvStruct = &commEnvCode{
	currentEnv: EnvRelease, //默认当前使用的环境变量
}

// SetEnvStruct 设置默认环境的实例
func SetEnvStruct(commEnv EnvStruct) {
	currentEnvInstance = commEnv
}

// SetEnv 设置环境变量
func SetEnv(env EnvCode) bool {
	return currentEnvInstance.SetEnv(env)
}

// GetEnv 获取环境变量
func GetEnv() EnvCode {
	return currentEnvInstance.GetEnv()
}

// IsTestEnv 是否是测试环境
func IsTestEnv() bool {
	return currentEnvInstance.IsTestEnv()
}
