// Package param 获取参数
package param

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/magic-lib/go-plat-utils/conv"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const (
	LocationHeader Location = "header"
	LocationCookie Location = "cookie"
	LocationQuery  Location = "query"
	LocationBody   Location = "body"
)

// Validator represents a validator.
type Validator interface {
	// Validate validates the value.
	Validate() error
}

type Location string

// ParsePathFunc 解析地址中 /name/:name 等特殊变量
type ParsePathFunc func(r *http.Request) map[string]string

// Param 可以全局定义获取参数的方法
type Param struct {
	defaultBodyKeyName      string
	querySplit              string
	locationOrder           []Location
	parsePathFunc           ParsePathFunc
	customErrorMessages     map[string]string
	errorMessagesTranslator func(r context.Context, errorKey, errorMsg string) error
}

// NewParam 新建
func NewParam() *Param {
	return &Param{
		querySplit:         ",",
		locationOrder:      []Location{LocationBody, LocationQuery},
		defaultBodyKeyName: "request___body___key_name",
	}
}

// SetQuerySplit 设置query传数组以后需要连接起来的字符串分隔符，因为一般不会有多个
func (p *Param) SetQuerySplit(splitStr string) *Param {
	p.querySplit = splitStr
	return p
}

// SetLocationOrder 设置获取参数的位置
func (p *Param) SetLocationOrder(list []Location) *Param {
	p.locationOrder = list
	return p
}

// SetParsePathFunc 设置获取参数的方法
func (p *Param) SetParsePathFunc(pathFunc ParsePathFunc) *Param {
	p.parsePathFunc = pathFunc
	return p
}

// SetValidatorCustomErrorMessages 设置获取参数的方法
func (p *Param) SetValidatorCustomErrorMessages(customErrorMessages map[string]string, errorMessagesTranslator func(ctx context.Context, errorKey, errorMsg string) error) *Param {
	if customErrorMessages != nil && len(customErrorMessages) > 0 {
		p.customErrorMessages = customErrorMessages
	}
	if errorMessagesTranslator != nil {
		p.errorMessagesTranslator = errorMessagesTranslator
	}
	return p
}

func getMapFromHeaderCookieQuery(l Location, r *http.Request, querySplit string, pathFunc ParsePathFunc) (map[string]any, any, bool) {
	allRetParamMap := make(map[string]any)
	if l == LocationHeader {
		headerMap := getParamFromHeader(r)
		for k, v := range headerMap {
			allRetParamMap[k] = v
		}
		return allRetParamMap, headerMap, true
	}
	if l == LocationCookie {
		headerMap := getParamFromCookie(r)
		for k, v := range headerMap {
			allRetParamMap[k] = v
		}
		return allRetParamMap, headerMap, true
	}
	if l == LocationQuery {
		//如果有多个，则为数组，否则为字符串
		headerMap, _ := getParamFromQuery(r, querySplit, pathFunc)
		for k, v := range headerMap {
			allRetParamMap[k] = v
		}
		return allRetParamMap, headerMap, true
	}

	return nil, nil, false
}
func getMapFromBody(r *http.Request, defaultBodyKeyName string, valueToString bool, querySplit string) (map[string]any, any, error) {
	allRetParamMap, bodyDataStr, err := getMapFromBodyForm(r, defaultBodyKeyName)

	if json.Valid([]byte(bodyDataStr)) {
		allRetParamMap1, paasParamMap, err1 := getMapFromBodyJsonString(bodyDataStr, valueToString)
		for key, one := range allRetParamMap1 {
			allRetParamMap[key] = one
		}
		if err1 == nil {
			return allRetParamMap, paasParamMap, nil
		} else {
			err = err1
		}
	} else {
		allRetParamMap1, err1 := getMapFromBodyQueryString(bodyDataStr, valueToString, querySplit)
		for key, one := range allRetParamMap1 {
			allRetParamMap[key] = one
		}
		if err1 == nil {
			return allRetParamMap, bodyDataStr, nil
		} else {
			err = err1
		}
	}
	return allRetParamMap, bodyDataStr, err
}

