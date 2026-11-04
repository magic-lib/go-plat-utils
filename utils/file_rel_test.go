package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/magic-lib/go-plat-utils/utils"
)

// dest 以 / 结尾时仍返回相对路径
func TestGetAllFileContentTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, dest := range []string{dir, dir + "/"} {
		files, err := utils.GetAllFileContent(dest)
		if err != nil || files["a.txt"] != "x" {
			t.Fatalf("GetAllFileContent(%q) = %v, %v", dest, files, err)
		}
	}
}
