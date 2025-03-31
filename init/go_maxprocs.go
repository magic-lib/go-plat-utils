package init

import (
	"math"
	"os"
	"runtime"
	"strconv"
)

// 并行执行最大效率
func setRunTimeProcess() bool {
	pocsSet := os.Getenv("GOMAXPROCS")
	// 环境变量已设置，就不用设置了。
	if pocsSet == "" {
		cpuNumStr := os.Getenv("CPU_LIMIT")
		cpuNum, _ := strconv.ParseFloat(cpuNumStr, 64)
		cpuNumInt := int(math.Ceil(cpuNum))
		if cpuNumInt <= 0 || cpuNumInt >= 20 {
			cpuNumInt = runtime.NumCPU()
		}

		// 如果是宿主机的数量,太大的话，则默认改为10
		if cpuNumInt <= 0 || cpuNumInt >= 20 {
			cpuNumInt = 10
		}
		runtime.GOMAXPROCS(cpuNumInt)
		return true
	}
	return false
}
