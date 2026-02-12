package templates_test

//import (
//	"fmt"
//	"github.com/bilibili/gengine/builder"
//	"github.com/bilibili/gengine/context"
//	"github.com/bilibili/gengine/engine"
//	"github.com/sirupsen/logrus"
//	"log"
//	"testing"
//	"time"
//)
//
//var riskRuleString = `
//// ======================
//// 风控规则集 - Gengine v1.5.7
//// ======================
//
//rule "黑名单校验" salience 100
//begin
//	riskLevel = "拒绝"
//	reason = "用户/手机号在黑名单"
//	isAllow = false
//end
//
//rule "单日交易次数超限" salience 90
//begin
//	if isAllow == true && dailyTradeCount > maxDailyCount {
//		riskLevel = "拒绝"
//		reason = "单日交易次数超限"
//		isAllow = false
//	}
//end
//`
//
//var riskRuleString2 = `
//rule "rulename" "rule-describtion" salience  10
//begin
//		if 7 == User.GetNum(7){
//			User.Age = User.GetNum(89767) + 10000000
//			User.Print("6666")
//		}else{
//			User.Name = "yyyy"
//		}
//end`
//
//// 定义想要注入的结构体
//type User struct {
//	Name string
//	Age  int64
//	Male bool
//}
//
//func (u *User) GetNum(i int64) int64 {
//	return i
//}
//
//func (u *User) Print(s string) {
//	fmt.Println(s)
//}
//
//func (u *User) Say() {
//	fmt.Println("hello world")
//}
//func TestTemplates3(t *testing.T) {
//	user := &User{
//		Name: "Calo",
//		Age:  0,
//		Male: true,
//	}
//
//	dataContext := context.NewDataContext()
//	//注入初始化的结构体
//	dataContext.Add("User", user)
//
//	//init rule engine
//	ruleBuilder := builder.NewRuleBuilder(dataContext)
//
//	start1 := time.Now().UnixNano()
//	//构建规则
//	err := ruleBuilder.BuildRuleFromString(riskRuleString2) //string(bs)
//	end1 := time.Now().UnixNano()
//
//	logrus.Infof("rules num:%d, load rules cost time:%d", len(ruleBuilder.Kc.RuleEntities), end1-start1)
//
//	if err != nil {
//		logrus.Errorf("err:%s ", err)
//	} else {
//		eng := engine.NewGengine()
//
//		start := time.Now().UnixNano()
//		//执行规则
//		err := eng.Execute(ruleBuilder, true)
//		println(user.Age)
//		end := time.Now().UnixNano()
//		if err != nil {
//			logrus.Errorf("execute rule error: %v", err)
//		}
//		logrus.Infof("execute rule cost %d ns", end-start)
//		logrus.Infof("user.Age=%d,Name=%s,Male=%t", user.Age, user.Name, user.Male)
//	}
//}
//
//func TestTemplates2(t *testing.T) {
//	//// 1. 读取规则
//	//ruleFile, err := os.ReadFile("risk.rule")
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//
//	// 2. 构建规则
//	dc := context.NewDataContext()
//
//	// 3. 输入参数（真实业务从RPC/HTTP/DB取）
//	data := map[string]interface{}{
//		// 基础信息
//		"userId":   "u123456",
//		"phone":    "13800138000",
//		"amount":   800.0,
//		"userDays": 3,
//
//		// 风控阈值
//		"maxDailyCount":    5,
//		"dailyTradeCount":  6,
//		"maxDailyAmount":   5000.0,
//		"dailyAmount":      4800.0,
//		"singleMaxAmount":  1000.0,
//		"newUserSingleMax": 500.0,
//
//		// 行为
//		"10mTradeCount": 8,
//		"lastCity":      "北京",
//		"currentCity":   "上海",
//		"isNewDevice":   true,
//
//		// 黑名单
//		"blackList":      []string{"u999999"},
//		"blackPhoneList": []string{"13888888888"},
//
//		"isAllow": true,
//	}
//
//	for k, v := range data {
//		dc.Add(k, v)
//	}
//
//	rb := builder.NewRuleBuilder(dc)
//	if err := rb.BuildRuleFromString(riskRuleString); err != nil {
//		log.Fatalf("规则解析失败: %v", err)
//	}
//
//	// 4. 执行引擎
//	eng := engine.NewGengine()
//	if err := eng.Execute(rb, true); err != nil {
//		log.Fatalf("执行失败: %v", err)
//	}
//
//	data2, _ := eng.GetRulesResultMap()
//
//	// 5. 输出结果
//	log.Println("是否允许:", data2["isAllow"])
//	log.Println("风险等级:", data2["riskLevel"])
//	log.Println("原因:", data2["reason"])
//}
