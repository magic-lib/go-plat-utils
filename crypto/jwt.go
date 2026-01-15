package crypto

import (
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/gorm/utils"
	"log"
	"reflect"
	"strings"
	"time"
)

const (
	defaultDataKey = "data_define_name" //内置名，避免覆盖外面对象的名称
)

var (
	standardClaimsKeys []string
)

type JwtCfg struct {
	EncryptJsonKeyList []string //是否对数据key使用加密，避免外部查看
	StandardClaims     *jwt.StandardClaims
	SigningMethod      *jwt.SigningMethodHMAC
}

// JwtEncrypt RSA加密数据，secretKey必须是成对出现
func JwtEncrypt(secretKey string, data any, cfgList ...*JwtCfg) (string, error) {
	var cfg = &jwt.StandardClaims{}
	var signingMethod = jwt.SigningMethodHS256
	var useEncryptList []string
	if len(cfgList) > 0 && cfgList[0] != nil {
		oneCfg := cfgList[0]
		if oneCfg.StandardClaims != nil {
			cfg = oneCfg.StandardClaims
		}
		if oneCfg.SigningMethod != nil {
			signingMethod = oneCfg.SigningMethod
		}
		useEncryptList = oneCfg.EncryptJsonKeyList
	}

	if cfg.IssuedAt == 0 {
		cfg.IssuedAt = time.Now().Unix()
	}
	if cfg.ExpiresAt == 0 {
		cfg.ExpiresAt = time.Now().Add(time.Hour).Unix()
	}
	mapClaims := jwtEncryptDataToMap(secretKey, useEncryptList, data)

	standardClaims := make(map[string]any)
	_ = conv.Unmarshal(cfg, &standardClaims)

	result := make(jwt.MapClaims)
	for k, v := range mapClaims {
		result[k] = v
	}
	for k, v := range standardClaims {
		if _, ok := result[k]; ok {
			continue
		}
		result[k] = v
	}
	token := jwt.NewWithClaims(signingMethod, result)
	return token.SignedString([]byte(secretKey))
}

// JwtDecrypt RSA解密数据，secretKey必须是成对出现
func JwtDecrypt(secretKey string, cipherText string, data any, cfgList ...*JwtCfg) (string, error) {
	var jsonStr, err = jwtPayload(cipherText)
	if err != nil {
		log.Println(err)
	}

	if data == nil {
		return jsonStr, fmt.Errorf("data is nil")
	}
	pointType := reflect.TypeOf(data)
	if pointType.Kind() != reflect.Ptr {
		return jsonStr, fmt.Errorf("data must be a pointer")
	}

	token, err := jwt.ParseWithClaims(cipherText, &jwt.MapClaims{}, func(tk *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return jsonStr, err
	}
	if decodeToken, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		var useEncryptList []string
		if len(cfgList) > 0 && cfgList[0] != nil {
			useEncryptList = cfgList[0].EncryptJsonKeyList
		}
		dataStr := jwtDecryptDataFromMap(decodeToken, secretKey, useEncryptList)
		//直接获取data的真实数据
		dataReal := gjson.Get(dataStr, defaultDataKey)
		if dataReal.Exists() {
			dataStr = dataReal.String()
		} else {
			//需要移除自带的
			dataStr = removeStandardClaimsKey(dataStr)
		}
		err = conv.Unmarshal(dataStr, data)
		if err != nil {
			return dataStr, err
		}
		return dataStr, nil
	}
	return jsonStr, fmt.Errorf("token wrong")
}

func removeStandardClaimsKey(dataString string) string {
	if len(standardClaimsKeys) == 0 {
		initKeys()
	}
	if len(standardClaimsKeys) > 0 {
		for _, key := range standardClaimsKeys {
			if gjson.Get(dataString, key).Exists() {
				dataString, _ = sjson.Delete(dataString, key)
			}
		}
	}
	return dataString
}
func initKeys() {
	sc := jwt.StandardClaims{
		Audience:  "demo",
		ExpiresAt: 5,
		Id:        "demo",
		IssuedAt:  5,
		Issuer:    "demo",
		NotBefore: 5,
		Subject:   "demo",
	}
	mapSc := make(map[string]any)
	_ = conv.Unmarshal(sc, &mapSc)
	keys := lo.Keys(mapSc)
	if len(keys) > 0 {
		standardClaimsKeys = keys
	}
}

// jwtPayload 仅解析 JWT 载荷数据
func jwtPayload(jwtToken string) (string, error) {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("无效的 JWT 令牌格式")
	}
	payloadPart := parts[1]
	padding := len(payloadPart) % 4
	if padding > 0 {
		payloadPart += strings.Repeat("=", 4-padding)
	}
	payloadBytes, err := base64.URLEncoding.DecodeString(payloadPart)
	if err != nil {
		return "", fmt.Errorf("载荷 Base64 解码失败：%v", err)
	}
	return string(payloadBytes), nil
}

func jwtEncryptDataToMap(secretKey string, useEncryptList []string, data any) map[string]any {
	dataStr := conv.String(data)

	if len(useEncryptList) > 0 {
		inAll := utils.Contains(useEncryptList, ".") // 如果包含所有的话，则部份的就不用管了
		if inAll {
			resultStr, err := ConfigEncryptSecret(dataStr, secretKey)
			if err == nil {
				return map[string]any{defaultDataKey: resultStr}
			}
		} else {
			lo.ForEach(useEncryptList, func(item string, index int) {
				result := gjson.Get(dataStr, item)
				if result.Exists() {
					resultStr, err := ConfigEncryptSecret(result.String(), secretKey)
					if err == nil {
						dataStr1, err := sjson.Set(dataStr, item, resultStr)
						if err == nil {
							dataStr = dataStr1
						}
					}
				}
			})
		}
	}

	dataType := reflect.TypeOf(data)
	oneData, err := conv.ConvertForType(dataType, dataStr)
	if err == nil {
		data = oneData
	} else {
		data = dataStr
	}

	mapClaims := make(map[string]any)
	err = conv.Unmarshal(data, &mapClaims)
	if err != nil || len(mapClaims) == 0 {
		return map[string]any{defaultDataKey: data}
	}
	return mapClaims
}
func jwtDecryptDataFromMap(decodeToken *jwt.MapClaims, secretKey string, useEncryptList []string) string {
	dataStr := conv.String(decodeToken)

	if len(useEncryptList) == 0 {
		return dataStr
	}

	inAll := utils.Contains(useEncryptList, ".") // 如果包含所有的话，则部份的就不用管了
	if inAll {
		allDataString := gjson.Get(dataStr, defaultDataKey).String()
		resultStr, err := ConfigDecryptSecret(allDataString, secretKey)
		if err == nil {
			dataStrTemp, err := sjson.Set(dataStr, defaultDataKey, resultStr)
			if err == nil {
				dataStr = dataStrTemp
			}
		}
	} else {
		lo.ForEach(useEncryptList, func(item string, index int) {
			result := gjson.Get(dataStr, item)
			if result.Exists() {
				resultStr, err := ConfigDecryptSecret(result.String(), secretKey)
				if err == nil {
					dataStr1, err := sjson.Set(dataStr, item, resultStr)
					if err == nil {
						dataStr = dataStr1
					}
				}
			}
		})
	}
	return dataStr
}
