package crypto

import (
	"encoding/base64"
	"encoding/hex"
)

type EnDeCoder interface {
	Encode(plainText []byte) string
	Decode(cipherText string) ([]byte, error)
}

type HexCoder struct {
}

func (c *HexCoder) Encode(plainText []byte) string {
	return hex.EncodeToString(plainText)
}

func (c *HexCoder) Decode(cipherText string) ([]byte, error) {
	return hex.DecodeString(cipherText)
}

type Base64Coder struct {
}

func (c *Base64Coder) Encode(plainText []byte) string {
	return base64.StdEncoding.EncodeToString(plainText)
}

func (c *Base64Coder) Decode(cipherText string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(cipherText)
}
