package utils

import (
	"sort"
	"sync"
)

// Hash集合数据结构
type hashSortSet map[any]int

// add 往集合添加值
func (h hashSortSet) add(value any, sortNum int) bool {
	if _, ok := h[value]; !ok {
		h[value] = sortNum
		return true
	}
	return false
}

// delete 往集合删除值
func (h hashSortSet) delete(value any) bool {
	if _, ok := h[value]; ok {
		delete(h, value)
		return true
	}
	return false
}

// contains 值是否存在集合中
func (h hashSortSet) contains(value any) bool {
	_, ok := h[value]
	return ok
}

// copy 复制hashSet
func (h hashSortSet) copy() hashSortSet {
	newSet := make(map[any]int, len(h))
	for k, v := range h {
		newSet[k] = v
	}
	return newSet
}

// NewSortSet 创建协程安全的HashSet
func NewSortSet() *syncHashSet {
	return &syncHashSet{
		values: hashSortSet{},
	}
}

// 协程安全的HashSet
type syncHashSet struct {
	values  hashSortSet
	sortNum int // 按顺序输出
	mutex   sync.Mutex
}

// Add 往set添加元素
func (s *syncHashSet) Add(value any) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sortNum = s.sortNum + 1
	return s.values.add(value, s.sortNum)
}

// Delete 删除元素
func (s *syncHashSet) Delete(value any) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.values.delete(value)
}

// Contains 检查元素存在性
func (s *syncHashSet) Contains(value any) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.values.contains(value)
}

// List 返回列表的元素
func (s *syncHashSet) List() []any {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	list := s.values.copy()

	listTemp := make([]any, 0)

	listKey := make([]int, 0)
	listOld := make(map[int]any)
	for one, sortTemp := range list {
		temp1 := sortTemp
		temp2 := one
		listKey = append(listKey, temp1)
		listOld[temp1] = temp2
	}
	sort.Ints(listKey)
	for _, one := range listKey {
		if temp, ok := listOld[one]; ok {
			listTemp = append(listTemp, temp)
		}
	}
	return listTemp
}
