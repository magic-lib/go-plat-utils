package jsoniter

import (
	"fmt"
	jsoniter2 "github.com/json-iterator/go"
	"testing"
)

type TestStruct struct {
	FuncCall any    `json:"funcs"`
	Name     string `json:"name"`
}

func TestUnmarshalFunc(t *testing.T) {
	json := `{"funcs":null, "name":"test"}`
	var testStruct TestStruct
	err := Unmarshal([]byte(json), &testStruct)
	fmt.Println(err, testStruct)

	err = jsoniter2.Unmarshal([]byte(json), &testStruct)

	fmt.Println(err, testStruct)

}

func TestMarshalFunc(t *testing.T) {
	var testStruct = TestStruct{
		FuncCall: func() string {
			return "aaaa"
		},
		Name: "test",
	}
	json, err := Marshal(testStruct)
	fmt.Println(err, string(json))

	json, err = jsoniter2.Marshal(testStruct)

	fmt.Println(err, string(json))
}
