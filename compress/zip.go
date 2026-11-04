// Package compress TODO
package compress

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateGzFile 读取路径并生成一个压缩包
func CreateGzFile(srcFilePath string, toGzFileName string) (fileContent []byte, err error) {
	if toGzFileName == "" {
		toGzFileName = strings.TrimRight(srcFilePath, string(os.PathSeparator))
		toGzFileName = fmt.Sprintf("%s.tar.gz", toGzFileName)
	}

	// 必须等 tar、gzip、文件都关闭（写入 tar 结束块和 gzip 尾部）后再读取，否则内容不完整
	if err = writeTarGz(srcFilePath, toGzFileName); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(toGzFileName)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// writeTarGz 将 srcFilePath 打包压缩到 toGzFileName，返回前按 tar、gzip、文件的顺序关闭
func writeTarGz(srcFilePath string, toGzFileName string) (err error) {
	// 创建 tar.gz 文件
	tarFile, err := os.Create(toGzFileName)
	if err != nil {
		return err
	}
	defer func() {
		err1 := tarFile.Close()
		if err == nil {
			err = err1
		}
	}()

	gw := gzip.NewWriter(tarFile)
	defer func() {
		err1 := gw.Close()
		if err == nil {
			err = err1
		}
	}()

	tw := tar.NewWriter(gw)
	defer func() {
		err1 := tw.Close()
		if err == nil {
			err = err1
		}
	}()

	// 遍历目录并添加文件到 tar.gz
	return filepath.Walk(srcFilePath, func(file string, fi os.FileInfo, walkErr error) (err1 error) {
		if walkErr != nil { // fi 此时可能为 nil
			return walkErr
		}
		header, err1 := tar.FileInfoHeader(fi, file)
		if err1 != nil {
			return err1
		}

		header.Name = filepath.ToSlash(file)

		if err1 = tw.WriteHeader(header); err1 != nil {
			return err1
		}
		if !fi.Mode().IsRegular() { // 目录等只需写入 header，没有内容可拷贝
			return nil
		}

		data, err1 := os.Open(file)
		if err1 != nil {
			return err1
		}
		defer func() {
			err2 := data.Close()
			if err1 == nil {
				err1 = err2
			}
		}()
		_, err1 = io.Copy(tw, data)
		return err1
	})
}

// UnZipFile 解压zip压缩文件
func UnZipFile(zipFile string, dest string) (err error) {
	var r *zip.ReadCloser
	r, err = zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer func(r *zip.ReadCloser) {
		// 只在没有错误时记录 Close 的错误，原来会把解压过程中返回的错误覆盖为 nil
		if closeErr := r.Close(); err == nil {
			err = closeErr
		}
	}(r)

	for _, f := range r.File {
		var rc io.ReadCloser
		rc, err = f.Open()
		if err != nil {
			return err
		}
		err = func(rc io.ReadCloser) error {
			defer func(rc io.ReadCloser) {
				err = rc.Close()
			}(rc)

			filePath := filepath.Join(dest, f.Name)
			// 防止 Zip Slip 路径穿越：确保解压目标始终落在 dest 目录内
			if rel, errRel := filepath.Rel(dest, filePath); errRel != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				return fmt.Errorf("illegal file path in zip: %q", f.Name)
			}
			if f.FileInfo().IsDir() {
				err = os.MkdirAll(filePath, os.ModePerm)
			} else {
				var dir string
				if lastIndex := strings.LastIndex(filePath, string(os.PathSeparator)); lastIndex > -1 {
					dir = filePath[:lastIndex]
				}
				err = os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					return err
				}

				f, err1 := os.OpenFile(
					filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err1 != nil {
					return err1
				}

				defer func(f *os.File) {
					err = f.Close()
				}(f)

				_, err = io.Copy(f, rc)
				if err != nil {
					return err
				}
			}
			return err
		}(rc)

		if err != nil {
			return err
		}
	}
	return nil
}
