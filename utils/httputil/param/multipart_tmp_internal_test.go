package param

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

// 超过 32MB 的上传文件落盘后，解析完成即删除临时文件，文件内容仍可完整读取
func TestParseMultipartFormRemovesTempFiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "big.bin")
	payload := bytes.Repeat([]byte("a"), 33<<20)
	_, _ = fw.Write(payload)
	_ = mw.WriteField("name", "x")
	_ = mw.Close()

	h := http.Header{"Content-Type": {mw.FormDataContentType()}}
	data, err := parseMultipartForm(h, body.Bytes())
	if err != nil {
		t.Fatalf("parseMultipartForm: %v", err)
	}
	if entries, _ := os.ReadDir(tmpDir); len(entries) != 0 {
		t.Fatalf("临时文件未清理: %d 个", len(entries))
	}
	f := data.Files["file"][0]
	defer func() { _ = f.Close() }()
	got, err := io.ReadAll(f.File)
	if err != nil || len(got) != len(payload) {
		t.Fatalf("读取文件 = %d 字节, %v; want %d", len(got), err, len(payload))
	}
}
