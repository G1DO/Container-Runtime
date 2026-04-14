// Package tests contains integration tests that verify container isolation.
// These tests run actual containers and verify that namespaces correctly
// prevent the container from seeing or affecting the host.
//
// IMPORTANT: These tests require root privileges and a Linux kernel with
// namespace and cgroup v2 support. Run with: sudo go test -v ./tests/ -tags=integration
//
// Milestones: M9.1 (namespace isolation tests)
package tests

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func needsRoot(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("requires linux")
	}
	out, err := exec.Command("id", "-u").Output()
	if err != nil || strings.TrimSpace(string(out)) != "0" {
		t.Skip("requires root")
	}
}

func runtimeBinary(t *testing.T) string {
	t.Helper()
	// Build the binary into a temp location.
	bin := t.TempDir() + "/myruntime"
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/myruntime/")
	cmd.Dir = ".."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

// TestPIDNamespacePID1 verifies the container process sees itself as PID 1.
func TestPIDNamespacePID1(t *testing.T) {
	needsRoot(t)
	bin := runtimeBinary(t)

	out, err := exec.Command(bin, "run", "/bin/sh", "-c", "echo $$").CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	pid := strings.TrimSpace(string(out))
	if pid != "1" {
		t.Errorf("expected PID 1 inside namespace, got %q", pid)
	}
}

// TestPIDIsolation verifies that a container only sees its own processes.
func TestPIDIsolation(t *testing.T) {
	needsRoot(t)
	
	// NOTE: ps reads /proc — since we don't have mount namespace isolation yet
	// (M1.2), /proc is the host's, so ps will show host processes.
	// This test will become meaningful after M1.2 mounts a fresh /proc.
	t.Skip("requires mount namespace (M1.2) to mount fresh /proc")
}

// TestHostnameIsolation verifies that setting the hostname inside the container
// does not affect the host's hostname.
func TestHostnameIsolation(t *testing.T) {
	// TODO(M9.1): Record host hostname
	// TODO(M9.1): Start container that sets hostname to "test-container"
	// TODO(M9.1): Verify host hostname is unchanged
	t.Skip("not yet implemented")
}

// TestFilesystemIsolation verifies that writes inside the container
// do not appear on the host filesystem.
func TestFilesystemIsolation(t *testing.T) {
	// TODO(M9.1): Start container, write file to /tmp/testfile
	// TODO(M9.1): Verify /tmp/testfile does NOT exist on host
	t.Skip("not yet implemented")
}

// TestUserNamespaceIsolation verifies that root inside the container
// maps to an unprivileged UID on the host.
func TestUserNamespaceIsolation(t *testing.T) {
	// TODO(M9.1): Start container, run "id" — should show uid=0(root)
	// TODO(M9.1): From host, check /proc/<pid>/status — should show unprivileged UID
	t.Skip("not yet implemented")
}

// TestIPCIsolation verifies that IPC resources created in the container
// are not visible on the host.
func TestIPCIsolation(t *testing.T) {
	// TODO(M9.1): Create shared memory segment inside container
	// TODO(M9.1): Verify it's not visible via ipcs on host
	t.Skip("not yet implemented")
}
