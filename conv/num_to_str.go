package conv

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// FormatNumber 转换数字为不同普通格式
func FormatNumber[T cond.Number](format string, n T, tag ...language.Tag) string {
	if !cond.IsNumeric(n) {
		return String(n)
	}
	if format == "" {
		if cond.IsInteger(n) { //整数
			format = "%d"
		} else { //小数
			format = "%f"
		}
	}
	if len(tag) == 0 {
		return fmt.Sprintf(format, n)
	}
	p := message.NewPrinter(tag[0])
	return p.Sprintf(format, n)
}
