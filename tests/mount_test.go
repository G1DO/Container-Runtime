// Package tests — mount_test.go verifies mount namespace and pivot_root isolation.
// These tests confirm that the container gets its own filesystem root and that
// writes inside the container do not leak to the host.
//
// IMPORTANT: Requires root privileges and a Linux kernel with mount namespace support.
// Run with: sudo go test -v ./tests/ -tags=integration
//
// Milestones: M1.2 (mount namespace + pivot_root)
package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testRootFS(t *testing.T) string {
	t.Helper()
	p := filepath.Clean("../testdata/rootfs")
	if _, err := os.Stat(p); err != nil {
		t.Skip("test rootfs missing; run ./scripts/setup-test-env.sh")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs rootfs: %v", err)
	}
	return abs
}

// TestPivotRootIsolation verifies that the container's root filesystem is
// completely separate from the host. After pivot_root, the old root should
// be unmounted and inaccessible.
func TestPivotRootIsolation(t *testing.T) {
	needsRoot(t)
	bin := runtimeBinary(t)
	rootfs := testRootFS(t)

	out, err := exec.Command(bin, "run", "--rootfs", rootfs, "/bin/sh", "-c", "cat /etc/os-release").CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Alpine") {
		t.Fatalf("expected Alpine rootfs after pivot_root, got:\n%s", out)
	}
}

// TestFilesystemWriteIsolation verifies that files created inside the container
// do not appear on the host filesystem.
func TestFilesystemWriteIsolation(t *testing.T) {
	needsRoot(t)
	bin := runtimeBinary(t)
	rootfs := testRootFS(t)

	name := "test-m12-" + strings.ReplaceAll(t.Name(), "/", "-")
	hostPath := filepath.Join("/tmp", name)
	_ = os.Remove(hostPath)
	t.Cleanup(func() { _ = os.Remove(hostPath) })

	out, err := exec.Command(bin, "run", "--rootfs", rootfs, "/bin/sh", "-c", "touch /tmp/"+name).CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(hostPath); err == nil {
		t.Fatalf("file leaked to host: %s", hostPath)
	}

	// Best-effort cleanup: our rootfs is a real directory on the host.
	_ = os.Remove(filepath.Join(rootfs, "tmp", name))
}

// TestHostMountsUnaffected verifies that the host's mount table is unchanged
// after a container exits. Mount namespace prevents mount propagation.
func TestHostMountsUnaffected(t *testing.T) {
	needsRoot(t)
	bin := runtimeBinary(t)
	rootfs := testRootFS(t)

	before, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		t.Fatalf("read mountinfo before: %v", err)
	}

	out, err := exec.Command(bin, "run", "--rootfs", rootfs, "/bin/true").CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}

	after, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		t.Fatalf("read mountinfo after: %v", err)
	}

	if string(before) != string(after) {
		t.Fatalf("host mount table changed (mount propagation leak)")
	}
}

// TestProcMountShowsContainerProcesses verifies that /proc inside the container
// reflects only the container's PID namespace, not host processes.
// Requires both PID namespace (M1.1) and mount namespace (M1.2).
func TestProcMountShowsContainerProcesses(t *testing.T) {
	needsRoot(t)
	bin := runtimeBinary(t)
	rootfs := testRootFS(t)

	out, err := exec.Command(bin, "run", "--rootfs", rootfs, "/bin/sh", "-c", "cat /proc/1/comm").CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "sh" {
		t.Fatalf("expected /proc to reflect PID namespace (PID 1 comm = sh), got %q", got)
	}
}