func getMapFromBodyJsonString(bodyDataStr string, valueToString bool) (map[string]any, any, error) {
	allRetParamMap := make(map[string]any)
	paasParamMap := make(map[string]any)
	err := conv.Unmarshal(bodyDataStr, &paasParamMap)
	if err == nil && len(paasParamMap) > 0 {
		//表示是map格式
		for key, one := range paasParamMap {
			allRetParamMap[key] = one
		}
		if valueToString {
			for key, one := range allRetParamMap {
				allRetParamMap[key] = conv.String(one)
			}
		}
		return allRetParamMap, paasParamMap, nil
	}
	//array
	paasParamArray := make([]any, 0)
	err = conv.Unmarshal(bodyDataStr, &paasParamArray)
	if err == nil && len(paasParamArray) > 0 {
		return allRetParamMap, paasParamArray, nil
	}

	if err == nil {
		err = fmt.Errorf("bodyDataStr is not a valid json")
	}

	return allRetParamMap, bodyDataStr, err
}
func getMapFromBodyQueryString(bodyDataStr string, valueToString bool, querySplit string) (map[string]any, error) {
	allRetParamMap := make(map[string]any)

	// 如果body中有aaa=bbb&ccc=ddd的格式的话，则直接转换过来
	paramMap, err := getMapByQueryString(bodyDataStr)

	for key, one := range paramMap {
		if valueToString {
			tempVal := ""
			if len(one) == 1 {
				tempVal = one[0]
			} else if len(one) > 1 {
				if querySplit == "" {
					tempVal = conv.String(one)
				} else {
					tempVal = strings.Join(one, querySplit)
				}
			}
			allRetParamMap[key] = tempVal
		} else {
			allRetParamMap[key] = one
		}
	}

	return allRetParamMap, err
}

func getMapFromBodyForm(r *http.Request, defaultBodyKeyName string) (map[string]any, string, error) {
	bodyForm := make(map[string]any)

	//先取form的值
	forms := getParamFromForm(r)
	for key, _ := range forms {
		bodyForm[key] = forms.Get(key)
	}

	var bodyDataStr string
	var err error
	bodyDataStr, err = getParamFromBody(r)
	if bodyDataStr == "" {
		//如果返回为空，则可能是因为header中没有添加 Content-Type: application/json，造成获取不到的情况
		if len(forms) == 1 {
			bodyContent := ""
			for key := range bodyForm {
				if key != "" {
					bodyContent = key
				}
				break
			}
			if bodyContent != "" {
				//为json串才记录，否则与后面的bodyForm重复了
				if json.Valid([]byte(bodyContent)) {
					bodyDataStr = bodyContent
				}
			}
		}
	}

	allRetParamMap := make(map[string]any)
	for key, val := range bodyForm {
		if key != "" {
			allRetParamMap[key] = val
		}
	}

	if defaultBodyKeyName != "" {
		allRetParamMap[defaultBodyKeyName] = bodyDataStr
	}

	return allRetParamMap, bodyDataStr, err
}

func getAllByLocation(r *http.Request, l Location, defaultBodyKeyName string, valueToString bool, querySplit string, pathFunc ParsePathFunc) (map[string]any, any, error) {
	if l == "" {
		return nil, nil, fmt.Errorf("location is null")
	}

	if l == LocationHeader || l == LocationCookie || l == LocationQuery {
		allRetParam, headerMap, found := getMapFromHeaderCookieQuery(l, r, querySplit, pathFunc)
		if found {
			return allRetParam, headerMap, nil
		}
	}

	if l == LocationBody {
		return getMapFromBody(r, defaultBodyKeyName, valueToString, querySplit)
	}

	allRetParamMap := make(map[string]any)
	return allRetParamMap, nil, fmt.Errorf("location error: %s", l)
}

// GetAll 转成any
func (p *Param) GetAll(r *http.Request) any {
	allParamMap := make(map[string]any)
	for _, one := range p.locationOrder {
		oneParamMap, oneParamRet, err := getAllByLocation(r, one, p.defaultBodyKeyName, false, p.querySplit, p.parsePathFunc)
		for key, val := range oneParamMap {
			allParamMap[key] = val
		}
		if err == nil && oneParamRet != nil {
			if oneParamMapTemp, ok := oneParamRet.(map[string]any); ok {
				for key, val := range oneParamMapTemp {
					allParamMap[key] = val
				}
			} else if oneParamList, ok := oneParamRet.([]any); ok {
				return oneParamList
			}
		}
	}

	//为空，则不添加key
	if bodyData, ok := allParamMap[p.defaultBodyKeyName]; ok {
		if conv.String(bodyData) == "" || conv.String(bodyData) == "null" {
			delete(allParamMap, p.defaultBodyKeyName)
		}
	}

	return allParamMap
}

// GetAllMap 转成map hasAll 包含所有的变量，false则去掉默认的data返回值，简化内容
func (p *Param) GetAllMap(r *http.Request, hasAll bool) map[string]any {
	paramMap := p.getAllBasic(r, false)
	if hasAll {
		return paramMap
	}
	//重新复制，不破坏原来的值
	newParamMap := make(map[string]any)
	err := conv.Unmarshal(paramMap, &newParamMap)
	if err != nil {
		return paramMap
	}
	if bodyAll, ok := newParamMap[p.defaultBodyKeyName]; ok {
		keyAll := conv.String(bodyAll)
		if valAll, ok := newParamMap[keyAll]; ok {
			if conv.String(valAll) == "[\"\"]" {
				delete(newParamMap, keyAll)
			}
		}
		delete(newParamMap, p.defaultBodyKeyName)
	}
	return newParamMap
}

