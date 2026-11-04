package compress_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/magic-lib/go-plat-utils/compress"
)

// 压缩结果必须是完整的 brotli 流，严格解码不能报 unexpected EOF
func TestBrCompressCompleteStream(t *testing.T) {
	data, err := compress.BrCompress([]byte("tianlin0"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(brotli.NewReader(bytes.NewReader(data)))
	if err != nil || string(out) != "tianlin0" {
		t.Fatalf("严格解码 = %q, %v", out, err)
	}
}
