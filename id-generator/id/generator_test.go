package id_test

import (
	"fmt"
	id2 "github.com/magic-lib/go-plat-utils/v2/id-generator/id"
	"testing"
)

func TestGeneratorBase32(t *testing.T) {
	aa := id2.GeneratorBase32()
	fmt.Println(aa)
	aa = id2.XId()
	fmt.Println(aa)
}
