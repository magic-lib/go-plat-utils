// Package param 获取参数
package param

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"mime/multipart"
	"net/http"
	"strings"
)

// MultipartFile 上传文件的信息
type MultipartFile struct {
	FieldName   string         // 表单字段名，如 "file"
	FileName    string         // 原始文件名，如 "collection_order_template.xlsx"
	Size        int64          // 文件大小（字节）
	ContentType string         // MIME 类型，如 "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	File        multipart.File // 文件流，实现了 io.Reader，可直接读取
}

// multipartFormData 解析后的 multipart 表单数据
type multipartFormData struct {
	Fields map[string]string           // 普通表单字段，key=字段名, value=字段值
	Files  map[string][]*MultipartFile // 上传的文件列表
}

// - *multipart.Form: 与 r.MultipartForm 完全相同的结构
func parseMultipartBytes(body []byte, boundary string, maxMemory int64) (*multipart.Form, error) {
	// 1. 将 []byte 包装成 io.Reader
	reader := bytes.NewReader(body)

	// 2. 创建 multipart.Reader（需要 boundary）
	mr := multipart.NewReader(reader, boundary)

	// 3. 调用 ReadForm，返回的 *multipart.Form 与 r.MultipartForm 类型完全一致
	form, err := mr.ReadForm(maxMemory)
	if err != nil {
		return nil, fmt.Errorf("解析 multipart 字节失败: %w", err)
	}

	return form, nil
}

// getBoundaryFromHeader 从 Content-Type header 中提取 boundary
func getBoundaryFromHeader(h http.Header) (string, error) {
	if len(h) == 0 {
		return "", fmt.Errorf("request is nil")
	}

	contentType := h.Get("Content-Type")
	if contentType == "" {
		return "", fmt.Errorf("Content-Type header is empty")
	}

	// 检查是否为 multipart 类型
	if !strings.HasPrefix(contentType, "multipart/") {
		return "", fmt.Errorf("Content-Type is not multipart: %s", contentType)
	}

	// 查找 boundary 参数
	parts := strings.Split(contentType, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "boundary=") {
			boundary := strings.TrimPrefix(part, "boundary=")
			boundary = strings.Trim(boundary, "\"' ") // 去除可能的引号
			if boundary == "" {
				return "", fmt.Errorf("boundary is empty")
			}
			return boundary, nil
		}
	}
	return "", fmt.Errorf("boundary not found in Content-Type: %s", contentType)
}
func parseMultipartForm(h http.Header, body []byte) (*multipartFormData, error) {
	if len(h) == 0 {
		return nil, fmt.Errorf("request 不能为空")
	}
	if body == nil {
		return nil, fmt.Errorf("body 不能为空")
	}

	boundary, err := getBoundaryFromHeader(h)
	if err != nil {
		return nil, fmt.Errorf("获取 boundary 失败: %w", err)
	}

	multipartForm, err := parseMultipartBytes(body, boundary, 32*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("解析 multipart 字节失败: %w", err)
	}

	result := &multipartFormData{
		Fields: make(map[string]string),
		Files:  make(map[string][]*MultipartFile, 0),
	}

	// 第二步：提取所有普通表单字段
	if multipartForm != nil && multipartForm.Value != nil {
		for key, values := range multipartForm.Value {
			if len(values) > 0 {
				result.Fields[key] = values[0] // 取第一个值
			}
		}
	}

	// 第三步：提取所有上传文件
	if multipartForm != nil && multipartForm.File != nil {
		allFile := make([]*MultipartFile, 0)
		for fieldName, fileHeaders := range multipartForm.File {
			for _, header := range fileHeaders {
				file, err := header.Open()
				if err != nil {
					return nil, fmt.Errorf("打开上传文件 [%s] 失败: %w", header.Filename, err)
				}
				allFile = append(allFile, &MultipartFile{
					FieldName:   fieldName,
					FileName:    header.Filename,
					Size:        header.Size,
					ContentType: header.Header.Get("Content-Type"),
					File:        file,
				})
			}
		}
		if len(allFile) > 0 {
			lo.ForEach(allFile, func(item *MultipartFile, index int) {
				if _, ok := result.Files[item.FieldName]; !ok {
					result.Files[item.FieldName] = make([]*MultipartFile, 0)
				}
				result.Files[item.FieldName] = append(result.Files[item.FieldName], item)
			})
		}

	}

	return result, nil
}
