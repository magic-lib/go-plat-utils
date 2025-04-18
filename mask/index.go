package mask

import (
	"strings"
)

// Character 屏蔽字符码
func Character(s string, front, end int, maskCode string) string {
	// 处理边界情况：如果 front 或 end 为负数，将其置为 0
	if front <= 0 {
		front = 0
	}
	if end <= 0 {
		end = 0
	}
	// 如果掩码字符为空，默认使用 *
	if maskCode == "" {
		maskCode = "*"
	}

	runes := []rune(s)
	length := len(runes)
	// 如果字符串长度小于等于 front + end，直接返回原字符串
	if length <= front+end {
		return s
	}
	// 计算需要掩码的字符数量
	maskedCount := length - front - end
	// 初始化一个足够大的切片来存储结果
	result := make([]rune, 0, length)
	// 添加字符串开头的字符
	result = append(result, runes[:front]...)

	maskRunes := []rune(maskCode)
	maskLen := len(maskRunes)

	if maskLen >= maskedCount {
		result = append(result, maskRunes[:maskedCount]...)
	} else {
		i := 0
		for i < maskedCount {
			for _, one := range maskRunes {
				result = append(result, one)
				i++
				if i >= maskedCount {
					break
				}
			}
		}
	}
	result = append(result, runes[front+maskedCount:]...)
	return string(result)
}

// Phone 隐去手机号中间 4 位地区码, 如 155****8888
func Phone(phone string) string {
	if len(phone) >= 8 {
		return Character(phone, 3, 4, "*")
	}
	return Character(phone, 3, 0, "*")
}

// Email 隐藏邮箱ID的中间部分 zhang@go-mall.com ---> z***g@go-mall.com
func Email(address string) string {
	atIndex := strings.LastIndex(address, "@")
	if atIndex == -1 {
		return address
	}
	id := address[0:atIndex]
	domain := address[atIndex:]

	padNumber := 2
	if len(id) <= 4 {
		padNumber = 1
	}
	return Character(id, padNumber, padNumber, "*") + domain
}

// RealName 保留姓名首末位 如：张三--->张* 赵丽颖--->赵*颖 欧阳娜娜--->欧**娜
func RealName(realName string) string {
	realNameRunes := []rune(realName)
	if len(realNameRunes) <= 1 {
		return realName
	}
	if len(realNameRunes) == 2 {
		return Character(realName, 1, 0, "*")
	}
	return Character(realName, 1, 1, "*")
}
