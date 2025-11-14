package state

import "fmt"

// BitSwitch 二进制开关
type BitSwitch uint64

// NewBitSwitch 初始化二进制开关
func NewBitSwitch() BitSwitch {
	return 0
}

func (bs *BitSwitch) checkPos(pos int) error {
	if pos < 0 || pos >= 64 {
		return fmt.Errorf("switch position %d out of range (0~63)", pos)
	}
	return nil
}

// TurnOn 打开指定位置的开关（pos：开关编号，0~63）
func (bs *BitSwitch) TurnOn(pos int) error {
	if err := bs.checkPos(pos); err != nil {
		return err
	}

	*bs |= 1 << pos // 对应位设为 1
	return nil
}

// TurnOff 关闭指定位置的开关（pos：开关编号，0~63）
func (bs *BitSwitch) TurnOff(pos int) error {
	if err := bs.checkPos(pos); err != nil {
		return err
	}
	*bs &= ^(1 << pos) // 对应位设为 0
	return nil
}

// Toggle 切换指定位置的开关状态（开→关，关→开）
func (bs *BitSwitch) Toggle(pos int) error {
	if err := bs.checkPos(pos); err != nil {
		return err
	}
	*bs ^= 1 << pos // 对应位翻转
	return nil
}

// IsOn 查询指定位置的开关是否打开
func (bs *BitSwitch) IsOn(pos int) (bool, error) {
	if err := bs.checkPos(pos); err != nil {
		return false, err
	}
	return (*bs & (1 << pos)) != 0, nil
}

// GetAllStates 获取所有开关状态（返回 map：开关编号→是否打开）
func (bs *BitSwitch) GetAllStates() map[int]bool {
	states := make(map[int]bool, 64)
	for pos := 0; pos < 64; pos++ {
		states[pos], _ = bs.IsOn(pos)
	}
	return states
}

// SetAllStates 批量设置所有开关状态（通过 map：开关编号→是否打开）
func (bs *BitSwitch) SetAllStates(states map[int]bool) error {
	// 先重置所有开关为关闭
	*bs = 0
	for pos, isOn := range states {
		if isOn {
			if err := bs.TurnOn(pos); err != nil {
				return err
			}
		} else {
			if err := bs.TurnOff(pos); err != nil {
				return err
			}
		}
	}
	return nil
}

// ToUint64 转换为 uint64（用于存储/传输）
func (bs *BitSwitch) ToUint64() uint64 {
	return uint64(*bs)
}

// FromUint64 从 uint64 恢复开关状态
func (bs *BitSwitch) FromUint64(val uint64) {
	*bs = BitSwitch(val)
}

// String 格式化输出（二进制字符串，便于调试）
func (bs *BitSwitch) String() string {
	return fmt.Sprintf("0b%064b", *bs)
}
