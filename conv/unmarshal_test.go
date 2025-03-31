package conv_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/shopspring/decimal"
	"testing"
)

type OrderListOutput struct {
	DisburseAmount decimal.Decimal `json:"disburse_amount"` // 转账金额
	BankBrance     *string         `json:"bank_brance"`     // 分行名称
}

func TestUnmarshal(t *testing.T) {
	var aa = &OrderListOutput{}
	aa.DisburseAmount = decimal.New(-34567, -2)
	dst := new(OrderListOutput)
	err := conv.Unmarshal(aa, dst)
	fmt.Println(err, dst)
}
