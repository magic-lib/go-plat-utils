// +build go1.15
// remove these tests temporarily until https://github.com/golang/go/issues/38105 and
// https://github.com/golang/go/issues/38940 is fixed

package test

import (
	"encoding"
	"strings"
)

func init() {
	// map[stringKeyType]string 不参与对称性测试：key 底层为 string 且实现了 TextMarshaler 时，
	// 标准库直接使用字符串本身，jsoniter 调用 MarshalText，二者输出不同（已知差异）
	testCases = append(testCases,
		(*map[structKeyType]string)(nil),
	)
}

type stringKeyType string

func (k stringKeyType) MarshalText() ([]byte, error) {
	return []byte("MANUAL__" + k), nil
}

func (k *stringKeyType) UnmarshalText(text []byte) error {
	*k = stringKeyType(strings.TrimPrefix(string(text), "MANUAL__"))
	return nil
}

var _ encoding.TextMarshaler = stringKeyType("")
var _ encoding.TextUnmarshaler = new(stringKeyType)

type structKeyType struct {
	X string
}

func (k structKeyType) MarshalText() ([]byte, error) {
	return []byte("MANUAL__" + k.X), nil
}

func (k *structKeyType) UnmarshalText(text []byte) error {
	k.X = strings.TrimPrefix(string(text), "MANUAL__")
	return nil
}

var _ encoding.TextMarshaler = structKeyType{}
var _ encoding.TextUnmarshaler = &structKeyType{}
