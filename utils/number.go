package utils

import (
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/shopspring/decimal"
	"math/rand"
	"time"
)

// Percent 计算百分数，并保留小数点位数，后面位数直接舍去
func Percent[T1, T2 float32 | float64 | int64 | int32 | int16 | int8 | int | uint64 | uint32 | uint16 | uint8 | uint](a T1, b T2, places ...int) float64 {
	if b == 0 {
		return 0
	}
	af, _ := conv.Convert[decimal.Decimal](a)
	bf, _ := conv.Convert[decimal.Decimal](b)
	if bf.IsZero() {
		return 0
	}
	var onePlaces = 2
	if len(places) > 0 {
		onePlaces = places[0]
	}
	return af.Mul(decimal.NewFromInt(100)).Div(bf).RoundFloor(int32(onePlaces)).InexactFloat64()
}

func RandomIntInRange(minNum, maxNum int64) int64 {
	if minNum > maxNum {
		minNum, maxNum = maxNum, minNum
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int63n(maxNum-minNum+1) + minNum
}
func RandomInt(length int) int64 {
	if length < 1 {
		return 0
	}
	if length > 18 {
		length = 18
	}
	minNum := pow10(length - 1)
	maxNum := pow10(length) - 1
	return RandomIntInRange(minNum, maxNum)
}

func pow10(n int) int64 {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
	return int64(result)
}

// DecimalAdd 相加数字
func DecimalAdd(b ...any) decimal.Decimal {
	var af decimal.Decimal
	for _, a := range b {
		if bf, err := conv.Convert[decimal.Decimal](a); err == nil {
			af = af.Add(bf)
		}
	}
	return af
}
