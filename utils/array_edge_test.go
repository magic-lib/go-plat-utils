package utils_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/utils"
)

func TestArrayChunkSplitInvalidSize(t *testing.T) {
	arr := []int{1, 2, 3}
	for _, size := range []int{0, -1} {
		if got := utils.ArrayChunk(arr, size); len(got) != 1 || len(got[0]) != 3 {
			t.Errorf("ArrayChunk(arr, %d) = %v, want 整体一组", size, got)
		}
		if got := utils.ArraySplit(arr, size); len(got) != 1 || len(got[0]) != 3 {
			t.Errorf("ArraySplit(arr, %d) = %v, want 整体一组", size, got)
		}
	}
	if got := utils.ArrayChunk([]int{}, 0); len(got) != 0 {
		t.Errorf("空数组 ArrayChunk = %v", got)
	}
	if got := utils.ArrayChunk(arr, 2); len(got) != 2 || len(got[1]) != 1 {
		t.Errorf("ArrayChunk(arr, 2) = %v", got)
	}
}

func TestArrayRingNext(t *testing.T) {
	r, err := utils.NewArrayRing([]int{10, 20, 30}, func(v int) int { return v })
	if err != nil {
		t.Fatal(err)
	}
	for cur, want := range map[int]int{10: 20, 20: 30, 30: 10} {
		if got := r.Next(cur); got != want {
			t.Errorf("Next(%d) = %d, want %d", cur, got, want)
		}
	}
}
