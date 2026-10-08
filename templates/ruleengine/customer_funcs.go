package ruleengine

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/internal/govaluate-3.0.0"
	"github.com/magic-lib/go-plat-utils/mask"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"reflect"
	"strings"
	"time"
)

// customerFunc 自定义方法列表
type customerFunc struct {
	// engLogic 关联的表达式引擎，给 Find 这类需要二次求值的函数复用函数表和表达式缓存，可能为 nil
	engLogic *EngineLogic
}

func (r *customerFunc) getAllDecimalList(args ...any) []decimal.Decimal {
	decimalList := make([]decimal.Decimal, 0)
	for _, arg := range args {
		d, err := conv.Convert[decimal.Decimal](arg)
		if err == nil {
			decimalList = append(decimalList, d)
		}
	}
	return decimalList
}

// relationByNumber 两数相互运算
func (r *customerFunc) relationByNumber(f func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal, args ...any) float64 {
	decimalList := r.getAllDecimalList(args...)
	if len(decimalList) == 0 {
		return 0
	}
	var total decimal.Decimal
	for i, d := range decimalList {
		if i == 0 {
			total = d
			continue
		}
		total = f(total, d)
	}
	return total.InexactFloat64()
}

// Add 两数相加
func (r *customerFunc) Add(args ...any) (any, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Add(d2)
	}, args...), nil
}

// Sub 两数相减
func (r *customerFunc) Sub(args ...any) (any, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Sub(d2)
	}, args...), nil
}

// Mul 两数相乘
func (r *customerFunc) Mul(args ...any) (any, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Mul(d2)
	}, args...), nil
}

// Div 两数相除
func (r *customerFunc) Div(args ...any) (any, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Div(d2)
	}, args...), nil
}

