package compress_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/magic-lib/go-plat-utils/compress"
)

// 解压出错（路径穿越）时必须返回错误，不能被 Close 的结果覆盖为 nil
func TestUnZipFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "evil.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create("../evil.txt")
	_, _ = w.Write([]byte("x"))
	_ = zw.Close()
	_ = f.Close()

	if err := compress.UnZipFile(zipPath, filepath.Join(dir, "out")); err == nil {
		t.Fatal("路径穿越应返回错误")
	}
}
