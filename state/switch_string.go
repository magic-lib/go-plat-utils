package state

import (
	"fmt"
	"strings"
)

const (
	closeState = "0"
	openState  = "1"
)

// StringSwitch 字符串开关
type StringSwitch struct {
	switchStr string
}

// NewStringSwitch 初始化字符串开关
func NewStringSwitch(initSize int) *StringSwitch {
	if initSize <= 0 {
		initSize = 8
	}
	return &StringSwitch{
		switchStr: strings.Repeat(closeState, initSize),
	}
}

// expandIfNeeded 扩展字符串长度
func (ss *StringSwitch) expandIfNeeded(pos int) {
	if pos >= len(ss.switchStr) {
		ss.switchStr += strings.Repeat(closeState, pos-len(ss.switchStr)+1)
	}
}

// TurnOn 打开指定位置的开关
func (ss *StringSwitch) TurnOn(pos int) error {
	if pos < 0 {
		return fmt.Errorf("switch position %d must be non-negative", pos)
	}
	ss.expandIfNeeded(pos)
	// 替换对应位置的字符为 '1'
	runes := []rune(ss.switchStr)
	runes[pos] = rune(openState[0])
	ss.switchStr = string(runes)
	return nil
}

// TurnOff 关闭指定位置的开关
func (ss *StringSwitch) TurnOff(pos int) error {
	if pos < 0 {
		return fmt.Errorf("switch position %d must be non-negative", pos)
	}
	ss.expandIfNeeded(pos)
	runes := []rune(ss.switchStr)
	runes[pos] = rune(closeState[0])
	ss.switchStr = string(runes)
	return nil
}

// Toggle 切换指定位置的开关状态
func (ss *StringSwitch) Toggle(pos int) error {
	if pos < 0 {
		return fmt.Errorf("switch position %d must be non-negative", pos)
	}
	ss.expandIfNeeded(pos)
	runes := []rune(ss.switchStr)
	if runes[pos] == rune(openState[0]) {
		runes[pos] = rune(closeState[0])
	} else {
		runes[pos] = rune(openState[0])
	}
	ss.switchStr = string(runes)
	return nil
}

// IsOn 查询指定位置的开关是否打开
func (ss *StringSwitch) IsOn(pos int) (bool, error) {
	if pos < 0 {
		return false, fmt.Errorf("switch position %d must be non-negative", pos)
	}
	if pos >= len(ss.switchStr) {
		return false, nil // 未初始化的开关默认关闭
	}
	return ss.switchStr[pos] == openState[0], nil
}

// GetAllStates 获取所有开关状态（开关编号→是否打开）
func (ss *StringSwitch) GetAllStates() map[int]bool {
	states := make(map[int]bool, len(ss.switchStr))
	for pos, c := range ss.switchStr {
		states[pos] = c == rune(openState[0])
	}
	return states
}

// SetAllStates 批量设置所有开关状态
func (ss *StringSwitch) SetAllStates(states map[int]bool) error {
	// 找到最大开关编号，确定字符串长度
	maxPos := -1
	for pos := range states {
		if pos > maxPos {
			maxPos = pos
		}
	}
	if maxPos < 0 {
		ss.switchStr = ""
		return nil
	}
	// 初始化字符串为全关
	ss.switchStr = strings.Repeat(closeState, maxPos+1)
	runes := []rune(ss.switchStr)
	for pos, isOn := range states {
		if isOn {
			runes[pos] = rune(openState[0])
		} else {
			runes[pos] = rune(closeState[0])
		}
	}
	ss.switchStr = string(runes)
	return nil
}

// FromString 从字符串恢复开关状态
func (ss *StringSwitch) FromString(s string) error {
	// 校验字符串合法性（仅允许 '0' 和 '1'）
	for _, c := range s {
		if c != rune(closeState[0]) && c != rune(openState[0]) {
			return fmt.Errorf("invalid character '%c' in switch string", c)
		}
	}
	ss.switchStr = s
	return nil
}

// String 格式化输出
func (ss *StringSwitch) String() string {
	return ss.switchStr
}
