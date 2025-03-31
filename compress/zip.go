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

	// 创建 tar.gz 文件
	tarFile, err := os.Create(toGzFileName)
	if err != nil {
		return nil, err
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
	err = filepath.Walk(srcFilePath, func(file string, fi os.FileInfo, err error) (err1 error) {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(file)

		if err = tw.WriteHeader(header); err != nil {
			return err
		}

		data, err := os.Open(file)
		if err != nil {
			return err
		}
		defer func() {
			err2 := data.Close()
			if err1 == nil {
				err = err2
			}
		}()
		_, err1 = io.Copy(tw, data)
		return err1
	})

	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(toGzFileName)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// UnZipFile 解压zip压缩文件
func UnZipFile(zipFile string, dest string) (err error) {
	var r *zip.ReadCloser
	r, err = zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer func(r *zip.ReadCloser) {
		err = r.Close()
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