// Has 数组是否包含某元素
func (r *customerFunc) Has(args ...any) (any, error) {
	if len(args) != 2 {
		if len(args) == 1 {
			return false, nil
		}
		if len(args) > 2 {
			//这是一个bug，会将数组变成动态参数
			arg1 := args[0 : len(args)-1]
			return r.Has(arg1, args[len(args)-1])
		}
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	listInterface := args[0]
	item, _ := conv.Convert[string](args[1])
	listType := reflect.TypeOf(listInterface)
	listValue := reflect.ValueOf(listInterface)
	if listType.Kind() == reflect.Slice {
		for i := 0; i < listValue.Len(); i++ {
			if conv.String(listValue.Index(i).Interface()) == item {
				return true, nil
			}
		}
	} else if listType.Kind() == reflect.String {
		//这种字符串的格式：`["a", "b"]`
		list := make([]any, 0)
		_ = conv.Unmarshal(listInterface, &list)
		for _, v := range list {
			if conv.String(v) == item {
				return true, nil
			}
		}
	}
	return false, nil
}

// In 是否存在某数组中
func (r *customerFunc) In(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	return r.Has(args[1], args[0])
}

// Between 是否在某个范围内 Between(num, "[3,6]")
func (r *customerFunc) Between(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	rangeTemplate := conv.String(args[1])
	rangeTemplate = strings.Join(strings.Fields(rangeTemplate), "")

	rangeData, err := parseRange(rangeTemplate)
	if err != nil {
		return false, err
	}
	if rangeData == nil {
		return false, nil
	}
	num, err := conv.Convert[float64](args[0])
	if err != nil {
		return false, err
	}
	return rangeData.match(num), nil
}

// Is 是否是某一个类型
func (r *customerFunc) Is(args ...any) (any, error) {
	if len(args) <= 1 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	typeName := conv.String(args[0])
	typeName = strings.ToLower(typeName)
	if typeName == "nil" {
		return cond.IsNil(args[1]), nil
	}
	if typeName == "zero" {
		return cond.IsZero(args[1]), nil
	}
	if typeName == "number" {
		return cond.IsNumeric(args[1]), nil
	}
	if typeName == "time" {
		return cond.IsTime(conv.String(args[1])), nil
	}
	return false, fmt.Errorf("不支持的格式：%s", typeName)
}
func (r *customerFunc) As(args ...any) (any, error) {
	if len(args) <= 1 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	typeName := conv.String(args[0])
	typeName = strings.ToLower(typeName)
	if typeName == "nil" {
		return nil, nil
	}
	if typeName == "string" {
		return conv.String(args[1]), nil
	}
	if typeName == "int" || typeName == "int64" {
		if intTemp, err := conv.Convert[int64](args[1]); err == nil {
			return intTemp, nil
		}
		return 0, fmt.Errorf("参数不是int64类型：%v", args[1])
	}
	if typeName == "bool" {
		if boolTemp, err1 := conv.Convert[bool](args[1]); err1 == nil {
			return boolTemp, nil
		}
		return false, fmt.Errorf("参数不是bool类型：%v", args[1])
	}
	if typeName == "time" {
		if timeTemp, err1 := conv.Convert[time.Time](args[1]); err1 == nil {
			return timeTemp, nil
		}
		return time.Time{}, fmt.Errorf("参数不是time类型：%v", args[1])
	}
	if typeName == "float" || typeName == "float64" {
		if timeTemp, err1 := conv.Convert[float64](args[1]); err1 == nil {
			return timeTemp, nil
		}
		return time.Time{}, fmt.Errorf("参数不是float类型：%v", args[1])
	}
	return false, fmt.Errorf("不支持的格式：%s", typeName)
}
func (r *customerFunc) Replace(args ...any) (any, error) {
	if len(args) <= 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	oldStr := conv.String(args[1])
	newStr := conv.String(args[2])
	num := -1
	if len(args) >= 4 {
		if numTemp, err1 := conv.Convert[int](args[3]); err1 == nil {
			num = numTemp
		}
	}
	return strings.Replace(conv.String(args[0]), oldStr, newStr, num), nil
}
func (r *customerFunc) Split(args ...any) (any, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	splitArr := make([]string, 0)
	lo.ForEach(args[1:], func(item any, _ int) {
		splitArr = append(splitArr, conv.String(item))
	})
	return utils.Split(conv.String(args[0]), splitArr), nil
}
func (r *customerFunc) Contains(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	return strings.Contains(conv.String(args[0]), conv.String(args[1])), nil
}

// Array 将动态参数组成一个数组，供 Array(1,2,3) 使用
func (r *customerFunc) Array(args ...any) (any, error) {
	return args, nil
}

// Find 查找数组中第一个满足条件的元素并返回它，找不到返回 nil。
// 两种用法：
//  1. 相等匹配（原逻辑）：Find(Array(1,2,3), 2)
//  2. 表达式匹配：第二个参数传表达式字符串，对每个元素逐个求值，返回 true 即为命中。
//     表达式里用变量 item 指代“当前正在遍历的元素”，返回值不是布尔时会转成布尔判断。
//     元素是 map/结构体时，它的字段也会作为变量直接暴露出来，因此下面三种写法等价：
//     Find(items, 'age > 18')、Find(items, 'item.age > 18')、Find(items, '[item.age] > 18')
//     注意 1：本仓库的 govaluate 不支持点号取属性，item.age 会在编译前自动转成 [item.age]。
//     注意 2：外层已经把第二个参数包起来了，表达式里再出现引号必须转义，
//     例如 Find(items, "item.name == \'jack\'")。
//
// 兜底规则：第二个参数不是字符串、或这个字符串没法当成表达式（无法编译、没有引用任何
// 当前元素可用的变量）时，退化成原来的相等比较，保证 Find(Array('a','b'), 'a') 这类
// 老写法不受影响。
func (r *customerFunc) Find(args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("参数数量不对：%v", args)
	}

	// govaluate 的 separatorStage 在左值本身就是 []any 时，会把它摊平进函数的变长参数：
	// Find([]any{e1,e2}, expr) 实际收到的是 [e1, e2, expr]（数组被拆成了逐个参数）。
	// 这里做对应的还原：最后一个参数是表达式，前面的全是数组元素。
	// 若第一个参数本身就是数组，说明没有发生摊平（例如传的是 []string 这种带类型的切片）。
	var list []any
	var exprArg any
	if oneList := anySlice(args[0]); oneList != nil { //没有摊平
		list, exprArg = oneList, args[1]
	} else { //发生了摊平
		list, exprArg = args[:len(args)-1], args[len(args)-1]
	}

	//第二个参数是字符串时，才尝试把它当成“对当前元素求值”的表达式
	var itemCheck *itemExprChecker
	if exprStr, ok := exprArg.(string); ok {
		itemCheck = r.newItemExprChecker(exprStr)
	}

	for _, one := range list {
		if itemCheck != nil {
			ok, matched := itemCheck.match(itemParams(one), "Find")
			if matched { //表达式确实生效了，以表达式的结果为准
				if ok {
					return one, nil
				}
				continue
			}
			//表达式没生效（引用的变量对不上），走下面的相等比较
		}
		if reflect.DeepEqual(one, exprArg) {
			return one, nil
		}
	}
	//一个都没命中，且表达式执行有报错时，把错误抛出来，方便排查表达式写错的情况
	if itemCheck != nil && itemCheck.err != nil {
		return nil, itemCheck.err
	}
	return nil, nil
}
func (r *customerFunc) Filter(args ...any) (any, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("参数数量不对：%v", args)
	}

	var list []any
	var exprArg any
	if oneList := anySlice(args[0]); oneList != nil { //没有摊平
		list, exprArg = oneList, args[1]
	} else { //发生了摊平
		list, exprArg = args[:len(args)-1], args[len(args)-1]
	}

	//第二个参数是字符串时，才尝试把它当成“对当前元素求值”的表达式
	var itemCheck *itemExprChecker
	if exprStr, ok := exprArg.(string); ok {
		itemCheck = r.newItemExprChecker(exprStr)
	}
	retList := make([]any, 0)
	for _, one := range list {
		if itemCheck != nil {
			ok, matched := itemCheck.match(itemParams(one), "Filter")
			if matched { //表达式确实生效了，以表达式的结果为准
				if ok {
					retList = append(retList, one)
				}
				continue
			}
			//表达式没生效（引用的变量对不上），走下面的相等比较
		}
		if reflect.DeepEqual(one, exprArg) {
			retList = append(retList, one)
			continue
		}
	}
	return retList, nil
}

