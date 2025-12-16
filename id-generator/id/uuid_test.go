package id_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/id-generator/id"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestAnyToBool(t *testing.T) {
	for i := 0; i < 20; i++ {
		ss := id.Generator(700)
		fmt.Println(ss)
	}
	fmt.Println("ok")
}

func TestGeneratorId(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"XId", []any{}, []any{true}, func() bool {
			ids := id.XId()
			fmt.Println(ids) //d50g1de29if2gokrg1og
			return true
		}},
		{"ObjectId", []any{}, []any{true}, func() bool {
			ids := id.ObjectId()
			fmt.Println(ids) //694100b5ee0ffe9d049396e4
			return true
		}},
		{"Snowflake", []any{}, []any{true}, func() bool {
			ids := id.Snowflake()
			fmt.Println(ids.String()) //2000820275230281728
			fmt.Println(ids.Int64())  //2000820275230281728
			fmt.Println(ids.Base2())  //1101111000100010101100001010110000100000000000001000000000000
			fmt.Println(ids.Base32()) //bztnsnsnyyryy
			fmt.Println(ids.Base36()) //f78v2ckken0g
			return true
		}},
		{"KrandId", []any{}, []any{true}, func() bool {
			ids := id.KrandId()
			fmt.Println(ids) //18819efeddc17a3e
			return true
		}},
		{"KrandIdInt", []any{}, []any{true}, func() bool {
			ids := id.KrandIdInt()
			fmt.Println(ids) //1765867346385941939
			return true
		}},
		{"UUID", []any{}, []any{true}, func() bool {
			ids := id.UUID()
			fmt.Println(ids) //f82715fa-68de-46b1-af19-772cc1b09806
			return true
		}},
		{"ShotUUID", []any{}, []any{true}, func() bool {
			ids := id.ShotUUID()
			fmt.Println(ids) //FkqjzpKcHVPnKLmuRrn9MD
			return true
		}},
		{"NewUUID", []any{}, []any{true}, func() bool {
			ids := id.NewUUID()
			fmt.Println(ids) //6285de18-da4a-11f0-bed7-52f61404d6c2
			return true
		}},
	}
	utils.TestFunction(t, testCases, nil)
}
