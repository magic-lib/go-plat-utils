package utils

import (
	dockerJson "github.com/docker/go/canonical/json"
	"github.com/magic-lib/go-plat-utils/id-generator/id"
)

// UniqueJsonString
// 顺序不同 / 格式不同 → 只要内容一样，返回相同唯一string
func UniqueJsonString(data any) (string, error) {
	canonicalBytes, err := dockerJson.MarshalCanonical(data)
	if err != nil {
		return "", err
	}
	return string(canonicalBytes), nil
}
func UniqueJsonId(data any) (string, error) {
	jsonString, err := UniqueJsonString(data)
	if err != nil {
		return "", err
	}
	return id.GetUUID(jsonString), nil
}
