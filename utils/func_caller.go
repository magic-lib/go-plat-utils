package utils

import (
	"path/filepath"
	"runtime"
	"strings"
)

type CallerInfo struct {
	Func string `json:"func"`
	File string `json:"file"`
	Line int    `json:"line"`
}

var unknownCaller = CallerInfo{Func: "unknown", File: "unknown", Line: 0}

func GetCaller(skip int) CallerInfo {
	// skip: 0=GetCaller, 1=调用GetCaller的函数, 2=再往上一层...
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return unknownCaller
	}

	fn := "unknown"
	if f := runtime.FuncForPC(pc); f != nil {
		fn = f.Name() // 包名/方法名一坨全在这
	}

	return CallerInfo{
		Func: fn,
		File: filepath.Base(file), // 线上日志别打全路径，太长还泄露目录结构
		Line: line,
	}
}

func StackTrace(maxDepth int, skip int) []CallerInfo {
	// skip 这里是“从谁开始算起”，一般要跳过 StackTrace 自己 + 上层 wrapper
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(skip, pcs)
	pcs = pcs[:n]

	frames := runtime.CallersFrames(pcs)
	out := make([]CallerInfo, 0, n)

	for {
		fr, more := frames.Next()

		fn := fr.Function
		file := filepath.Base(fr.File)
		line := fr.Line

		// 很多时候我们只关心业务栈，标准库栈就别刷屏了
		if !strings.HasPrefix(fn, "runtime.") &&
			!strings.HasPrefix(fn, "testing.") {
			out = append(out, CallerInfo{Func: fn, File: file, Line: line})
		}

		if !more {
			break
		}
	}
	return out
}

func FirstBusinessCaller(maxDepth int, skip int, ignorePkgList []string) CallerInfo {
	stack := StackTrace(maxDepth, skip)
	for _, ci := range stack {
		ignore := false
		for _, p := range ignorePkgList {
			if strings.Contains(ci.Func, p) {
				ignore = true
				break
			}
		}
		if !ignore {
			return ci
		}
	}
	return unknownCaller
}
