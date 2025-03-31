package crypto

import (
	"bytes"
	"encoding/gob"
)

// GobEncode 序列化一个对象
func GobEncode(s interface{}, en ...EnDeCoder) (string, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	err := enc.Encode(s)
	if err != nil {
		return "", err
	}

	var enCode EnDeCoder = new(Base64Coder)
	if en != nil && len(en) > 0 {
		enCode = en[0]
	}

	return enCode.Encode(buf.Bytes()), nil
}

// GobDecode 反序列化一个对象
func GobDecode(s string, data interface{}, de ...EnDeCoder) error {
	var b bytes.Buffer

	var deCode EnDeCoder = new(Base64Coder)
	if de != nil && len(de) > 0 {
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
