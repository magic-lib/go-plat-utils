package state_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/state"
	"testing"
)

type AA struct {
	Name string
}

func TestCacheMap(t *testing.T) {

	aa := state.NewBaseFsm[string, string]()
	err1 := aa.Register("a", "b", "c")
	err2 := aa.Register("c", "b", "d")
	err3 := aa.Register("a", "b", "d")

	fmt.Println(aa.StateMap(), err1, err2, err3)
}
