// Package utils TODO
package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetAllFileContent 获取文件夹下所有文件的相对路径和文件内容
func GetAllFileContent(dest string) (map[string]string, error) {
	fileMap := make(map[string]string)
	_, err := os.Stat(dest)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			var content []byte
			content, err = os.ReadFile(path)
			if err != nil {
				return err
			}
			fileMap[path] = string(content)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	newFileMap := map[string]string{}
	for key, con := range fileMap {
		keyTemp := strings.TrimPrefix(key, dest+"/")
		newFileMap[keyTemp] = con
	}

	return newFileMap, nil
}

type fileLine struct {
	FileName string `json:"fileName"`
	FuncName string `json:"funcName"`
	Line     int    `json:"line"`
}

// GetRuntimeCallers 获取调用方法文件名和行号
func GetRuntimeCallers(baseFileName string, baseLine int, startInt int, length int) []*fileLine {
	fileList := make([]*fileLine, 0)

	baseIndex := 1
	maxLen := 100
	if baseFileName != "" {
		for i := 1; i < maxLen; i++ {
			pc, fileTemp, lineTemp, ok := runtime.Caller(i)
			if !ok {
				break
			}
			oneFile := &fileLine{
				FileName: fileTemp,
				FuncName: runtime.FuncForPC(pc).Name(),
				Line:     lineTemp,
			}
			if oneFile.FuncName == baseFileName ||
				oneFile.Line == baseLine {
				baseIndex = i
				break
			}
		}
	}
	for i := baseIndex + startInt; i < maxLen; i++ {
		pc, fileTemp, lineTemp, ok := runtime.Caller(i)
		if !ok {
			break
		}
		oneFile := &fileLine{
			FileName: fileTemp,
			FuncName: runtime.FuncForPC(pc).Name(),
			Line:     lineTemp,
		}
		fileList = append(fileList, oneFile)
		if length > 0 {
			if len(fileList) == length {
				break
			}
		}
	}

	return fileList
}
