package id

import (
	"fmt"
	"sync/atomic"
	"time"
)

const (
	// milliTimeCodeLen MilliTimeCode 的长度：14 位时间戳（yyyyMMddHHmmss）+ 3 位毫秒
	milliTimeCodeLen = 17
	// milliSeqDigits MilliTimeId 里自增序号的位数
	milliSeqDigits = 4
	milliSeqMod    = 10000 // 10 ^ milliSeqDigits，序号 10000 一循环
	// MilliTimeIdLen MilliTimeId 的长度：17 位时间码 + 4 位序号
	MilliTimeIdLen = milliTimeCodeLen + milliSeqDigits
)

var (
	milliSeq uint64 // MilliTimeId 的进程级自增序号
)

// MilliTimeCode 生成毫秒级时间码：14 位时间戳（yyyyMMddHHmmss）+ 3 位毫秒，共 17 位。
// 它只是「可读的时间码」，**不保证唯一**：同一毫秒内（并发调用、多机调用）会得到完全相同的串。
// 需要唯一 ID 请用 MilliTimeId；需要严格分布式唯一且不要求可读，请用 id.Generator() 或 snowflake。
// 注意：不要用 Format("...05.999")，它会去掉小数末尾的 0（920 -> ".92"、000 -> 小数点整个消失）。
func MilliTimeCode() string {
	return formatMilliTime(time.Now())
}

// MilliTimeId 生成 21 位纯数字的毫秒级唯一 ID：
//
//	yyyyMMddHHmmssSSS + SSSS
//	└──── 17 位时间码 ────┘   序号
//
// 用以解决 MilliTimeCode「同一毫秒内生成出相同 ID」的问题：
//   - 4 位序号用 atomic 自增：同一进程内并发调用绝不重复，且不需要 sleep、不阻塞，QPS 远高于 MilliTimeCode；
//
// 限制：序号 10000 一循环，只有「同一毫秒内调用超过 10000 次」才可能重复，正常业务遇不到。
// 若要严格全局唯一（不要求时间可读），请用 id.Generator() 或 snowflake.Worker.NextId()。
func MilliTimeId() string {
	seq := atomic.AddUint64(&milliSeq, 1) % milliSeqMod
	return fmt.Sprintf("%s%0*d", formatMilliTime(time.Now()),
		milliSeqDigits, seq)
}

// formatMilliTime 格式化为 14 位时间戳 + 3 位毫秒。
// 毫秒单独取模再补零，保证稳定 3 位（避免 .999 去尾零导致的 15~16 位结果）。
func formatMilliTime(t time.Time) string {
	return t.Format("20060102150405") + fmt.Sprintf("%03d", t.UnixMilli()%1000)
}
