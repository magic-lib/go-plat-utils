package conv_test

import (
	"database/sql"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/samber/lo"
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

type AAA struct {
	CreateTime string `json:"create_time"`
}
type BBB struct {
	CreateTime map[string]any `json:"create_time"`
}

func TestUnmarshalList(t *testing.T) {
	//aa := AAA{
	//	CreateTime: `{"tag_names":["M6","K1","N"]}`,
	//}
	//bb := new(BBB)
	//_ = conv.Unmarshal(aa, bb)
	//fmt.Println(bb)

	aaaList := []*AAA{
		{
			CreateTime: `{"tag_names":["M6","K1","N"]}`,
		},
	}
	bbbList := make([]*BBB, 0)
	_ = conv.Unmarshal(aaaList, &bbbList)
	fmt.Println(conv.String(bbbList))

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
func TestConvert(t *testing.T) {
	aa, ok := conv.Convert[int](55)
	fmt.Println(aa, ok)
	bb, ok := conv.Convert[string](55)
	fmt.Println(bb, ok)
	cc, ok := conv.Convert[int64](55)
	fmt.Println(cc, ok)
}
func TestConvert111(t *testing.T) {
	var targetPtrValue int
	err := conv.AssignTo(true, &targetPtrValue)

	fmt.Println(err, targetPtrValue)
}

type dbType struct {
	//Name        sql.NullString `json:"name"`
	//Age         sql.NullBool   `json:"age"`
	CreatorTime sql.NullTime `db:"creator_time"`
}
type jsonType struct {
	//Name        string    `json:"name"`
	//Age         bool      `json:"age"`
	CreatorTime time.Time `json:"creator_time"`
}

func TestConvertSql(t *testing.T) {
	oldList := []*dbType{
		{
			//Name: sql.NullString{
			//	String: "123",
			//	Valid:  true,
			//},
			//Age: sql.NullBool{
			//	Bool:  true,
			//	Valid: true,
			//},
			CreatorTime: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
		},
	}

	aaa, _ := conv.Time(oldList[0].CreatorTime)
	fmt.Println(aaa)

	newData := make([]*jsonType, 0)

	lo.ForEach(oldList, func(item *dbType, _ int) {
		newTemp := new(jsonType)

		_ = conv.Unmarshal(item, newTemp)

		newData = append(newData, newTemp)
	})

	fmt.Println(conv.String(newData))
}
func TestConvertList(t *testing.T) {
	ruleIdList := make([]int64, 0)
	_ = conv.Unmarshal(utils.Split("1,2,3", []string{","}), &ruleIdList)

	fmt.Println(ruleIdList)
}
