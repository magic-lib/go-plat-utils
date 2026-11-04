package crypto

// XorEncode 异或加密
func XorEncode(msg string, key int, en ...EnDeCoder) string {
	byteList := []byte(msg)
	pwd := make([]byte, len(byteList))
	for i := 0; i < len(byteList); i++ {
		pwd[i] = byteList[i] ^ byte(key)
	}

	var enCode EnDeCoder = new(Base64Coder)
	if len(en) > 0 && en[0] != nil { // 显式传入 nil 时使用默认编码，否则会空指针 panic
		enCode = en[0]
	}

	return enCode.Encode(pwd)
}

// XorDecode 异或解密
func XorDecode(msg string, key int, de ...EnDeCoder) (string, error) {
	var deCode EnDeCoder = new(Base64Coder)
	if len(de) > 0 && de[0] != nil { // 显式传入 nil 时使用默认编码，否则会空指针 panic
		deCode = de[0]
	}

	pwdList, err := deCode.Decode(msg)
	if err != nil {
		return "", err
	}
	old := make([]byte, len(pwdList))
	for i := 0; i < len(pwdList); i++ {
		old[i] = pwdList[i] ^ byte(key)
	}
	return string(old), nil
}
