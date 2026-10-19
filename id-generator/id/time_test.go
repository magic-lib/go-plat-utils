package id_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/magic-lib/go-plat-utils/id-generator/id"
)

// isAllDigit 判断字符串是否全部由数字组成
func isAllDigit(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

// TestMilliTimeCode 循环 1000 次生成时间码，把长度不等于 17 的串打印出来
func TestMilliTimeCode(t *testing.T) {
	const total = 1000
	const wantLen = 17 // 14 位时间戳 + 3 位毫秒

	bad := 0
	for i := 0; i < total; i++ {
		code := id.MilliTimeCode()
		fmt.Println(code)
		if len(code) != wantLen {
			bad++
			fmt.Printf("第 %d 次：长度=%d，期望 %d，值=%q\n", i, len(code), wantLen, code)
		}
	}

	fmt.Printf("共 %d 次，长度不为 %d 的有 %d 次\n", total, wantLen, bad)
	if bad > 0 {
		t.Errorf("MilliTimeCode 有 %d/%d 次结果不是 %d 位，详见上面的输出", bad, total, wantLen)
	}
}

// TestMilliTimeCodeConcurrentDup 验证 MilliTimeCode 的重复问题：
// 并发 1000 次调用会大量撞号（它只是时间码，不保证唯一），这里只打印不判失败
func TestMilliTimeCodeConcurrentDup(t *testing.T) {
	const total = 1000

	ids := make([]string, total)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v := id.MilliTimeCode()
			mu.Lock()
			ids[i] = v
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	seen := make(map[string]int, total)
	for _, v := range ids {
		seen[v]++
	}
	dup := total - len(seen)
	fmt.Printf("MilliTimeCode 并发 %d 次：不同值 %d 个，重复 %d 个\n", total, len(seen), dup)
	if dup > 0 {
		fmt.Println("  -> 这正是同一毫秒内撞号，唯一场景请改用 MilliTimeId")
	}
}

// TestMilliTimeId 验证 MilliTimeId：长度固定为 MilliTimeIdLen(26) 位且全是数字，并发 1000 次不重复
func TestMilliTimeId(t *testing.T) {
	const total = 1000

	// 长度与格式
	code := id.MilliTimeId()
	if len(code) != id.MilliTimeIdLen {
		t.Errorf("MilliTimeId 长度期望 %d，实际 %d（%q）", id.MilliTimeIdLen, len(code), code)
	}
	if !isAllDigit(code) {
		t.Errorf("MilliTimeId 期望全是数字，实际 %q", code)
	}
	// 后 5 位是机器码，同一进程内应当固定
	if tail := code[len(code)-5:]; tail != id.MilliTimeId()[id.MilliTimeIdLen-5:] {
		t.Errorf("同一进程内机器码应当固定：%q", tail)
	}

	// 并发唯一性
	ids := make([]string, total)
	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ids[i] = id.MilliTimeId()
		}(i)
	}
	wg.Wait()

	seen := make(map[string]int, total)
	for _, v := range ids {
		if len(v) != id.MilliTimeIdLen {
			t.Errorf("长度异常：%q", v)
		} else {
			fmt.Println(v)
		}
		seen[v]++
	}
	dup := 0
	for v, c := range seen {
		if c > 1 {
			dup += c - 1
			fmt.Printf("重复 %d 次：%q\n", c, v)
		}
	}
	fmt.Printf("MilliTimeId 并发 %d 次：不同值 %d 个，重复 %d 个\n", total, len(seen), dup)
	if dup > 0 {
		t.Errorf("MilliTimeId 并发出现 %d 个重复", dup)
	}
}
