package utils

import "github.com/shopspring/decimal"

// Percent 计算百分数，并保留小数点位数，后面位数直接舍去
func Percent[T1, T2 float32 | float64 | int64 | int32 | int16 | int8 | int | uint64 | uint32 | uint16 | uint8 | uint](a T1, b T2, places int) float64 {
	if b == 0 {
		return 0
	}
	var af, bf decimal.Decimal
	switch any(a).(type) {
	case float32, float64:
		af = decimal.NewFromFloat(float64(a))
	case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
		af = decimal.NewFromInt(int64(a))
	}
	switch any(b).(type) {
	case float32, float64:
		bf = decimal.NewFromFloat(float64(b))
	case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint:
		bf = decimal.NewFromInt(int64(b))
	}
	if bf.IsZero() {
		return 0
	}
	return af.Mul(decimal.NewFromInt(100)).Div(bf).RoundFloor(int32(places)).InexactFloat64()
}
