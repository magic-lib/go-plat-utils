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
	defer func(writer *brotli.Writer) {
		_ = writer.Close()
	}(writer)
	_, err := writer.Write(input)
	if err != nil {
		return nil, err
	}
	err = writer.Flush()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BrUnCompress 读取路径并生成一个压缩包
func BrUnCompress(comData []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(comData))
	return io.ReadAll(reader)
}
