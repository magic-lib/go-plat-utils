package id

import (
	"fmt"
	gguid "github.com/google/uuid"
	"github.com/marspere/goencrypt"
	gouuid "github.com/nu7hatch/gouuid"
	"github.com/rs/xid"
)
import "github.com/google/uuid"

// GetXId 20字符 id 生成器,如：cvhmhh6s295a4l56g4a0
func GetXId() string {
	guid := xid.New()
	return guid.String()
}

// getUUIDv7 36字符 id 生成器,如：0195d052-4c80-7217-ad19-1acb84b04d4f
func getUUIDv7() string {
	id, _ := uuid.NewV7()
	return id.String()
}

// NewUUID 新建uuid
func NewUUID() string {
	uuidGenerators := []func() (string, error){ // 定义一个切片，存储不同的UUID生成函数
		func() (string, error) {
			uuids, err := gguid.NewUUID()
			if err != nil {
				return "", err
			}
			return uuids.String(), nil
		}, // 使用gguid生成UUID
		func() (string, error) { return gguid.New().String(), nil }, // 使用gguid的另一个生成方法
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
				uuidTemp = GetXId()
			}
			return GetUUID(uuidTemp), nil
		}, // 使用sonyflake生成UUID
	}

	for _, generator := range uuidGenerators { // 遍历每个UUID生成函数
		uuidStr, err := generator()      // 尝试生成UUID
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