// Parse 简单赋值
// openValidate 打开检查，默认是true; dst 传指针
func (p *Param) Parse(r *http.Request, dst any, openValidate ...bool) error {
	paramMap := p.GetAllMap(r, true)
	err := conv.Unmarshal(paramMap, dst)
	if err != nil {
		//用post里的数据整体进行，如果传的是数组的话
		var listErr error
		dstValue := reflect.Indirect(reflect.ValueOf(dst))
		if dstValue.Kind() == reflect.Slice {
			if listStr, ok := paramMap[p.defaultBodyKeyName]; ok {
				listErr = conv.Unmarshal(listStr, dst)
			}
		}
		if listErr != nil {
			return err
		}
	}
	isCheck := true

	if len(openValidate) == 1 {
		isCheck = openValidate[0]
	}

	if !isCheck {
		return nil
	}

	//使用对象自身的验证
	if valid, ok := dst.(Validator); ok {
		return valid.Validate()
	}
	//使用默认的validator标签验证
	//如果不是struct类型，则直接跳过默认验证
	val := reflect.ValueOf(dst)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	timeType := reflect.TypeOf(time.Time{})
	if val.Kind() != reflect.Struct || val.Type().ConvertibleTo(timeType) {
		return nil
	}
	validate := validator.New()
	if err = validate.StructCtx(r.Context(), dst); err != nil {
		if p.customErrorMessages == nil || len(p.customErrorMessages) == 0 {
			return err
		}
		if errors.Is(err, new(validator.InvalidValidationError)) {
			return err
		}

		// 获取包路径,可以对struct名相同的进行区分开来，避免冲突
		typ := reflect.TypeOf(dst)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		pkgPath := typ.PkgPath()
		for _, errOne := range err.(validator.ValidationErrors) {
			errKey := fmt.Sprintf("%s.%s", errOne.StructNamespace(), errOne.Tag())
			errKey2 := ""
			if pkgPath != "" {
				errKey2 = fmt.Sprintf("%s/%s", pkgPath, errKey)
			}
			if customMsg, exists := p.customErrorMessages[errKey2]; exists {
				if p.errorMessagesTranslator != nil {
					return p.errorMessagesTranslator(r.Context(), errKey2, customMsg)
				}
				return fmt.Errorf(customMsg)
			}
			if customMsg, exists := p.customErrorMessages[errKey]; exists {
				if p.errorMessagesTranslator != nil {
					return p.errorMessagesTranslator(r.Context(), errKey, customMsg)
				}
				return fmt.Errorf(customMsg)
			}
			log.Default().Println("customErrorMessages map key has package: ", errKey2)
			log.Default().Println("customErrorMessages map key: ", errKey)
		}
		return err
	}
	return nil
}

// GetAllString 转成字符串，主要解决[]string问题
func (p *Param) GetAllString(r *http.Request) map[string]string {
	allParamMap := p.getAllBasic(r, true)
	allParamStr := make(map[string]string)
	for k, v := range allParamMap {
		allParamStr[k] = conv.String(v)
	}
	return allParamStr
}

func (p *Param) getAllBasic(r *http.Request, valueToString bool) map[string]any {
	allParamMap := make(map[string]any)
	for _, one := range p.locationOrder {
		oneParamMap, _, err := getAllByLocation(r, one, p.defaultBodyKeyName, valueToString, p.querySplit, p.parsePathFunc)
		if err == nil && oneParamMap != nil {
			for key, val := range oneParamMap {
				allParamMap[key] = val
			}
		}
	}
	return allParamMap
}

// GetAllHeaders 获取所有Headers
func (p *Param) GetAllHeaders(r *http.Request) http.Header {
	return getAllHeaders(r)
}

// GetAllCookies 获取所有Cookies
func (p *Param) GetAllCookies(r *http.Request) map[string]*http.Cookie {
	return getAllCookies(r)
}

// GetAllQuery 获取所有Query
func (p *Param) GetAllQuery(r *http.Request) map[string]string {
	_, allParamMap, _ := getMapFromHeaderCookieQuery(LocationQuery, r, p.querySplit, p.parsePathFunc)
	if allParamStringMap, ok := allParamMap.(map[string]string); ok {
		return allParamStringMap
	}
	return map[string]string{}
}

