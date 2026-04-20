package namespace

import (
	"syscall"
	"testing"
)

func TestCloneFlagsPhase1(t *testing.T) {
	got := CloneFlags(nil)
	want := uintptr(syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC)
	if got != want {
		t.Fatalf("CloneFlags(nil) = %#x, want %#x", got, want)
	}
}
