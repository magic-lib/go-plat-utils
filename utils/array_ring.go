package utils

import (
	"container/ring"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
)

// NextByRing 从一个循环里取下一个，数组会构成一个圈
func NextByRing[K comparable, V any](
	vsList []V,
	last V,
	key func(this V) K, // 用于元素唯一标识的提取函数，判断是否相等使用
	next func(this V, last V) bool, // 如果未找到key，则用此来判断元素下一个元素的条件函数
) V {
	// 处理空切片情况
	vLen := len(vsList)
	if vLen == 0 {
		return last
	}
	// 只有一个元素时直接返回该元素
	if vLen == 1 || cond.IsNil(last) {
		return vsList[0]
	}

	idx := -1
	var nextOne V
	var foundNext bool
	lo.ForEachWhile(vsList, func(item V, index int) bool {
		if cond.IsNil(item) {
			return true
		}
		if key(item) == key(last) {
			idx = index
			return false
		}
		if !foundNext {
			// 只取第一个符合条件的元素
			if next(item, last) {
				foundNext = true
				nextOne = item
			}
		}
		return true
	})
	if idx >= 0 {
		return vsList[(idx+1)%vLen]
	}
	// 返回找到的候选元素或列表第一个元素
	if foundNext {
		return nextOne
	}

	return vsList[0]
}

type ArrayRing[K comparable, V any] struct {
	vsList  []V
	keyFunc func(this V) K
	ring    *ring.Ring
}

type ringData[K comparable, V any] struct {
	key K
	val V
}

func NewArrayRing[K comparable, V any](vsList []V, keyFunc func(this V) K) (*ArrayRing[K, V], error) {
	vsLen := len(vsList)
	if vsLen == 0 {
		return nil, fmt.Errorf("list is empty")
	}
	if keyFunc == nil {
		return nil, fmt.Errorf("keyFunc is nil")
	}
	oneRing := ring.New(len(vsList))
	for i := 0; i < vsLen; i++ {
		oneData := vsList[i]
		oneKey := keyFunc(oneData)
		oneRing.Value = &ringData[K, V]{
			key: oneKey,
			val: oneData,
		}
		oneRing = oneRing.Next()
	}
	return &ArrayRing[K, V]{
		vsList:  vsList,
		keyFunc: keyFunc,
		ring:    oneRing,
	}, nil
}

func (a *ArrayRing[K, V]) Next(curr V) V {
	currKey := a.keyFunc(curr)
	p := a.ring
	for i := 0; i < a.ring.Len(); i++ {
		oneData := p.Value.(*ringData[K, V])
		if oneData.key == currKey {
			nextData := p.Next()
			return nextData.Value.(*ringData[K, V]).val
		}
	}
	// 如果没找到有可能会数据删除了
	return NextByRing(a.vsList, curr, a.keyFunc, func(this V, last V) bool {
		thisKey := a.keyFunc(this)
		lastKey := a.keyFunc(last)
		thisKeyNum, err1 := conv.Convert[int64](thisKey)
		lastKeyNum, err2 := conv.Convert[int64](lastKey)
		if err1 == nil && err2 == nil {
			return thisKeyNum > lastKeyNum
		}

		thisKeyStr, err1 := conv.Convert[string](thisKey)
		lastKeyStr, err2 := conv.Convert[string](lastKey)
		if err1 == nil && err2 == nil {
			return thisKeyStr > lastKeyStr
		}
		return false
	})
}