// GetAllBody 获取所有Body内容
func (p *Param) GetAllBody(r *http.Request) string {
	allParamBody, _ := getParamFromBody(r)
	if allParamBody != "" {
		return allParamBody
	}
	//form里的数据
	forms := getParamFromForm(r)
	if len(forms) == 0 {
		return ""
	}

	retMap := make(map[string]any)
	for key, one := range forms {
		retMap[key] = one
	}

	//如果是query的，则直接覆盖，避免有数组的格式，读取数据不正确
	rawQuery := r.URL.RawQuery
	queryMap, err := url.ParseQuery(rawQuery)
	if err == nil {
		for key, one := range queryMap {
			if old, ok := forms[key]; ok {
				if len(one) == 1 && len(old) == 1 && one[0] == old[0] {
					retMap[key] = one[0] //进行替换
				}
			}
		}
	}

	return conv.String(retMap)
}

//// GetGinParamFromUrl 取得地址栏的参数
//func (p *paramStruct) GetGinParamFromUrl(ginParamList gin.Params, paramName ...string) map[string]string {
//	retMap := make(map[string]string)
//	if len(paramName) == 0 {
//		for _, entry := range ginParamList {
//			retMap[entry.Key] = entry.Value
//		}
//		return retMap
//	}
//	for _, key := range paramName {
//		if key == "" {
//			continue
//		}
//		if va, ok := ginParamList.Get(key); ok {
//			retMap[key] = va
//		}
//	}
//	return retMap
//}

func getParamFromHeader(req *http.Request) map[string]string {
	allParamMap := make(map[string]string)
	headers := getAllHeaders(req)
	for i, _ := range headers {
		allParamMap[i] = headers.Get(i)
	}
	return allParamMap
}

func getParamFromCookie(req *http.Request) map[string]string {
	allParamMap := make(map[string]string)
	cookies := getAllCookies(req)
	for i, one := range cookies {
		allParamMap[i] = one.Value
	}
	return allParamMap
}

func getParamFromQuery(req *http.Request, splitString string, pathFunc ParsePathFunc) (map[string]string, error) {
	allParamMap := make(map[string]string)
	if req == nil {
		return allParamMap, nil
	}

	query, err := getMapByQueryString(req.URL.RawQuery)
	if len(query) > 0 {
		for i, one := range query {
			tempVal := ""
			if len(one) == 0 {
				tempVal = ""
			} else if len(one) == 1 {
				tempVal = one[0]
			} else if len(one) > 1 {
				if splitString == "" {
					tempVal = conv.String(one)
				} else {
					tempVal = strings.Join(one, splitString)
				}
			}
			allParamMap[i] = tempVal
		}
	}

	if pathFunc != nil {
		pathMap := pathFunc(req)
		if pathMap != nil && len(pathMap) > 0 {
			for i, v := range pathMap {
				allParamMap[i] = v
			}
		}

	}

	return allParamMap, err
}

func getParamFromBody(req *http.Request) (string, error) {
	if req == nil || req.Body == nil {
		return "", nil
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(b))

	return string(b), nil
}
func getParamFromForm(req *http.Request) url.Values {
	queryMap := make(url.Values)
	if req == nil {
		return queryMap
	}

	var formValue url.Values

	err := req.ParseForm()
	if err == nil {
		formValue = req.Form
	}
	postValue := req.PostForm
	if formValue == nil {
		formValue = postValue
	} else {
		if postValue != nil {
			for keyName, one := range postValue {
				for _, oneVal := range one {
					oneVal = strings.TrimSpace(oneVal)
					if oneVal == "" {
						continue
					}
					if !formValue.Has(keyName) {
						formValue.Add(keyName, oneVal)
						continue
					}
					isFind := false
					for _, oneVal1 := range formValue[keyName] {
						if oneVal1 == oneVal {
							isFind = true
							break
						}
					}
					//不相同才进行添加，避免相同的情况产生，重复了
					if !isFind {
						formValue.Add(keyName, oneVal)
					}
				}
			}
		}
	}

	if formValue != nil && len(formValue) > 0 {
		return formValue
	}

	return queryMap
}

func getMapByQueryString(query string) (url.Values, error) {
	queryMap := make(url.Values)

	q, err := url.ParseQuery(query)
	if err == nil {
		if q == nil {
			return queryMap, nil
		}

		for k, v := range q {
			newV := make([]string, 0)
			for _, one := range v {
				one = strings.TrimSpace(one)
				if one != "" {
					newV = append(newV, one)
				}
			}
			queryMap[k] = newV
		}
		return queryMap, nil
	}

	p := make(url.Values)
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		oldValue := value
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			value = oldValue
		}
		p[key] = append(p[key], value)
	}
	if p != nil && len(p) > 0 {
		for k, v := range p {
			newV := make([]string, 0)
			for _, one := range v {
				if one != "" {
					newV = append(newV, one)
				}
			}
			queryMap[k] = newV
		}
	}
	return queryMap, err
}
