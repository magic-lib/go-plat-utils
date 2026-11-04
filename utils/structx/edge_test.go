package structx

import "testing"

type edgeA struct {
	Name   string
	Age    int
	hidden string
}

type edgeB struct {
	Name string
	Age  string // 与 edgeA 同名不同类型
}

type edgeIface interface{ Foo() }

type edgeC struct {
	Name string
	Age  int
}

func TestCopyToNonPointer(t *testing.T) {
	src := edgeC{Name: "tom", Age: 3}
	if got := CopyTo[edgeA](src); got.Name != "tom" || got.Age != 3 {
		t.Fatalf("CopyTo[struct] = %+v", got)
	}
	if p := CopyTo[*edgeA](&src); p == nil || p.Name != "tom" || p.Age != 3 {
		t.Fatalf("CopyTo[*struct] = %+v", p)
	}
}

func TestNewInstanceKinds(t *testing.T) {
	if v := NewInstance[edgeIface](); v != nil {
		t.Fatalf("接口类型应返回零值 nil")
	}
	if v := NewInstance[[3]int](); v != [3]int{} {
		t.Fatalf("数组类型应返回零值, got %v", v)
	}
	if v := NewInstance[[]int](); v == nil {
		t.Fatal("切片类型应返回非 nil 空切片")
	}
}

func TestToMapSkipUnexported(t *testing.T) {
	m := ToMap(edgeA{Name: "a", hidden: "x"})
	if _, ok := m["hidden"]; ok || m["Name"] != "a" || m["Age"] != 0 {
		t.Fatalf("ToMap = %v", m)
	}
}

func TestMergeStructsTypeMismatch(t *testing.T) {
	target := &edgeA{Name: "old", Age: 1}
	if err := MergeStructs(target, edgeB{Name: "new", Age: "x"}); err != nil {
		t.Fatalf("MergeStructs: %v", err)
	}
	if target.Name != "new" || target.Age != 1 {
		t.Fatalf("类型不同的字段应跳过, got %+v", target)
	}
}

func TestDiffFieldsEdge(t *testing.T) {
	a, b := &edgeA{Name: "a", Age: 1}, &edgeA{Name: "b", Age: 1}
	got := DiffFields(a, b, "Name", "Age", "NotExist", "hidden")
	if len(got) != 1 || got[0] != "Name" {
		t.Fatalf("DiffFields = %v", got)
	}
}
