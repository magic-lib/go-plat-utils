package crypto

// AesEncrypt 对字符串对称加密，输入输出都为base64字符
func AesEncrypt(origString string, keyBase64 string, en ...EnDeCoder) (string, error) {
	if origString == "" {
		return "", nil
	}
	origData := []byte(origString)

	key := getAesKeyFromBase64(keyBase64)
	iv := ""

	encrypted, err := aesEncrypt(origData, key, iv)
	if err != nil {
		return "", err
	}

	var enCode EnDeCoder = new(Base64Coder)
	if en != nil && len(en) > 0 {
		enCode = en[0]
	}

	return enCode.Encode(encrypted), nil
}

// AesDecrypt 对字符串对称解密，输入输出都为base64字符
func AesDecrypt(encodeString string, keyBase64 string, de ...EnDeCoder) (string, error) {
	if encodeString == "" {
		return "", nil
	}
	key := getAesKeyFromBase64(keyBase64)
	iv := ""

	var deCode EnDeCoder = new(Base64Coder)
	if de != nil && len(de) > 0 {
		deCode = de[0]
	}
	encodeByte, err := deCode.Decode(encodeString)
	if err != nil {
		return "", err
	}

	decodeByte, err := aesDecrypt(encodeByte, key, iv)
	if err != nil {
		return "", err
	}
	return string(decodeByte), nil
}
