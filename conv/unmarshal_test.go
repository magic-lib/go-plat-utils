package conv_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
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

type AA struct {
	CreateTime time.Time `json:"create_time"`
}
type BB struct {
	CreateTime *timestamppb.Timestamp `json:"create_time"`
}

func TestToTime(t *testing.T) {
	//now := time.Now()

	//bb := BB{
	//	CreateTime: timestamppb.New(time.Now()),
	//}
	//aa := new(AA)
	//
	//_ = conv.Unmarshal(bb, aa)
	//
	//fmt.Println(conv.String(aa))

	aa := AA{
		CreateTime: time.Now(),
	}
	bb := new(BB)

	//createTime := timestamppb.New(time.Time{})
	//aa := createTime.AsTime()
	//fmt.Println(conv.String(aa))

	_ = conv.Unmarshal(aa, bb)

	fmt.Println(conv.String(bb.CreateTime.AsTime()))

	//kk, _ := conv.Time(createTime)
	//fmt.Println(conv.String(kk))

}
