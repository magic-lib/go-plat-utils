package crypto

import (
	"bytes"
	"encoding/gob"
)

// GobEncode 序列化一个对象
func GobEncode(s any, en ...EnDeCoder) (string, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	err := enc.Encode(s)
	if err != nil {
		return "", err
	}

	var enCode EnDeCoder = new(Base64Coder)
	if len(en) > 0 && en[0] != nil { // 显式传入 nil 时使用默认编码，否则会空指针 panic
		enCode = en[0]
	}

	return enCode.Encode(buf.Bytes()), nil
}

// GobDecode 反序列化一个对象
func GobDecode(s string, data any, de ...EnDeCoder) error {
	var b bytes.Buffer

	var deCode EnDeCoder = new(Base64Coder)
	if len(de) > 0 && de[0] != nil { // 显式传入 nil 时使用默认编码，否则会空指针 panic
		deCode = de[0]
	}

	old, err := deCode.Decode(s)
	if err != nil {
		return err
	}
	_, err = b.Write(old)
	if err != nil {
		return err
	}
	enc := gob.NewDecoder(&b)
	err = enc.Decode(data)
	if err != nil {
		return err
	}
	return nil
}
