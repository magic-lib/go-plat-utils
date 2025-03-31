package utils_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"strconv"
	"strings"
	"testing"
)

type foo struct {
	bar string
}

func (f foo) Clone() foo {
	return foo{f.bar}
}

/*
%d：十进制整数
%f：浮点数
%s：字符串
%t：布尔值
%v：值的默认格式
%T：值的类型
%p：指针地址
%c：字符
%b：二进制
%o：八进制
%x：十六进制（小写字母）
%X：十六进制（大写字母）
%e：科学计数法（小写字母）
%E：科学计数法（大写字母）
%g：根据情况自动选择 %f 或 %e
%G：根据情况自动选择 %f 或 %E
%%：百分号
*/
func TestIO(t *testing.T) {
	caught := false

	lo.TryCatchWithErrorValue(func() error {
		panic("error")
		return nil
	}, func(val any) {
		caught = val == "error"
	})

	fmt.Println(caught)

	even := lo.Filter([]int{1, 2, 3, 4}, func(x int, index int) bool {
		return x%2 == 0
	})
	//[2,4]
	fmt.Println(conv.String(even))

	arr := lo.Map([]int64{1, 2, 3, 4}, func(x int64, index int) string {
		return strconv.FormatInt(x, 10)
	})
	fmt.Println(conv.String(arr))
	// []string{"1", "2", "3", "4"}

	matching := lo.FilterMap([]string{"cpu", "gpu", "mouse", "keyboard"}, func(x string, _ int) (string, bool) {
		if strings.HasSuffix(x, "pu") {
			return "xpu", true
		}
		return "", false
	})
	fmt.Println(conv.String(matching))
	// []string{"xpu", "xpu"}

	mm := lo.FlatMap([]int64{0, 1, 2, 43}, func(x int64, _ int) []string {
		return []string{
			strconv.FormatInt(x, 16),
			strconv.FormatInt(x, 10),
		}
	})
	fmt.Println(conv.String(mm))

	sum := lo.Reduce([]int{1, 2, 3, 4}, func(agg int, item int, _ int) int {
		return agg + item
	}, 0)
	fmt.Println(sum)

	result := lo.ReduceRight([][]int{{0, 1}, {2, 3}, {4, 5}}, func(agg []int, item []int, _ int) []int {
		return append(agg, item...)
	}, []int{})
	fmt.Println(conv.String(result))

	result1 := make([]string, 0)
	lo.ForEach([]string{"hello", "world"}, func(x string, _ int) {
		result1 = append(result1, x)
	})
	fmt.Println(conv.String(result1))

	lop.ForEach([]string{"hello", "world"}, func(x string, _ int) {
		println(x)
	})
	// prints "hello\nworld\n" or "world\nhello\n"

	list := []int64{1, 2, -42, 4}

	lo.ForEachWhile(list, func(x int64, _ int) bool {
		if x < 0 {
			return false
		}
		fmt.Println(x)
		return true
	})
	// 1
	// 2

	lo.Times(3, func(i int) string {
		return strconv.FormatInt(int64(i), 10)
	})
	// []string{"0", "1", "2"}

	lop.Times(3, func(i int) string {
		return strconv.FormatInt(int64(i), 10)
	})
	// []string{"0", "1", "2"}

	lo.Uniq([]int{1, 2, 2, 1})

	lo.UniqBy([]int{0, 1, 2, 3, 4, 5}, func(i int) int {
		return i % 3
	})
	// []int{0, 1, 2}

	lo.GroupBy([]int{0, 1, 2, 3, 4, 5}, func(i int) int {
		return i % 3
	})
	// map[int][]int{0: []int{0, 3}, 1: []int{1, 4}, 2: []int{2, 5}}

	lop.GroupBy([]int{0, 1, 2, 3, 4, 5}, func(i int) int {
		return i % 3
	})
	// map[int][]int{0: []int{0, 3}, 1: []int{1, 4}, 2: []int{2, 5}}

	lo.Chunk([]int{0, 1, 2, 3, 4, 5}, 2)
	// [][]int{{0, 1}, {2, 3}, {4, 5}}

	lo.Chunk([]int{0, 1, 2, 3, 4, 5, 6}, 2)
	// [][]int{{0, 1}, {2, 3}, {4, 5}, {6}}

	lo.Chunk([]int{}, 2)
	// [][]int{}

	lo.Chunk([]int{0}, 2)
	// [][]int{{0}}

	lo.PartitionBy([]int{-2, -1, 0, 1, 2, 3, 4, 5}, func(x int) string {
		if x < 0 {
			return "negative"
		} else if x%2 == 0 {
			return "even"
		}
		return "odd"
	})
	// [][]int{{-2, -1}, {0, 2, 4}, {1, 3, 5}}

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	// 使用 GroupBy 根据奇偶性分组
	groupedByParity := lo.GroupBy(numbers, func(x int) string {
		if x%2 == 0 {
			return "even"
		}
		return "odd"
	})
	fmt.Println("GroupBy Parity:", groupedByParity)

	// 使用 PartitionBy 根据奇偶性分割
	partitionedByParity := lo.PartitionBy(numbers, func(x int) bool {
		return x%2 == 0
	})
	fmt.Println("PartitionBy Parity:", partitionedByParity)

	//GroupBy Parity: map[even:[2 4 6 8] odd:[1 3 5 7 9]]
	//PartitionBy Parity: [[1 3 5 7 9] [2 4 6 8]]

	lo.Flatten([][]int{{0, 1}, {2, 3, 4, 5}})
	// []int{0, 1, 2, 3, 4, 5}

	lo.Interleave([]int{1, 4, 7}, []int{2, 5, 8}, []int{3, 6, 9})
	// []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	lo.Interleave([]int{1}, []int{2, 5, 8}, []int{3, 6}, []int{4, 7, 9, 10})
	// []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	lo.Shuffle([]int{0, 1, 2, 3, 4, 5}) //随机打乱
	// []int{1, 4, 0, 3, 5, 2}

	lo.Reverse([]int{0, 1, 2, 3, 4, 5})
	// []int{5, 4, 3, 2, 1, 0}

	lo.Fill([]foo{foo{"a"}, foo{"a"}}, foo{"b"})
	// []foo{foo{"b"}, foo{"b"}}

	lo.Repeat(2, foo{"a"}) // 与 Fill 功能一样

	in := []int{0, 1, 2, 3, 4}
	lo.Slice(in, 0, 5)

	lo.ValueOr(map[string]int{"foo": 1, "bar": 2}, "foo", 42)
	// 1
	lo.ValueOr(map[string]int{"foo": 1, "bar": 2}, "baz", 42)
	// 42

	lo.PickBy(map[string]int{"foo": 1, "bar": 2, "baz": 3}, func(key string, value int) bool {
		return value%2 == 1
	})
	// map[string]int{"foo": 1, "baz": 3}

	lo.Invert(map[string]int{"a": 1, "b": 2, "c": 1})
	// map[int]string{1: "c", 2: "b"}

	m := map[int]int64{1: 4, 2: 5, 3: 6}
	lo.MapToSlice(m, func(k int, v int64) string {
		return fmt.Sprintf("%d_%d", k, v)
	})
	// []string{"1_4", "2_5", "3_6"}

	lo.RandomString(5, lo.LettersCharset)
	// example: "eIGbt"

	lo.Duration3(func() (string, int, error) {
		// very long job
		return "hello", 42, nil
	})
	// hello
	// 42
	// nil
	// 3s

	lo.Ternary(true, "a", "b")
	// "a"
	lo.Ternary(false, "a", "b")
	// "b"

	// Output:
	// true
	// [2,4]
	// ["1","2","3","4"]
	// ["xpu","xpu"]
	// ["0","0","1","1","2","2","2b","43"]
	// 10
	// [4,5,2,3,0,1]
	// ["hello","world"]
}
