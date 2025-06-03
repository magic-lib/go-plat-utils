package conv_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"testing"
)

type OrderListOutput struct {
	//DisburseAmount decimal.Decimal `json:"disburse_amount"` // 转账金额
	//BankBrance     *string         `json:"bank_brance"`     // 分行名称
	Type int `json:"type"` // 分行名称
}

func TestUnmarshal(t *testing.T) {
	//var aa = &OrderListOutput{}
	//aa.DisburseAmount = decimal.New(-34567, -2)
	//dst := new(OrderListOutput)
	//err := conv.Unmarshal(aa, dst)
	//fmt.Println(err, dst)

	var aa = map[string]any{
		"type": "1",
	}
	dst := new(OrderListOutput)
	err := conv.Unmarshal(aa, dst)
	fmt.Println(dst, err)
}
func TestInt(t *testing.T) {
	str := "\u0000"
	num, ok := conv.Int(str)

	fmt.Println(num, ok)
}
func TestToString(t *testing.T) {
	str := []uint8{0}
	num := conv.String(str)

	fmt.Println(num)
}