// Map 把数组映射成新数组，功能类似 lo.Map，输入多长输出就多长。
// 两种用法：
//  1. 表达式：第二个参数是表达式，对每个元素逐个求值，结果组成新数组。
//     例：Map(Array(1,2,3), 'item + 1')、Map(users, 'item.age * 2')
//  2. 取字段：第二个参数是字段名，取每个元素该字段的值组成新数组。
//     例：Map(users, 'name')、Map(users, 'addr.city')（支持点号取嵌套字段）
//
// 表达式里用变量 item 指代“当前正在遍历的元素”，元素的字段也会作为变量直接暴露，
// 所以 Map(users, 'item.name')、Map(users, '[item.name]')、Map(users, 'name') 三种写法等价。
//
// 兜底规则：第二个参数不是字符串时报错；字符串无法当成表达式时按字段名处理；
// 字段名在元素里不存在时，该位置填 nil（数组长度保持不变）。
func (r *customerFunc) Map(args ...any) (any, error) {
	//第一个参数是空数组时，govaluate 摊平后只会剩下表达式一个参数，映射结果也是空数组
	if len(args) == 1 {
		if _, ok := args[0].(string); ok {
			return []any{}, nil
		}
	}
	if len(args) < 2 {
		return nil, fmt.Errorf("参数数量不对：%v", args)
	}

	var list []any
	var exprArg any
	if oneList := anySlice(args[0]); oneList != nil { //没有摊平
		list, exprArg = oneList, args[1]
	} else { //发生了摊平
		list, exprArg = args[:len(args)-1], args[len(args)-1]
	}

	exprStr, ok := exprArg.(string)
	if !ok {
		return nil, fmt.Errorf("第二个参数必须是表达式或字段名：%v", exprArg)
	}

	//第二个参数是字符串时，先尝试把它当成“对当前元素求值”的表达式
	var itemCheck *itemExprChecker
	if exprStr != "" {
		itemCheck = r.newItemExprChecker(exprStr)
	}

	retList := make([]any, 0, len(list))
	for _, one := range list {
		params := itemParams(one)
		if itemCheck != nil {
			retVal, matched := itemCheck.eval(params, "Map")
			if matched { //表达式确实生效了，用它的求值结果
				retList = append(retList, retVal)
				continue
			}
		}
		//表达式没生效（或压根不是表达式），当成字段名取值，取不到填 nil
		retList = append(retList, params[exprStr])
	}
	return retList, nil
}
func (r *customerFunc) Len(args ...any) (any, error) {
	if len(args) == 0 {
		return float64(0), nil
	}
	var list []any
	if oneList := anySlice(args[0]); oneList != nil {
		list = oneList
	} else {
		list = args
	}
	return float64(len(list)), nil
}

