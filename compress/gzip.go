// Package compress TODO
package compress

import (
	"bytes"
	"compress/gzip"
	"io"
)

// GZipCompress 读取路径并生成一个压缩包
func GZipCompress(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	// 写入数据
	if _, err = gz.Write(input); err != nil {
		_ = gz.Close()
		return nil, err
	}

	// 确保将缓冲区的内容刷新到gzip.Writer中
	if err = gz.Flush(); err != nil {
		_ = gz.Close()
		return nil, err
	}

	// 关闭gzip.Writer
	if err = gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GZipUnCompress 读取路径并生成一个压缩包
func GZipUnCompress(comData []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(comData))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
