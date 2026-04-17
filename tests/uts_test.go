// Package tests — uts_test.go verifies UTS namespace hostname isolation.
// The container should be able to set its own hostname without affecting the host.
//
// IMPORTANT: Requires root privileges and a Linux kernel with UTS namespace support.
// Run with: sudo go test -v ./tests/ -tags=integration
//
// Milestones: M1.3 (UTS namespace — hostname isolation)
package tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestContainerHostname verifies that the container can set its own hostname
// and that it takes effect inside the namespace.
func TestContainerHostname(t *testing.T) {
	needsRoot(t)
	bin := runtimeBinary(t)
	rootfs := testRootFS(t)

	out, err := exec.Command(
		bin,
		"run",
		"--rootfs",
		rootfs,
		"--hostname",
		"test-container",
		"/bin/hostname",
	).CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}

	if got := strings.TrimSpace(string(out)); got != "test-container" {
		t.Fatalf("expected container hostname %q, got %q", "test-container", got)
	}
}

// TestHostHostnameUnchanged verifies that setting a hostname inside the container
// does not modify the host's hostname.
func TestHostHostnameUnchanged(t *testing.T) {
	needsRoot(t)
	bin := runtimeBinary(t)
	rootfs := testRootFS(t)

	before, err := os.Hostname()
	if err != nil {
		t.Fatalf("hostname before: %v", err)
	}

	out, err := exec.Command(
		bin,
		"run",
		"--rootfs",
		rootfs,
		"--hostname",
		"container-test",
		"/bin/true",
	).CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}

	after, err := os.Hostname()
	if err != nil {
		t.Fatalf("hostname after: %v", err)
	}

	if before != after {
		t.Fatalf("host hostname changed: before=%q after=%q", before, after)
	}
}