// itemVarName 表达式里指代“当前元素”的变量名
const itemVarName = "item"

// itemExprChecker 针对单个元素求值的表达式校验器
type itemExprChecker struct {
	expression *govaluate.EvaluableExpression
	vars       []string //表达式里引用到的变量名
	err        error    //上一次求值的错误
}

// buildItemExpression 编译子表达式：优先直接编译；
// 本仓库的 govaluate 不支持 item.age 这种点号访问，编译失败时再转成 [item.age] 重试一次
func (r *customerFunc) buildItemExpression(exprStr string) (*govaluate.EvaluableExpression, error) {
	if r.engLogic != nil {
		expression, err := r.engLogic.getExpressionByRuleString(exprStr)
		if err == nil && expression != nil {
			return expression, nil
		}
	} else {
		expression, err := govaluate.NewEvaluableExpression(exprStr)
		if err == nil && expression != nil {
			return expression, nil
		}
	}
	//点号访问不被支持，转成 [item.xxx] 的写法再试一次
	if escaped := escapeItemAccessor(exprStr); escaped != exprStr {
		if r.engLogic != nil {
			return r.engLogic.getExpressionByRuleString(escaped)
		}
		return govaluate.NewEvaluableExpression(escaped)
	}
	return nil, fmt.Errorf("表达式格式错误：%s", exprStr)
}

// newItemExprChecker 将 exprStr 编译成对元素求值的表达式，无法当成表达式时返回 nil
func (r *customerFunc) newItemExprChecker(exprStr string) *itemExprChecker {
	if exprStr == "" {
		return nil
	}
	expression, err := r.buildItemExpression(exprStr)
	if err != nil || expression == nil {
		return nil //不是合法的表达式，交给原来的相等比较处理
	}

	varList := make([]string, 0)
	for _, token := range expression.Tokens() {
		if token.Kind != govaluate.VARIABLE {
			continue
		}
		varName, ok := token.Value.(string)
		if !ok || varName == "" {
			continue
		}
		//形如 item.age、item.a.b 的变量，取根变量名即可
		if pos := strings.IndexAny(varName, ".[ "); pos >= 0 {
			varName = varName[:pos]
		}
		if varName != "" {
			varList = utils.AppendUniq(varList, varName)
		}
	}
	if len(varList) == 0 {
		return nil //没有引用任何变量，说明不是针对元素的表达式
	}
	return &itemExprChecker{expression: expression, vars: varList}
}

// match 返回 (是否符合条件, 表达式是否生效)，供 Find / Filter 使用。
// funcName 只用于拼错误信息，方便区分是哪个函数调用失败了。
func (i *itemExprChecker) match(params map[string]any, funcName string) (bool, bool) {
	retVal, matched := i.eval(params, funcName)
	if !matched {
		return false, false
	}
	if boolVal, ok := retVal.(bool); ok {
		return boolVal, true
	}
	if boolVal, err1 := conv.Convert[bool](retVal); err1 == nil {
		return boolVal, true
	}
	return false, true
}

