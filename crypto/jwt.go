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

type jwtPayLoad struct {
	Data any `json:"data"`
	jwt.StandardClaims
}
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
	if cfg.ExpiresAt == 0 {
		cfg.ExpiresAt = time.Now().Add(time.Hour).Unix()
	}

	//Audience  string `json:"aud,omitempty"`
	//ExpiresAt int64  `json:"exp,omitempty"`
	//Id        string `json:"jti,omitempty"`
	//IssuedAt  int64  `json:"iat,omitempty"`
	//Issuer    string `json:"iss,omitempty"`
	//NotBefore int64  `json:"nbf,omitempty"`
	//Subject   string `json:"sub,omitempty"`

	if cfg.ExpiresAt == 0 {
		cfg.ExpiresAt = time.Now().Add(time.Hour).Unix()
	}
	if len(useEncryptList) > 0 {
		dataStr := conv.String(data)
		inAll := utils.Contains(useEncryptList, ".") // 如果包含所有的话，则部份的就不用管了
		if inAll {
			resultStr, err := ConfigEncryptSecret(dataStr, secretKey)
			if err == nil {
				dataStr = resultStr
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

		dataType := reflect.TypeOf(data)
		oneData, err := conv.ConvertForType(dataType, dataStr)
		if err == nil {
			data = oneData
		} else {
			data = dataStr
		}
	}

	c := jwtPayLoad{
		Data:           data,
		StandardClaims: *cfg,
	}
	token := jwt.NewWithClaims(signingMethod, c)
	return token.SignedString([]byte(secretKey))
}

// JwtDecrypt RSA解密数据，secretKey必须是成对出现
func JwtDecrypt(secretKey string, cipherText string, data any, cfgList ...*JwtCfg) (string, error) {
	var jsonStr, err = jwtPayload(cipherText)
	if err != nil {
		log.Println(err)
	} else {
		jsonStr = gjson.Get(jsonStr, "data").String()
	}

	if data == nil {
		return jsonStr, fmt.Errorf("data is nil")
	}
	pointType := reflect.TypeOf(data)
	if pointType.Kind() != reflect.Ptr {
		return jsonStr, fmt.Errorf("data must be a pointer")
	}

	token, err := jwt.ParseWithClaims(cipherText, &jwtPayLoad{}, func(tk *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return jsonStr, err
	}
	if decodeToken, ok := token.Claims.(*jwtPayLoad); ok && token.Valid {
		var useEncryptList []string
		if len(cfgList) > 0 && cfgList[0] != nil {
			useEncryptList = cfgList[0].EncryptJsonKeyList
		}

		if len(useEncryptList) > 0 {
			dataStr := conv.String(decodeToken.Data)
			inAll := utils.Contains(useEncryptList, ".") // 如果包含所有的话，则部份的就不用管了
			if inAll {
				resultStr, err := ConfigDecryptSecret(dataStr, secretKey)
				if err == nil {
					dataStr = resultStr
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
			decodeToken.Data = dataStr
		}
		retData := conv.String(decodeToken.Data)
		err = conv.Unmarshal(decodeToken.Data, data)
		if err != nil {
			return retData, err
		}
		return retData, nil
	}
	return jsonStr, fmt.Errorf("token wrong")
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
