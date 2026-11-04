//+build go1.8

package test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

func Test_new_encoder(t *testing.T) {
	should := require.New(t)
	buf1 := &bytes.Buffer{}
	encoder1 := json.NewEncoder(buf1)
	encoder1.SetEscapeHTML(false)
	encoder1.Encode([]int{1})
	should.Equal("[1]\n", buf1.String())
	buf2 := &bytes.Buffer{}
	encoder2 := jsoniter.NewEncoder(buf2)
	encoder2.SetEscapeHTML(false)
	encoder2.Encode([]int{1})
	should.Equal("[1]\n", buf2.String())
}

func Test_string_encode_with_std_without_html_escape(t *testing.T) {
	api := jsoniter.Config{EscapeHTML: false}.Froze()
	should := require.New(t)
	for i := 0; i < utf8.RuneSelf; i++ {
		input := string([]byte{byte(i)})
		buf := &bytes.Buffer{}
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(input)
		should.Nil(err)
		stdOutput := buf.String()
		stdOutput = normalizeStdEscape(stdOutput[:len(stdOutput)-1])
		jsoniterOutputBytes, err := api.Marshal(input)
		should.Nil(err)
		jsoniterOutput := string(jsoniterOutputBytes)
		should.Equal(stdOutput, jsoniterOutput)
	}
}

// normalizeStdEscape 将标准库输出中的 \b、\f 短转义统一为 jsoniter 使用的 \u0008、\u000c。
// 新版 Go 的 encoding/json 改用短转义，两种写法均为合法 JSON 且解析结果相同。
// 按转义序列逐个解析，避免误改 \\b（转义的反斜杠后跟字母 b）。
func normalizeStdEscape(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			sb.WriteByte(s[i])
			continue
		}
		switch s[i+1] {
		case 'b':
			sb.WriteString(`\u0008`)
		case 'f':
			sb.WriteString(`\u000c`)
		default:
			sb.WriteByte(s[i])
			sb.WriteByte(s[i+1])
		}
		i++
	}
	return sb.String()
}
