package activity_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/plugins/activity"
)

func TestActivityClone(t *testing.T) {
	ac := &activity.Activity{Id: "self", ActName: "selfAct"}
	other := &activity.Activity{Id: "other", ActName: "otherAct"}

	if got := ac.Clone(); got.Id != "self" || got.ActName != "selfAct" || got == ac {
		t.Fatalf("Clone() = %+v, want 复制自身", got)
	}
	if got := ac.Clone(other); got.Id != "other" || got.ActName != "otherAct" {
		t.Fatalf("Clone(other) = %+v, want 复制 other", got)
	}
	if got := ac.Clone(nil); got.Id != "self" {
		t.Fatalf("Clone(nil) = %+v, want 复制自身", got)
	}
}
