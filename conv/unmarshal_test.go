package conv_test

import (
	"database/sql"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"testing"
	"time"
)

type OrderListOutput struct {
	//DisburseAmount decimal.Decimal `json:"disburse_amount"` // 转账金额
	//BankBrance     *string         `json:"bank_brance"`     // 分行名称
	Type int `json:"type"` // 分行名称
}

type AAA struct {
	CreateTime string `json:"create_time"`
}
type BBB struct {
	CreateTime map[string]any `json:"create_time"`
}

type AA struct {
	CreateTime time.Time `json:"create_time"`
}
type BB struct {
	CreateTime *timestamppb.Timestamp `json:"create_time"`
}

type dbType struct {
	Name        sql.NullString `json:"name"`
	Age         sql.NullBool   `json:"age"`
	CreatorTime sql.NullTime   `db:"creator_time"`
}
type jsonType struct {
	Name        string    `json:"name"`
	Age         bool      `json:"age"`
	CreatorTime time.Time `json:"creator_time"`
}

func TestUnmarshal(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"map to struct", []any{
			map[string]any{
				"type": "1",
			}}, []any{true}, func(value any) bool {
			dst := new(OrderListOutput)
			_ = conv.Unmarshal(value, dst)
			if dst.Type == 1 {
				return true
			}
			return false
		}},
		{"list to list", []any{
			[]*AAA{
				{
					CreateTime: `{"tag_names":["M6","K1","N"]}`,
				},
			}}, []any{true}, func(value any) bool {
			bbbList := make([]*BBB, 0)
			if aaaList, ok := value.([]*AAA); ok {
				_ = conv.Unmarshal(aaaList, &bbbList)
				if len(bbbList) == 1 && len(bbbList[0].CreateTime) > 0 {
					log.Println(conv.String(bbbList))
					return true
				}
			}

			return false
		}},
		{"time.Time to timestamppb.Timestamp", []any{
			AA{
				CreateTime: time.Now(),
			}}, []any{true}, func(value any) bool {
			bb := new(BB)
			if aaaList, ok := value.(AA); ok {
				_ = conv.Unmarshal(aaaList, bb)
				if bb.CreateTime != nil {
					log.Println(conv.String(bb.CreateTime.AsTime()))
					return true
				}
			}
			return false
		}},
		{"bool to int", []any{
			true}, []any{true}, func(value any) bool {
			var targetPtrValue int
			if aaaList, ok := value.(bool); ok {
				_ = conv.Unmarshal(aaaList, &targetPtrValue)
				if targetPtrValue == 1 {
					return true
				}
			}
			return false
		}},
		{"string list to int list", []any{
			utils.Split("1,2,3", []string{","})}, []any{true}, func(value any) bool {
			ruleIdList := make([]int64, 0)
			if aaaList, ok := value.([]string); ok {
				_ = conv.Unmarshal(aaaList, &ruleIdList)
				if len(ruleIdList) == 3 {
					return true
				}
			}
			return false
		}},
		{"sqlNull list to golang list", []any{
			[]*dbType{
				{
					Name: sql.NullString{
						String: "123",
						Valid:  true,
					},
					Age: sql.NullBool{
						Bool:  true,
						Valid: true,
					},
					CreatorTime: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				},
			},
		}, []any{true}, func(value any) bool {
			newData := make([]*jsonType, 0)

			_ = conv.Unmarshal(value, &newData)

			if len(newData) == 1 {
				if newData[0].Name == "123" && newData[0].Age == true {
					log.Println(conv.String(newData[0]))
					return true
				}
			}
			return false
		}},
	}
	utils.TestFunction(t, testCases, nil)
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

	//createTime := timestamppb.New(time.Time{})
	//aa := createTime.AsTime()
	//fmt.Println(conv.String(aa))

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

type AuditRule struct {
	RuleId             int64     `json:"rule_id"`                     // 规则ID
	MainId             int64     `json:"main_id" validate:"required"` // 所属主体id
	RuleName           string    `json:"rule_name"`                   // 规则名称
	RiskWeight         float64   `json:"risk_weight"`                 // 风险权重
	StartTime          time.Time `json:"start_time"`                  // 开始时间
	EndTime            time.Time `json:"end_time"`                    // 到期时间
	ConditionType      string    `json:"condition_type"`              // 条件的执行：AND 满足所有条件  OR 满足以下任意条件
	ConditionAllString string    `json:"condition_all_string"`        // 所有的条件表达式
	RuleNumber         int64     `json:"rule_number"`                 // 该规则命中了多少数
	Creator            string    `json:"creator,omitempty"`           // 创建者
	CreatedAt          time.Time `json:"created_at,omitempty"`        // 创建时间
	UpdatedAt          time.Time `json:"updated_at,omitempty"`        // 更新时间
}

type JsonAuditRule struct {
	RuleId             int64          `json:"rule_id"`              // 规则ID
	RuleName           string         `json:"rule_name"`            // 规则名称
	ConditionType      string         `json:"condition_type"`       // 条件的执行：AND 满足所有条件  OR 满足以下任意条件
	ConditionAllString string         `json:"condition_all_string"` // 所有的条件表达式
	ParamDefault       map[string]any `json:"param_default"`        // 默认参数，如果没传，则用这个默认
	ParamReplace       map[string]any `json:"param_replace"`        // 覆盖参数，不管传没传，全都用这个
	IsReturn           bool           `json:"is_return"`
}

func TestConvertList2(t *testing.T) {
	ruleInfoList := []*AuditRule{
		{
			RuleId:             1,
			RuleName:           "1",
			ConditionType:      "1",
			ConditionAllString: "1",
		},
		{
			RuleId:             2,
			RuleName:           "1",
			ConditionType:      "1",
			ConditionAllString: "1",
		},
	}
	jsonRuleList := make([]*JsonAuditRule, 0)
	_ = conv.Unmarshal(ruleInfoList, &jsonRuleList)

	fmt.Println(jsonRuleList)
}
