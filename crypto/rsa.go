package crypto

// EncryptRSA RSA加密数据，key必须是成对出现
func EncryptRSA(oneKeyStr string, message string) (string, error) {
	en := new(Base64Coder)

	rsa := new(RSASecurity)
	err := rsa.SetPublicAndPrivateKey(oneKeyStr, "")
	if err != nil {
		err = rsa.SetPublicAndPrivateKey("", oneKeyStr)
		if err != nil {
			return "", err
		}
		return rsa.PriKeyEncrypt(message, en)
	}
	return rsa.PubKeyEncrypt(message, en)
}

// DecryptRSA RSA解密数据，key必须是成对出现
func DecryptRSA(otherKeyStr string, cipherText string) (string, error) {
	de := new(Base64Coder)

	rsa := new(RSASecurity)
	err := rsa.SetPublicAndPrivateKey("", otherKeyStr)
	if err != nil {
		err = rsa.SetPublicAndPrivateKey(otherKeyStr, "")
		if err != nil {
			return "", err
		}
		return rsa.PubKeyDecrypt(cipherText, de)
	}
	return rsa.PriKeyDecrypt(cipherText, de)
}
