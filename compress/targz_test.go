package compress_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/magic-lib/go-plat-utils/compress"
)

// 压缩目录后返回的内容必须是完整可解压的 tar.gz，且包含所有文件
func TestCreateGzFileDir(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(src, "a.txt"):        "hello",
		filepath.Join(src, "sub", "b.txt"): "world",
	}
	for p, c := range files {
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	data, err := compress.CreateGzFile(src, filepath.Join(root, "out.tar.gz"))
	if err != nil {
		t.Fatalf("CreateGzFile err: %v", err)
	}
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	tr := tar.NewReader(gr)
	got := map[string]string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("读取 tar 失败（内容不完整）: %v", err)
		}
		if hdr.Typeflag == tar.TypeReg {
			b, _ := io.ReadAll(tr)
			got[filepath.FromSlash(hdr.Name)] = string(b)
		}
	}
	if _, err = io.ReadAll(gr); err != nil {
		t.Fatalf("gzip 尾部校验失败: %v", err)
	}
	for p, c := range files {
		if got[p] != c {
			t.Errorf("%s = %q, want %q", p, got[p], c)
		}
	}
}

// 压缩单个文件，返回内容需与落盘文件一致且包含 gzip 尾部
func TestCreateGzFileSingle(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "a.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "a.tar.gz")
	data, err := compress.CreateGzFile(src, out)
	if err != nil {
		t.Fatalf("CreateGzFile err: %v", err)
	}
	onDisk, _ := os.ReadFile(out)
	if !bytes.Equal(data, onDisk) {
		t.Fatalf("返回内容 %d 字节与落盘文件 %d 字节不一致", len(data), len(onDisk))
	}
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	if _, err = io.ReadAll(gr); err != nil {
		t.Fatalf("返回内容不完整: %v", err)
	}
}
