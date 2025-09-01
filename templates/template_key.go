package templates

import (
	"github.com/tidwall/gjson"
	"strings"
)

// JsonGet 递归解析路径，尝试所有可能的.含义（键名一部分/层级分隔符）
// jsonStr: 输入的JSON字符串
// path: 待解析的路径（如"name.age.age1"）
// 返回: 找到的第一个匹配值（按优先级尝试）
func JsonGet(jsonStr, path string) (gjson.Result, bool) {
	// 解析JSON根节点
	root := gjson.Parse(jsonStr)
	if !root.IsObject() {
		if path == "" || path == "." {
			return root, true //如果没有路径，则返回根节点
		}
		return root, false // 非对象结构，又存在path，直接返回不存在
	}

	currNode, exists := longestKeyMatch(root, path, "")
	if exists && currNode.Exists() {
		return currNode, true
	}
	return getResultValue(root, path)
}

// getResultValueByEscapedKey：将key中的.全转义（视为键名一部分），获取值
func getResultValueByEscapedKey(currentNode gjson.Result, key string) (gjson.Result, bool) {
	var value gjson.Result
	// 处理包含点的键
	if strings.Contains(key, ".") {
		// 全转义点字符
		escapedKey := strings.ReplaceAll(key, ".", "\\.")
		value = currentNode.Get(escapedKey)
		return value, value.Exists()
	}
	value = currentNode.Get(key)
	return value, value.Exists()
}

// getResultValue：先尝试全转义（.作为键名），失败则直接解析（.作为层级）
func getResultValue(currentNode gjson.Result, key string) (gjson.Result, bool) {
	if value, exists := getResultValueByEscapedKey(currentNode, key); exists {
		return value, true
	}
	// 含有.，但不是完全的EscapedKey，则尝试普通方式解析
	value := currentNode.Get(key)
	return value, value.Exists()
}

// longestKeyMatch 按最长键名优先匹配：从完整路径开始，逐级缩短键名，直到找到匹配值
// currentNode: 当前JSON节点（初始为根节点）
// path: 完整路径（如"name.age.age1"）
func longestKeyMatch(currentNode gjson.Result, shortenedKey string, remainingPath string) (gjson.Result, bool) {
	if shortenedKey == "" || shortenedKey == "." {
		return currentNode, true //如果没有路径，则返回根节点
	}

	// 1. 尝试当前最长键名（完整路径，.作为键名一部分）
	currValue, exists := getResultValueByEscapedKey(currentNode, shortenedKey)
	if exists {
		if remainingPath == "" {
			return currValue, true
		}
		return longestKeyMatch(currValue, remainingPath, "")
	}

	// 2. 找到最后一个.的位置，缩短路径（去掉最后一级）
	lastDotIndex := strings.LastIndex(shortenedKey, ".")
	if lastDotIndex == -1 {
		// 路径中没有.，无法再缩短，尝试直接作为层级键名
		return currentNode, false
	}

	// 3. 缩短路径：保留lastDotIndex前的部分作为新键名，剩余部分作为下一层级路径
	shortenedKeyNew := shortenedKey[:lastDotIndex]    // 如"name.age.age1" → "name.age"
	shortenedKeyLeft := shortenedKey[lastDotIndex+1:] // 如"name.age.age1" → "age1"
	remainingPathList := append(strings.Split(shortenedKeyLeft, "."), strings.Split(remainingPath, ".")...)
	remainingPathNew := strings.Join(remainingPathList, ".")

	return longestKeyMatch(currentNode, shortenedKeyNew, remainingPathNew)
}
