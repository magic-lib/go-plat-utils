package cond

// Contains 数组中是否含有某元素
func Contains[T comparable](s []T, e T) (bool, int) {
	for i, a := range s {
		if a == e {
			return true, i
		}
	}
	return false, -1
}
