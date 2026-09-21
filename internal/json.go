package internal

// SimpleCheckJson 简单检查 text 是否「像」一个合法的 JSON 串（顶层必须是对象或数组），
// 用于在调用 json.Unmarshal 之前做低成本预筛，避免明显非 JSON 的输入带来解析开销。
//
// 特性：
//   - 单次线性扫描，O(n)，不解析、不构建 AST、除浅嵌套栈外几乎零分配；
//   - **无假阴性**：任何合法的顶层对象/数组 JSON 必定返回 true；
//   - 存在假阳性（如 {,}），即返回 true 不代表 Unmarshal 一定成功，仍需 Unmarshal 最终确认。
//
// 检查项：首尾空白裁剪、顶层必须是 {..} 或 [..]、括号配对、字符串与转义、非法控制字符、
// 顶层多余内容、起始/尾随逗号、重复逗号、非期待位置的值、对象外/数组内的冒号。
//
// 注意：JSON 顶层标量（123、"str"、true、null）返回 false，与 IsJson 的口径保持一致。
func SimpleCheckJsonArrayOrObject(text string) bool {
	// JSON 规范（及 encoding/json）认可的空白只有空格、制表符、换行、回车四种，
	// 因此这里不使用 strings.TrimSpace，以免放过 Unicode 空白后又被 Unmarshal 拒绝。
	start, end := 0, len(text)
	for start < end && isJsonSpace(text[start]) {
		start++
	}
	for end > start && isJsonSpace(text[end-1]) {
		end--
	}
	if end-start < 2 { // 最短的合法顶层值是 {} / []
		return false
	}

	open := text[start]
	if open != '{' && open != '[' {
		return false
	}

	// 对应的闭合符
	closeChar := byte('}')
	if open == '[' {
		closeChar = ']'
	}
	if text[end-1] != closeChar {
		return false
	}

	lastSig := jsonStateNone
	inString := false
	escaped := false
	spaceAfterValue := false // 值是否已结束并紧跟了空白，用于把 "1 2" 与 "true" 区分开

	// 嵌套深度通常很浅，用栈上数组承载括号栈以避免堆分配；
	// 超过容量时 append 会自动扩容，结果不变，只是退化成一次分配。
	var stackArr [16]byte
	stack := stackArr[:0]

	for i := start; i < end; i++ {
		b := text[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}
			switch b {
			case '\\':
				escaped = true
			case '"':
				inString = false
				lastSig = jsonStateValue
				spaceAfterValue = false
			default:
				if b < 0x20 {
					return false // 字符串内不得出现未转义的控制字符
				}
			}
			continue
		}

		if b <= ' ' { // ASCII 中 <= 0x20 的只有空白和控制字符，空格(0x20)也在其中
			if !isJsonSpace(b) {
				return false // 串外只允许四种空白
			}
			if lastSig == jsonStateValue {
				spaceAfterValue = true // 空白隔断了当前值，其后若还有同类字符应视为非法的新值
			}
			continue // 空白本身不改变 lastSig
		}

		switch b {
		case '"':
			// 值可能出现在 { 或 [ 之后、冒号/逗号之后，否则属于多余内容
			if !isExpecting(lastSig) {
				return false
			}
			inString = true

		case '{', '[':
			if !isExpecting(lastSig) {
				return false
			}
			stack = append(stack, b)
			lastSig = b

		case '}', ']':
			if lastSig == jsonStateNone || lastSig == ',' || lastSig == ':' {
				return false // 空的首元素是非法；尾逗号、尾冒号同样非法
			}
			if len(stack) == 0 {
				return false // 多余的闭合符
			}
			top := stack[len(stack)-1]
			wantClose := byte('}')
			if top == '[' {
				wantClose = ']'
			}
			if b != wantClose {
				return false // 括号配对不匹配
			}
			stack = stack[:len(stack)-1]
			lastSig = b

		case ',':
			// 逗号只能出现在值或闭合符之后
			if lastSig != jsonStateValue && lastSig != '}' && lastSig != ']' {
				return false
			}
			lastSig = ','

		case ':':
			// 冒号只能出现在对象中，且必须紧跟 key
			if lastSig != jsonStateValue || len(stack) == 0 || stack[len(stack)-1] != '{' {
				return false
			}
			lastSig = ':'

		default:
			// 数字、true/false/null 等字面量字符
			if isExpecting(lastSig) { // 一个新值的开始
				lastSig = jsonStateValue
				spaceAfterValue = false
				continue
			}
			// 多字符字面量（如 null、true、1.25e+3）的后续字符应保持 jsonStateValue；
			// 但值已经被空白隔断后又出现同类字符（如 [1 2]）则是多余的第二个值
			if lastSig == jsonStateValue && !spaceAfterValue {
				continue
			}
			return false
		}
	}

	return len(stack) == 0 && !inString
}

// lastSig 记录上一个有语义的字符，用于在扫描过程中判断「当前位置是否期待一个新值」。
const (
	jsonStateNone  byte = 0   // 尚未遇到任何有效字符
	jsonStateValue byte = 'V' // 刚读完一个完整的字面量值（含字符串、数字、true/false/null）
)

// isExpecting 判断 lastSig 之后的下一个位置是否期待一个新值
func isExpecting(lastSig byte) bool {
	switch lastSig {
	case jsonStateNone, // 整串开头，期待第一个值
		',', // 期待下一个元素
		':', // 期待 key 对应的值
		'{', // 期待第一个 key
		'[': // 期待第一个元素
		return true
	}
	return false
}

// isJsonSpace 判断是否是 JSON 认可的空白字符
func isJsonSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
