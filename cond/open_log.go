package cond

var openLog = false

// OpenShowLog 打开日志
func OpenShowLog() {
	openLog = true
}
func IsOpenLog() bool {
	return openLog
}
