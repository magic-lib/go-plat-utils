package compress_test

import (
	"fmt"
	"testing"

	"github.com/magic-lib/go-plat-utils/compress"
)

func TestZZBrProbe(t *testing.T) {
	b, err := compress.BrCompress([]byte("tianlin0"))
	fmt.Printf("BYTES %v %v\n", b, err)
}
