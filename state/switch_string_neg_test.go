package state

import "testing"

func TestSetAllStatesNegative(t *testing.T) {
	ss := NewStringSwitch(4)
	_ = ss.TurnOn(1)
	if err := ss.SetAllStates(map[int]bool{-1: true, 2: true}); err == nil {
		t.Fatal("negative position should return error")
	}
	if on, _ := ss.IsOn(1); !on {
		t.Fatal("state should be unchanged on error")
	}
}
