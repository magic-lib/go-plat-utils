package loadbalance_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/loadbalance"
)

// 权重全为 0 时等概率选择，不能恒返回零值
func TestWeightedRoundAllZero(t *testing.T) {
	s, err := loadbalance.NewWeightedRoundSelector([]string{"a", "b"}, []uint{0, 0})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		seen[s.Select()] = true
	}
	if seen[""] || !seen["a"] || !seen["b"] {
		t.Fatalf("选择结果 = %v", seen)
	}
}