// eval 返回 (表达式的原始求值结果, 表达式是否生效)，供 Map 使用。
// funcName 只用于拼错误信息，方便区分是哪个函数调用失败了。
func (i *itemExprChecker) eval(params map[string]any, funcName string) (any, bool) {
	//表达式引用的变量在当前元素里一个都不存在时，认为该表达式不适用
	applicable := false
	for _, varName := range i.vars {
		if _, ok := params[varName]; ok {
			applicable = true
			break
		}
	}
	if !applicable {
		return nil, false
	}

	retVal, err := i.expression.Evaluate(params)
	if err != nil {
		i.err = fmt.Errorf("%s表达式执行失败：%w", funcName, err)
		return nil, true
	}
	return retVal, true
}

// itemParams 构造元素求值时的变量表：
// 元素本身绑定到 item 变量；元素是 map/结构体时，它的字段也一并作为变量暴露出来
func itemParams(item any) map[string]any {
	params := make(map[string]any, 8)
	if itemMap, ok := item.(map[string]any); ok {
		fillItemParams(params, itemMap, "")
	} else if item != nil {
		val := reflect.ValueOf(item)
		if val.Kind() == reflect.Struct {
			var itemMap map[string]any
			if err := conv.Unmarshal(item, &itemMap); err == nil {
				fillItemParams(params, itemMap, "")
			}
		}
	}
	params[itemVarName] = item //item 变量名优先级最高，避免元素里刚好有个字段叫 item
	return params
}

// fillItemParams 把元素的字段铺到变量表里：
// 一级字段既能直接用（age），也能按 [item.age] 的方式用；嵌套字段按 a.b / item.a.b 铺开
func fillItemParams(params map[string]any, itemMap map[string]any, prefix string) {
	for key, val := range itemMap {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		if _, ok := params[fullKey]; !ok {
			params[fullKey] = val
		}
		params[itemVarName+"."+fullKey] = val
		if subMap, ok := val.(map[string]any); ok {
			fillItemParams(params, subMap, fullKey)
		}
	}
}

// escapeItemAccessor 把 item.a.b 这种点号访问，转成本引擎支持的 [item.a.b] 写法。
// 只处理引号外面的内容，避免把字符串字面量里的点也改掉。
func escapeItemAccessor(exprStr string) string {
	var buf strings.Builder
	inQuote := byte(0)
	for idx := 0; idx < len(exprStr); idx++ {
		char := exprStr[idx]
		if inQuote != 0 {
			buf.WriteByte(char)
			if char == '\\' && idx+1 < len(exprStr) { //转义字符原样保留
				idx++
				buf.WriteByte(exprStr[idx])
				continue
			}
			if char == inQuote {
				inQuote = 0
			}
			continue
		}
		if char == '\'' || char == '"' {
			inQuote = char
			buf.WriteByte(char)
			continue
		}
		//找到一个独立的 item 变量，且后面跟着点号
		if strings.HasPrefix(exprStr[idx:], itemVarName+".") &&
			(idx == 0 || !isVarChar(exprStr[idx-1])) {
			end := idx + len(itemVarName) + 1
			for end < len(exprStr) && isVarCharOrDot(exprStr[end]) {
				end++
			}
			chain := strings.TrimRight(exprStr[idx+len(itemVarName)+1:end], ".")
			if chain != "" {
				buf.WriteString("[" + itemVarName + "." + chain + "]")
				idx = end - 1
				continue
			}
		}
		buf.WriteByte(char)
	}
	return buf.String()
}

func isVarChar(char byte) bool {
	return char == '_' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
}

func isVarCharOrDot(char byte) bool {
	return char == '.' || isVarChar(char)
}

