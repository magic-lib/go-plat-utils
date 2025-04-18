package utils_test

import (
	"fmt"
	"testing"
)

type AAA struct {
	Name string `json:"bName" db:"aName"`
	Age  string `json:"age"`
	Key  string `db:"key"`
	Key1 string `db:"key1" json:"bKey1"`
	Key2 string
}

func TestTag(t *testing.T) {
	err := fmt.Errorf("aaaaaa")
	mm := fmt.Errorf("fdsfs: %w", err)
	fmt.Println(mm.Error())

}
