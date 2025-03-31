package templates_test

import (
	"fmt"
	"github.com/PaesslerAG/gval"
	"github.com/magic-lib/go-plat-utils/templates"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/rulego/rulego"
	"github.com/rulego/rulego/api/types"
	"strings"
	"testing"
)

func TestTemplates(t *testing.T) {
	dataMap := map[string]interface{}{
		"code": map[string]map[string]float32{
			"numMap": map[string]float32{
				"num":  4.5,
				"num2": 4.51,
				"num3": 4.51,
			},
		},
	}

	testCases := []*utils.TestStruct{
		{"strings.Index", []any{"111/<no value>", "<no value>"}, []any{4}, strings.Index},
		{"templates.Template", []any{"{{.aaa_aaa}}", map[string]interface{}{"aaa_aaa": "55555"}}, []any{"55555"}, templates.Template},
		{"templates.Template error", []any{"{{.aaa/aaa}}", map[string]interface{}{"aaa/aaa": "55555"}}, []any{"", fmt.Errorf("template: utils-template-e5a33834c17286b3865ded9360051c98:1: bad character U+002F '/'")}, templates.Template},
		{"templates.RuleExpr", []any{"==code", "code"}, []any{false}, templates.RuleExpr},
		{"templates.RuleExpr", []any{"code.numMap.num==code.numMap.num2", dataMap}, []any{false}, templates.RuleExpr},
		{"templates.RuleExpr ==", []any{"code.numMap.num3==code.numMap.num2", dataMap}, []any{true}, templates.RuleExpr},
		{"templates.Replace", []any{"这是一个{{code.numMap.num}}数字", dataMap}, []any{"这是一个4.5数字"}, func(temp string, tempMap interface{}) (string, error) {
			one := templates.NewTemplate(temp, "{{", "}}")
			return one.Replace(tempMap)
		}},
		{"templates.Eval", []any{"1+2"}, []any{int64(3)}, func(temp string) (int64, error) {
			one, err := templates.Eval(temp)
			if err != nil {
				return 0, err
			}
			ret := one.Num()
			return ret.Int64(), nil
		}},
	}
	utils.TestFunction(t, testCases, nil)
}

var amount = 0.8
var name = "awesome"

func TestEvaluate(t *testing.T) {

	testCases := []*utils.TestStruct{
		{"Single parameter modified by constant", []any{"foo + 2", map[string]interface{}{
			"foo": 2.0,
		}}, []any{4.0}, nil},
		{"Single parameter modified by variable", []any{"foo * bar", map[string]interface{}{
			"foo": 5.0,
			"bar": 2.0,
		}}, []any{10.0}, nil},
		{"Single parameter modified by variable", []any{`foo["hey"] * bar[1]`, map[string]interface{}{
			"foo": map[string]interface{}{"hey": 5.0},
			"bar": []interface{}{7., 2.0},
		}}, []any{10.0}, nil},
		{"Multiple multiplications of the same parameter", []any{"foo * foo * foo", map[string]interface{}{
			"foo": 10.0,
		}}, []any{1000.0}, nil},
		{"NoSpaceOperator", []any{"true&&name", map[string]interface{}{
			"name": true,
		}}, []any{true}, nil},
		{"Short-circuit coalesce", []any{`"foo" ?? fail()`, gval.Function("fail", func(arguments ...interface{}) (interface{}, error) {
			return nil, fmt.Errorf("Did not short-circuit")
		})}, []any{"foo"}, nil},
		{"Mixed function and parameters", []any{`sum(1.2, amount) + name`, gval.Function("sum", func(arguments ...interface{}) (interface{}, error) {
			sum := 0.0
			for _, v := range arguments {
				sum += v.(float64)
			}
			return sum, nil
		}), map[string]interface{}{
			"amount": 0.8,
			"name":   "awesome",
		}}, []any{"2awesome"}, nil},
	}
	utils.TestFunction(t, testCases, templates.Evaluate)
}

func TestRuleGo(t *testing.T) {
	ruleFile := []byte(`{
 "ruleChain": {
   "id":"chain_call_rest_api",
   "name": "测试规则链",
   "root": true
 },
 "metadata": {
   "nodes": [
     {
       "id": "s1",
       "type": "jsFilter",
       "name": "过滤",
       "debugMode": true,
       "configuration": {
         "jsScript": "return msg!='bb';"
       }
     },
     {
       "id": "s2",
       "type": "jsTransform",
       "name": "转换",
       "debugMode": true,
       "configuration": {
         "jsScript": "metadata['test']='test02';\n metadata['index']=52;\n msgType='TEST_MSG_TYPE2';\n  msg['aa']=66; return {'msg':msg,'metadata':metadata,'msgType':msgType};"
       }
     },
     {
       "id": "s3",
       "type": "restApiCall",
       "name": "推送数据",
       "debugMode": true,
       "configuration": {
         "restEndpointUrlPattern": "http://192.168.136.26:9099/api/msg",
         "requestMethod": "POST",
         "maxParallelRequestsCount": 200
       }
     }
   ],
   "connections": [
     {
       "fromId": "s1",
       "toId": "s2",
       "type": "True"
     },
     {
       "fromId": "s2",
       "toId": "s3",
       "type": "Success"
     }
   ]
 }
}`)
	ruleEngine, _ := rulego.New("rule01", ruleFile)

	metaData := types.NewMetadata()
	metaData.PutValue("productType", "test01")

	msg := types.NewMsg(0, "TELEMETRY_MSG", types.JSON, metaData, `{"temperature":35}`)

	ruleEngine.OnMsg(msg)

}
