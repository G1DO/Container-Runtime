// Package tests — mount_test.go verifies mount namespace and pivot_root isolation.
// These tests confirm that the container gets its own filesystem root and that
// writes inside the container do not leak to the host.
//
// IMPORTANT: Requires root privileges and a Linux kernel with mount namespace support.
// Run with: sudo go test -v ./tests/ -tags=integration
//
// Milestones: M1.2 (mount namespace + pivot_root)
package tests

import "testing"

// TestPivotRootIsolation verifies that the container's root filesystem is
// completely separate from the host. After pivot_root, the old root should
// be unmounted and inaccessible.
func TestPivotRootIsolation(t *testing.T) {
	// TODO(M1.2): Start container with Alpine rootfs
	// TODO(M1.2): Run "ls /" inside container — should show only container's rootfs contents
	// TODO(M1.2): Verify host "/" is NOT visible from inside the container
	t.Skip("not yet implemented")
}

// TestFilesystemWriteIsolation verifies that files created inside the container
// do not appear on the host filesystem.
func TestFilesystemWriteIsolation(t *testing.T) {
	// TODO(M1.2): Start container, run "touch /tmp/test-m12" inside
	// TODO(M1.2): Verify /tmp/test-m12 does NOT exist on host
	t.Skip("not yet implemented")
}

// TestHostMountsUnaffected verifies that the host's mount table is unchanged
// after a container exits. Mount namespace prevents mount propagation.
func TestHostMountsUnaffected(t *testing.T) {
	// TODO(M1.2): Record host mounts via /proc/self/mountinfo before container start
	// TODO(M1.2): Start and stop a container
	// TODO(M1.2): Compare host mounts after — should be identical
	t.Skip("not yet implemented")
}

// TestProcMountShowsContainerProcesses verifies that /proc inside the container
// reflects only the container's PID namespace, not host processes.
// Requires both PID namespace (M1.1) and mount namespace (M1.2).
func TestProcMountShowsContainerProcesses(t *testing.T) {
	// TODO(M1.2): Start container, run "ls /proc" or "ps aux"
	// TODO(M1.2): Verify only container processes are visible (PID 1 + ps)
	// TODO(M1.2): Verify host PIDs are NOT listed
	t.Skip("not yet implemented")
}
