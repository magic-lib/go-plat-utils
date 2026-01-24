package crypto

import (
	"google.golang.org/protobuf/proto"
)

// go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
// protoc --go_out=. person.proto

// ProtoEncode protobuf编码
func ProtoEncode(originData proto.Message) ([]byte, error) {
	return proto.Marshal(originData)
}

// ProtoDecode protobuf解码
func ProtoDecode(dataByte []byte, originData proto.Message) error {
	return proto.Unmarshal(dataByte, originData)
}