// anySlice 把任意类型的切片/数组统一转成 []any，非切片或 nil 返回 nil
func anySlice(src any) []any {
	if list, ok := src.([]any); ok {
		return list
	}
	if src == nil {
		return nil
	}
	val := reflect.ValueOf(src)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil
	}
	if val.Kind() == reflect.Slice && val.IsNil() {
		return nil
	}
	if val.Type().Elem().Kind() == reflect.Uint8 { //[]byte 之类不当成数组处理
		return nil
	}
	list := make([]any, 0, val.Len())
	for idx := 0; idx < val.Len(); idx++ {
		list = append(list, val.Index(idx).Interface())
	}
	return list
}

func (r *customerFunc) JsonGet(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	jsonStr := conv.String(args[0])
	pathStr := conv.String(args[1])
	gResult := gjson.Get(jsonStr, pathStr)
	if !gResult.Exists() {
		return "", nil
	}
	return gResult.Value(), nil
}

// Join 连接字符串，第一个字符为连接符
// Join("/", "a", "b", "c")
func (r *customerFunc) Join(args ...any) (any, error) {
	sep := ""
	var dataList []any
	if len(args) == 0 {
		return "", nil
	} else if len(args) == 1 {
		if list, ok := args[0].([]any); ok {
			dataList = list
		} else {
			return conv.String(args[0]), nil
		}
	} else if len(args) == 2 {
		sep = conv.String(args[0])
		if list, ok := args[1].([]string); ok {
			dataList = lo.Map(list, func(item string, index int) any {
				return any(item)
			})
		} else {
			dataList = args[1:]
		}
	} else {
		sep = conv.String(args[0])
		dataList = args[1:]
	}
	retStr := make([]string, 0)
	lo.ForEach(dataList, func(item any, _ int) {
		retStr = append(retStr, conv.String(item))
	})
	return strings.Join(retStr, sep), nil
}

// MaskMatch 含有掩码的字符串比较
// MaskMatch("a", "b", "*")
func (r *customerFunc) MaskMatch(args ...any) (any, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	args0 := conv.String(args[0])
	args1 := conv.String(args[1])

	maskStr := ""
	if len(args) >= 3 {
		maskStr = conv.String(args[2])
	}
	return mask.IsMatch(args0, args1, maskStr), nil
}

// If 三元运算符
func (r *customerFunc) If(args ...any) (any, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("ternary function requires exactly 3 arguments: condition, trueValue, falseValue")
	}
	// 第一个参数必须是布尔类型
	condition, ok := args[0].(bool)
	if !ok {
		return nil, fmt.Errorf("first argument to ternary must be a boolean")
	}
	if condition {
		return args[1], nil
	}
	return args[2], nil
}

// Switch 多分支选择，避免多重嵌套 If。
// 用法: Switch(value, case1, result1, case2, result2, ..., defaultValue)
//   - args[0]         : 待匹配的值
//   - 后续成对        : (候选值, 命中返回值)
//   - 最后一个参数     : 都不命中时的默认值
//
// 例: Switch(status, "A", "苹果", "B", "香蕉", "未知")
func (r *customerFunc) Switch(args ...any) (any, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("switch requires at least 3 arguments: value, [candidate, result]..., defaultValue")
	}
	if (len(args)-1)%2 == 0 {
		return nil, fmt.Errorf("switch arguments must be value, then pairs of (candidate, result), then a defaultValue")
	}
	value := args[0]
	// 逐对比较（用 conv.Equal 做跨类型宽松相等，如 "1" 与 1）
	for i := 1; i < len(args)-1; i += 2 {
		if cond.IsEqual(value, args[i]) {
			return args[i+1], nil
		}
	}
	// 全部未命中，返回默认值
	return args[len(args)-1], nil
}

// SwitchExpr 按布尔条件分支：SwitchExpr(cond1, res1, cond2, res2, ..., defaultValue)
// 例: SwitchExpr(score >= 90, "优", score >= 80, "良", "及格")
func (r *customerFunc) SwitchExpr(args ...any) (any, error) {
	if len(args) < 3 || (len(args)-1)%2 != 0 {
		return nil, fmt.Errorf("switchExpr requires value-less pairs: (condition, result)..., defaultValue")
	}
	for i := 0; i < len(args)-1; i += 2 {
		condBool, ok := args[i].(bool)
		if !ok {
			return nil, fmt.Errorf("switchExpr condition at position %d must be bool", i)
		}
		if condBool {
			return args[i+1], nil
		}
	}
	return args[len(args)-1], nil
}

