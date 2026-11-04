// Package compress TODO
package compress

import (
	"bytes"
	"github.com/andybalholm/brotli"
	"io"
)

// BrCompress 读取路径并生成一个压缩包
// Content-Encoding: br
func BrCompress(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := brotli.NewWriterLevel(&buf, brotli.BestCompression)
	_, err := writer.Write(input)
	if err != nil {
		_ = writer.Close()
		return nil, err
	}
	// 必须 Close 写出结束块后再取数据：原来在 defer Close 之前返回，压缩流缺少 ISLAST 块，严格的解码端会报 unexpected EOF
	if err = writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BrUnCompress 读取路径并生成一个压缩包
func BrUnCompress(comData []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(comData))
	return io.ReadAll(reader)
}
