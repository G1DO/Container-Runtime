// Package tests contains integration tests that verify container isolation.
// These tests run actual containers and verify that namespaces correctly
// prevent the container from seeing or affecting the host.
//
// IMPORTANT: These tests require root privileges and a Linux kernel with
// namespace and cgroup v2 support. Run with: sudo go test -v ./tests/ -tags=integration
//
// Milestones: M9.1 (namespace isolation tests)
package tests

import "testing"

// TestPIDIsolation verifies that a container only sees its own processes.
// Run "ps aux" inside the container — it should see at most 2 processes
// (PID 1 = the container init, and ps itself).
func TestPIDIsolation(t *testing.T) {
	// TODO(M9.1): Start container running "ps aux"
	// TODO(M9.1): Parse output, assert process count <= 2
	// TODO(M9.1): Verify host PIDs are NOT visible inside container
	t.Skip("not yet implemented")
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
