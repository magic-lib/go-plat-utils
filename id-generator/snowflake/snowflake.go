package snowflake

import (
	"fmt"
	"github.com/samber/lo"
	"sync"
	"time"
)

// Config xxxx
type Config struct {
	TimeBits    uint8 // 时间位数
	CenterBits  uint8 // 数据中心位数
	WorkerBits  uint8 // 机器id位数
	SeqBits     uint8 // 序号位数
	Epoch       int64 // 开始毫秒时间戳
	centerMax   int64 // 检测数据中心是否溢出
	workerMax   int64 // 检测机器中心是否溢出
	seqMax      int64 // 检测序号是否溢出
	timeShift   uint8 // 时间戳左移位数
	centerShift uint8 // 数据中心左移位数
	workerShift uint8 // 机器id左移位数
}

var (
	defaultFlakeConfig *Config
	bigIntBits         uint8 = 63
)

// DefaultConfig 设置默认配置
func DefaultConfig(cfg *Config) {
	if cfg == nil {
		cfg = &Config{
			CenterBits: 5,
			WorkerBits: 5,
			Epoch:      1711442064000, //北京时间：2024-03-26 16:34:24
			SeqBits:    12,
		}
	}
	if defaultFlakeConfig == nil {
		defaultFlakeConfig = cfg
	}

	defaultTimeBits := bigIntBits - defaultFlakeConfig.CenterBits - defaultFlakeConfig.WorkerBits - defaultFlakeConfig.SeqBits
	defaultFlakeConfig.TimeBits = lo.Ternary(cfg.TimeBits > 0, cfg.TimeBits, defaultFlakeConfig.TimeBits)
	defaultFlakeConfig.TimeBits = lo.Ternary(defaultFlakeConfig.TimeBits <= 0, defaultTimeBits, defaultFlakeConfig.TimeBits)

	defaultFlakeConfig.CenterBits = lo.Ternary(cfg.CenterBits > 0, cfg.CenterBits, defaultFlakeConfig.CenterBits)
	defaultFlakeConfig.WorkerBits = lo.Ternary(cfg.WorkerBits > 0, cfg.WorkerBits, defaultFlakeConfig.WorkerBits)
	defaultFlakeConfig.SeqBits = lo.Ternary(cfg.SeqBits > 0, cfg.SeqBits, defaultFlakeConfig.SeqBits)
	defaultFlakeConfig.Epoch = lo.Ternary(cfg.Epoch > defaultFlakeConfig.Epoch, cfg.Epoch, defaultFlakeConfig.Epoch)

	defaultFlakeConfig.centerMax = -1 ^ (-1 << defaultFlakeConfig.CenterBits)
	defaultFlakeConfig.workerMax = -1 ^ (-1 << defaultFlakeConfig.WorkerBits)
	defaultFlakeConfig.seqMax = -1 ^ (-1 << defaultFlakeConfig.SeqBits)
	defaultFlakeConfig.timeShift = defaultFlakeConfig.CenterBits + defaultFlakeConfig.WorkerBits + defaultFlakeConfig.SeqBits
	defaultFlakeConfig.centerShift = defaultFlakeConfig.WorkerBits + defaultFlakeConfig.SeqBits
	defaultFlakeConfig.workerShift = defaultFlakeConfig.SeqBits
}

type Worker struct {
	mu sync.Mutex
	// 开始毫秒时间戳 北京时间：2024-03-26 16:34:24
	Epoch    int64
	CenterId int64
	WorkerId int64
	// 当前毫秒已经生成的序列号，从0开始累加
	seq int64
	// 上次生成的毫秒时间戳，用来检查时钟回退
	lastTimestamp int64
}

// New 返回一个Worker实例
func New(w *Worker) (*Worker, error) {
	if w == nil {
		return nil, fmt.Errorf("w is nil")
	}
	if defaultFlakeConfig == nil {
		DefaultConfig(nil)
	}

	w.Epoch = lo.Ternary(w.Epoch <= 0, defaultFlakeConfig.Epoch, w.Epoch)

	if w.CenterId < 0 || w.CenterId > defaultFlakeConfig.centerMax ||
		w.WorkerId < 0 || w.WorkerId > defaultFlakeConfig.workerMax {
		return nil, fmt.Errorf("incorrect CenterId or WorkerId")
	}
	return w, nil
}

// NextId 获取序列号
func (w *Worker) NextId() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	currentTimestamp := time.Now().UnixMilli()
	if currentTimestamp < w.lastTimestamp {
		//时间调整了，需要将当前时间在以前的基础上增加1ms处理
		currentTimestamp = w.lastTimestamp + 1
	}

	if currentTimestamp == w.lastTimestamp {
		w.seq = (w.seq + 1) & defaultFlakeConfig.seqMax
		if w.seq == 0 {
			// 环状，超过了当前毫秒能够获取的最大序列号，那么就自旋等待下一个毫秒
			for currentTimestamp <= w.lastTimestamp {
				currentTimestamp = time.Now().UnixMilli()
			}
		}
	} else {
		// 当前时间和上一个毫秒数不一致, 返回0
		w.seq = 0
	}

	w.lastTimestamp = currentTimestamp
	total := ((currentTimestamp - w.Epoch) << defaultFlakeConfig.timeShift) |
		(w.CenterId << defaultFlakeConfig.centerShift) |
		(w.WorkerId << defaultFlakeConfig.workerShift) |
		w.seq
	//需要将最高位设置为0
	if total>>bigIntBits == -1 {
		return (total | 0x4000000000000000) & 0x7FFFFFFFFFFFFFFF
	}
	return total
}

func (w *Worker) NextIdList(num int) []int64 {
	idNumList := make([]int64, 0, num)
	if num <= 0 {
		return idNumList
	}

	for i := 0; i < num; i++ {
		var id int64
		for {
			id = w.NextId()
			if id == 0 {
				time.Sleep(time.Millisecond)
				continue
			}
			break
		}
		idNumList = append(idNumList, id)
	}
	return idNumList
}
