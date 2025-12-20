package utils_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestPercent(t *testing.T) {
	var a int = 3
	var b float64 = 2.5

	kk := utils.Percent(a, b, 2)
	fmt.Println(kk)
}
