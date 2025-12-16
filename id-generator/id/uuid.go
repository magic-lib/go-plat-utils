package id

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/go-dev-frame/sponge/pkg/krand"
	gguid "github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/marspere/goencrypt"
	gouuid "github.com/nu7hatch/gouuid"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetXId 20字符 id 生成器,如：
// d50fukm29if2e0rt3n60
func XId() string {
	guid := xid.New()
	return guid.String()
}

// ObjectId 24字符 id 生成器,如：
// 67c9d12e050656d79bb0c630
func ObjectId() string {
	return primitive.NewObjectID().Hex()
}

// Snowflake 24字符 id 生成器,如：
// 1897690027819798528
// 符号位	时间戳位	节点ID位		序列号位
// 1位		41位	10位		12位
// 0正数		69年	1024个节点	095个序列
func Snowflake(nodeInt ...int64) snowflake.ID {
	var nodeTemp int64 = 1
	if len(nodeInt) > 0 {
		nodeTemp = nodeInt[0]
	}
	if nodeTemp <= 0 || nodeTemp >= 1024 {
		nodeTemp = 1
	}
	node, _ := snowflake.NewNode(nodeTemp)
	return node.Generate()
}

// KrandId 24字符 id 生成器,如：
// 1897690027819798528
func KrandId() string {
	return krand.NewStringID()
}

// KrandIdInt 24字符 id 生成器,如：
// 1897690027819798528
func KrandIdInt() int64 {
	return krand.NewID()
}

// UUID 36字符 id 生成器,如：
// 98c04b4b-b865-47e6-b72b-03fe04389fdd
func UUID() string {
	return gguid.New().String()
}

// ShotUUID 24字符 id 生成器,如：
// k8PsWEDmsYUAbmeHcyjfeB
func ShotUUID() string {
	return shortuuid.New()
}

// getUUIDv7 36字符 id 生成器,如：0195d052-4c80-7217-ad19-1acb84b04d4f
func getUUIDv7() string {
	id, _ := gguid.NewV7() //时间排序
	return id.String()
}

// NewUUID 新建uuid
// 版本1: 基于时间戳和MAC地址
// 版本2: DCE安全UUID
// 版本3: 基于MD5哈希和命名空间
// 版本4: 基于随机数（最常用）
// 版本5: 基于SHA-1哈希和命名空间
func NewUUID() string {
	uuidGenerators := []func() (string, error){ // 定义一个切片，存储不同的UUID生成函数
		func() (string, error) {
			uuids, err := gguid.NewUUID() //版本1
			if err != nil {
				return "", err
			}
			return uuids.String(), nil
		},
		func() (string, error) { return gguid.New().String(), nil }, // 版本4
		func() (string, error) { return getUUIDv7(), nil },          // 使用gguid的另一个生成方法
		func() (string, error) {
			uuidTemp, err := gouuid.NewV4()
			if err != nil {
				return "", err
			}
			return uuidTemp.String(), nil
		}, // 使用gouuid生成UUID
		func() (string, error) {
			uuidTemp := GeneratorBase32()
			if uuidTemp == "" {
				uuidTemp = XId()
			}
			return GetUUID(uuidTemp), nil
		}, // 使用sonyflake生成UUID
	}

	for _, generatorExec := range uuidGenerators { // 遍历每个UUID生成函数
		uuidStr, err := generatorExec()  // 尝试生成UUID
		if err == nil && uuidStr != "" { // 如果没有错误且UUID字符串不为空
			return uuidStr // 返回生成的UUID字符串
		}
	}

	return "" // 如果所有方法都失败，返回空字符串
}

// GetUUID 获取uuid格式串
func GetUUID(s string) string {
	uuidTemp, err := goencrypt.MD5(s)
	if len(uuidTemp) != 32 || err != nil {
		return ""
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", uuidTemp[0:8], uuidTemp[8:12], uuidTemp[12:16], uuidTemp[16:20], uuidTemp[20:])
}
