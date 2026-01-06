package conv_test

import (
	"database/sql"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/magic-lib/go-plat-utils/utils/httputil"
	"github.com/shopspring/decimal"
	"golang.org/x/text/language"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	yaml "gopkg.in/yaml.v3"
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
	ValuesMap string `json:"valuesMap"`
}
type BBB struct {
	ValuesMap map[string]any `json:"valuesMap"`
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

type LogConf struct {
	// ServiceName represents the service name.
	ServiceName string `json:",optional"`
}

type ServiceConf struct {
	Name string
	Log  LogConf
}
type RestConf struct {
	ServiceConf
	Host string `json:",default=0.0.0.0"`
	Port int
}

type Config struct {
	RestConf RestConf `json:"restConf"`
	Log      *LogConf
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
		{"list to listMap", []any{
			[]*AAA{
				{
					ValuesMap: `{"tag_names":["M6","K1","N"]}`,
				},
			}}, []any{true}, func(value any) bool {
			bbbList := make([]*BBB, 0)
			if aaaList, ok := value.([]*AAA); ok {
				_ = conv.Unmarshal(aaaList, &bbbList)
				if len(bbbList) == 1 && len(bbbList[0].ValuesMap) == 1 && bbbList[0].ValuesMap["tag_names"] != nil {
					if tagNames, ok := bbbList[0].ValuesMap["tag_names"].([]any); ok {
						if len(tagNames) == 3 && conv.String(tagNames[0]) == "M6" {
							return true
						}
					}

					log.Println(conv.String(bbbList))
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
		{"timestamppb.Timestamp to Time", []any{
			AA{
				CreateTime: time.Now(),
			}}, []any{true}, func(value any) bool {
			now := time.Now()

			bb := BB{
				CreateTime: timestamppb.New(now),
			}
			aa := new(AA)

			_ = conv.Unmarshal(bb, aa)

			if aa.CreateTime.Equal(now) {
				return true
			}
			return false
		}},
		{"bool to int", []any{
			true}, []any{true}, func(aaaList bool) bool {
			var targetPtrValue int
			_ = conv.Unmarshal(aaaList, &targetPtrValue)
			if targetPtrValue == 1 {
				return true
			}
			return false
		}},
		{"string list to int list", []any{
			utils.Split("1,2,3", []string{","})}, []any{true}, func(aaaList []string) bool {
			ruleIdList := make([]int64, 0)
			_ = conv.Unmarshal(aaaList, &ruleIdList)
			if len(ruleIdList) == 3 {
				return true
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
		}, []any{true}, func(value []*dbType) bool {
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
		{"str to struct", []any{
			`{"code":5,"data":{"id":255,"mobile":"0968635986","password":"","realname":"443 4434","gender":0,"birthday":{"Time":"1995-10-31T00:00:00+02:00","Valid":true},"nrc":"334343/43/4","empno":"","province":"Eastern","district":"Lundazi","register_time":{"Time":"2025-10-31T03:10:45+02:00","Valid":true},"register_ip":"172.18.0.1","last_login_time":{"Time":"2025-10-31T03:10:45+02:00","Valid":true},"last_login_ip":"172.18.0.1","status":0,"is_delete":false,"source":"","agent_id":0,"promotion_id":255,"secend_mobile":"","check_info":"0,0,1","check_result":false,"salary_day":0,"monthly_income_grade":1,"sector":5,"created_at":{"Time":"2025-10-31T03:10:45+02:00","Valid":true},"updated_at":{"Time":"2025-10-31T03:34:30+02:00","Valid":true},"commission_amount":"K25","member_group":"","member_identity":"","member_type":"OTHER","member_credit":{"id":0,"member_id":255,"credit_limit":0,"credit_source":"OTHER","afford_ability":0,"update_time":"2025-10-31 03:10:46","create_time":"2025-10-31 03:10:46"},"member_exception":null,"contacts":[{"id":512,"member_id":255,"contact_type":"family","relation":"Brother","realname":"fdsfdsfs","mobile":"0343243242","address":"","remark":"","nrc":"","extra_properties":{"String":"","Valid":false},"created_at":{"Time":"2025-10-31T03:56:00+02:00","Valid":true},"updated_at":{"Time":"2025-10-31T04:05:20+02:00","Valid":true}},{"id":513,"member_id":255,"contact_type":"colleague","relation":"Colleague","realname":"dddddd","mobile":"0324324324","address":"","remark":"","nrc":"","extra_properties":{"String":"","Valid":false},"created_at":{"Time":"2025-10-31T03:56:00+02:00","Valid":true},"updated_at":{"Time":"2025-10-31T03:56:00+02:00","Valid":true}}],"banks":[{"id":260,"member_id":255,"wallet_type":0,"bank_type":"Mtn","bank_key":"","bank_account":"0968635986","is_default":true,"extra_properties":{"String":"","Valid":false},"created_at":{"Time":"2025-10-31T03:56:03+02:00","Valid":true},"updated_at":{"Time":"2025-10-31T03:56:02+02:00","Valid":true}}],"member_tags":[{"id":2891,"source_id":255,"tag_type":"member","tag_name":"B30","tag_source":"Reg","created_at":{"Time":"2025-10-31T04:05:57+02:00","Valid":true},"updated_at":{"Time":"2025-10-31T04:05:57+02:00","Valid":true}}],"company_list":[{"company_no":"ZB8882","company_name":"4_new_dp","nrc":"099271/48/1","mobile":"0968635986","true_name":"zhang yong","mou_salary_day":2,"creator_time":"2025-10-29T11:32:38+02:00"}]},"message":"success"}`,
		}, []any{true}, func(value string) bool {
			respInfo := new(httputil.CommResponse)

			_ = conv.Unmarshal(value, respInfo)

			if respInfo.Code == 5 {
				if respInfo.Data != nil {
					dat, ok := respInfo.Data.(map[string]any)
					if ok {
						inta, _ := conv.Convert[int](dat["id"])
						if inta == 255 {
							return true
						}
					}
				}
			}
			return false
		}},
		{"map to struct2", []any{
			`{"Log":{"ServiceName":"zamloanv1-proxy-server"},"RestConf":{"Name":"ussd-server-http","Host":"0.0.0.0","Port":10201}}`,
		}, []any{true}, func(value string) bool {
			v := new(Config)
			err := conv.Unmarshal(value, v)
			if err != nil {
				return false
			}
			if v.Log != nil && v.Log.ServiceName == "zamloanv1-proxy-server" &&
				v.RestConf.Name == "ussd-server-http" && v.RestConf.Host == "0.0.0.0" && v.RestConf.Port == 10201 {
				return true
			}

			return false
		}},
		{"struct to struct", []any{}, []any{true}, func() bool {
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

			if len(jsonRuleList) == 2 {
				if jsonRuleList[0].RuleId == 1 && jsonRuleList[0].RuleName == "1" &&
					jsonRuleList[1].RuleId == 2 && jsonRuleList[1].RuleName == "1" {
					return true
				}
			}
			return false
		}},
		{"yaml to struct", []any{}, []any{true}, func() bool {
			newConfigStr := "AppEnv : dev\nRestConf: # http服务配置\n  Name: ussd-server-http\n  Host: 0.0.0.0\n  Port: 10201\n  Timeout: 100000000000 # 10秒\n  MaxBytes: 10485760 # 设置为10MB，可根据需要调整\n\nPrefix: /ussd-server/api/v1\n\nLog: # 日志配置\n  ServiceName: zamloanv1-proxy-server\n  Mode: file    # 支持 console（控制台）或 file（文件）\n  Path: logs\n  Level: info   # 日志级别，支持 debug、info、warn、error\n  KeepDays: 3   # 日志保留天数\n  MaxSize: 100\n\napi:\n  Member:\n#    domain: http://192.168.2.84:10301\n    domain: http://182.168.2.84:10301\n    urls:\n      ReqUrlExceptionList: /member/api/v1/pendingMemberList\n      ReqUrlAgentList: /member/api/v1/memberAgentList\n      ReqUrlAgentComList: /member/api/v1/agentCommissionList\n      ReqUrlMemberUpStoreList: /member/api/v1/memberUpStoreList\n\n\n  Manager:\n    domain: http://192.168.2.84:10801\n#    domain: http://127.0.0.1:10801\n    urls:\n      ReqUrlInternalUserLogin: /manager/api/v1/internalUserLogin\n"
			var val any
			var err error
			if err = yaml.Unmarshal([]byte(newConfigStr), &val); err != nil {
				return false
			}

			v := new(Config)
			err = conv.Unmarshal(val, v)
			if err != nil {
				return false
			}

			if v.Log != nil && v.Log.ServiceName == "zamloanv1-proxy-server" &&
				v.RestConf.Name == "ussd-server-http" && v.RestConf.Host == "0.0.0.0" && v.RestConf.Port == 10201 {
				return true
			}

			return false
		}},
		{"any to decimal.Decimal", []any{}, []any{true}, func() bool {
			temp1 := new(decimal.Decimal)
			_ = conv.Unmarshal("100.05", temp1)
			temp2 := new(decimal.Decimal)
			_ = conv.Unmarshal("100", temp2)
			temp3 := new(decimal.Decimal)
			_ = conv.Unmarshal(100, temp3)
			temp4 := new(decimal.Decimal)
			_ = conv.Unmarshal(100.05, temp4)
			temp5 := new(decimal.Decimal)
			_ = conv.Unmarshal("1,000.05", temp5)

			if temp1.String() == "100.05" && temp2.String() == "100" && temp3.String() == "100" && temp4.String() == "100.05" && temp5.String() == "1000.05" {
				return true
			}

			return false
		}},
	}
	utils.TestFunction(t, testCases, nil)
}

func TestConvert(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"convert", []any{}, []any{true}, func() bool {
			aa, ok1 := conv.Convert[int](55)
			bb, ok2 := conv.Convert[string](55)
			cc, ok3 := conv.Convert[int64](55)

			if ok1 == nil && ok2 == nil && ok3 == nil {
				if aa == 55 && bb == "55" && cc == 55 {
					return true
				}
			}
			return false
		}},
		{"toInt", []any{}, []any{true}, func() bool {
			str := "\u0000"
			num, err := conv.Convert[int](str)

			if err != nil && num == 0 {
				return true
			}
			return false
		}},
		{"toString", []any{}, []any{true}, func() bool {
			str := []uint8{0}
			num := conv.String(str)

			if num == string(str) {
				return true
			}
			return false
		}},
		{"toTime", []any{}, []any{true}, func() bool {
			now := time.Now()

			createTime := timestamppb.New(now)
			aaa := createTime.AsTime()

			kk, _ := conv.Convert[time.Time](createTime)
			if kk.Equal(aaa) {
				return true
			}
			return false
		}},
	}
	utils.TestFunction(t, testCases, nil)
}
func TestFormatNumber(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"int", []any{"%d", 50000}, []any{"50000"}, conv.FormatNumber[int]},
		{"int", []any{"%d", 50000, language.English}, []any{"50,000"}, conv.FormatNumber[int]},
		{"int64", []any{"%d", 50000, language.English}, []any{"50,000"}, conv.FormatNumber[int64]},
		{"float64", []any{"%.2f", 50000.34343, language.English}, []any{"50,000.34"}, conv.FormatNumber[float64]},
	}
	utils.TestFunction(t, testCases, nil)
}
