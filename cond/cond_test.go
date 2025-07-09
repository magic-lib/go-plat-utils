package cond_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"testing"
)

func TestIsUUID(t *testing.T) {
	isUUID := cond.IsUUID("e4ff48d4-ea6b-45b6-9217-35bc23e8a57f")
	fmt.Println(isUUID)
}
