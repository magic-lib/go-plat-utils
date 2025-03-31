package id

import (
	"fmt"
	"github.com/sony/sonyflake"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	generator *sonyflake.Sonyflake
	maxTimes  = 10
	m         sync.Mutex
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// decimalToBase 函数用于将十进制数字转换为指定进制的字符串
func decimalToBase(num uint64, base int) string {
	charset := base62Chars
	if num == 0 {
		return string(charset[0])
	}
	if base < 2 {
		base = 2
	}
	if base > 62 {
		base = 62
	}
	var result []byte
	for num > 0 {
		remainder := num % uint64(base)
		result = append([]byte{charset[remainder]}, result...)
		num /= uint64(base)
	}
	return string(result)
}

// initGenerator 初始化,避免用init
func initGenerator() {
	if generator != nil { // 如果生成器已经初始化，直接返回
		return
	}
	if maxTimes <= 0 { // 如果最大次数小于等于0，记录错误
		log.Panic("initGenerator maxTimes should be greater than 0", maxTimes)
		return
	}

	t, err := time.Parse("2006-01-02 15:04:05", "2025-02-14 00:00:00") // 解析时间字符串
	if err != nil {                                                    // 如果时间解析出错，记录错误
		log.Panic("initGenerator failed to parse time: ", err)
		return
	}

	m.Lock() // 加锁，确保线程安全
	generator = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime:      t,            // 设置生成器的开始时间
		MachineID:      getMachineID, // 设置机器ID生成函数
		CheckMachineID: nil,          // 设置机器ID检查
	})

	if generator != nil {
		m.Unlock()
		return
	}
	// 如果生成器仍然为nil，减少maxTimes并循环尝试初始化
	maxTimes--
	m.Unlock()
	//循环检测
	initGenerator()
}

// Generator id 生成器，接受一个进制值，转换为对应进制的 id: qFS1rdQXe
func Generator(base int) string {
	id := GeneratorInt()
	if id == 0 {
		return ""
	}
	if base < 2 {
		base = 2
	}
	if base > 36 {
		return decimalToBase(id, base)
	}

	// FormatUint base最大36
	return strconv.FormatUint(id, base)
}

// GeneratorInt id 生成器:5817151654986320
func GeneratorInt() uint64 {
	initGenerator()

	id, err := generator.NextID()
	if err != nil {
		// 函数内部重试一次
		id, err = generator.NextID()
		if err != nil {
			return 0
		}
	}
	return id
}

// GeneratorBase32 11字符 generator 的 base=32 参数后的方法，如： 558ffag00ig
func GeneratorBase32() string {
	return Generator(32)
}

// getMachineID 给 sonyFlake 的 machineID 方法赋值
func getMachineID() (uint16, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return 0, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if ip == nil {
			continue
		}

		return uint16(ip[2])<<8 + uint16(ip[3]), nil
	}

	return 0, fmt.Errorf("get nil ip")
}