func (r *customerFunc) Max(args ...any) (any, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("参数为空")
	}
	var currentNum float64
	var found bool
	var notFirst bool
	lo.ForEach(args, func(item any, _ int) {
		one, err := conv.Convert[float64](item)
		if err != nil {
			return
		}
		found = true
		if !notFirst {
			currentNum = one
			notFirst = true
			return
		}
		if one > currentNum {
			currentNum = one
		}
	})
	if !found {
		return 0, fmt.Errorf("没有找到数字")
	}

	return currentNum, nil
}

func (r *customerFunc) Min(args ...any) (any, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("参数为空")
	}
	var currentNum float64
	var found bool
	var notFirst bool
	lo.ForEach(args, func(item any, _ int) {
		one, err := conv.Convert[float64](item)
		if err != nil {
			return
		}
		found = true
		if !notFirst {
			currentNum = one
			notFirst = true
			return
		}
		if one < currentNum {
			currentNum = one
		}
	})
	if !found {
		return 0, fmt.Errorf("没有找到数字")
	}
	return currentNum, nil
}

func (r *customerFunc) getDecimalBaseAndNum(args ...any) (int64, decimal.Decimal, error) {
	var initNum float64 = 0
	var initBase int64 = 0
	initDecimalNum := decimal.NewFromFloat(initNum)

	var baseNum any = 10
	var numDecimal any = 0

	if len(args) == 0 {
		return initBase, initDecimalNum, fmt.Errorf("参数数量不对，需要2个参数：位数和数字")
	} else if len(args) == 1 {
		numDecimal = args[0]
	} else if len(args) == 2 {
		baseNum = args[0]
		numDecimal = args[1]
	}

	// 获取基数参数（10, 100, 1000等）
	base, err := conv.Convert[int64](baseNum)
	if err != nil {
		return initBase, initDecimalNum, fmt.Errorf("基数参数转换失败：%v", baseNum)
	}

	// 验证基数是否为10的幂次方
	if !isValidBase(base) {
		return initBase, initDecimalNum, fmt.Errorf("基数必须是10的幂次方（1, 10, 100, 1000...），当前值：%v", base)
	}

	// 获取数字参数
	num, err := conv.Convert[float64](numDecimal)
	if err != nil {
		return initBase, initDecimalNum, fmt.Errorf("数字参数转换失败：%v", numDecimal)
	}

	return base, decimal.NewFromFloat(num), nil
}

// CeilToDigit 指定位数向上取整，默认是10进位
func (r *customerFunc) CeilToDigit(args ...any) (any, error) {
	intBase, decimalNum, err := r.getDecimalBaseAndNum(args...)
	if err != nil {
		return 0, err
	}

	decimalBase := decimal.NewFromInt(intBase)
	result := decimalNum.Div(decimalBase).Ceil().Mul(decimalBase)

	return result.InexactFloat64(), nil
}

// FloorToDigit 指定位数向下取整，默认是10进位
func (r *customerFunc) FloorToDigit(args ...any) (any, error) {
	intBase, decimalNum, err := r.getDecimalBaseAndNum(args...)
	if err != nil {
		return 0, err
	}

	decimalBase := decimal.NewFromInt(intBase)

	// 除以基数，向上取整，再乘以基数
	result := decimalNum.Div(decimalBase).Floor().Mul(decimalBase)

	return result.InexactFloat64(), nil
}

// isValidBase 验证基数是否为10的幂次方（1, 10, 100, 1000...）
func isValidBase(base int64) bool {
	if base <= 0 {
		return false
	}
	if base == 1 {
		return true
	}
	for base >= 10 {
		base = base / 10
	}
	return base == 1
}
