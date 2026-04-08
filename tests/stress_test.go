// Package tests — stress_test.go pushes the runtime to its limits.
//
// Milestones: M9.3 (stress tests)
package tests

import "testing"

// TestManyContainers starts 50 containers simultaneously and verifies
// they all get unique PIDs, IPs, and cgroups, then cleans up properly.
func TestManyContainers(t *testing.T) {
	// TODO(M9.3): Start 50 containers in goroutines
	// TODO(M9.3): Verify unique IPs and cgroup paths
	// TODO(M9.3): Stop all, verify cleanup (no leftover cgroups/mounts)
	t.Skip("not yet implemented")
}

// TestForkBomb starts a container with a strict PID limit and runs a fork bomb.
// The fork bomb must be contained — the host should be completely unaffected.
func TestForkBomb(t *testing.T) {
	// TODO(M9.3): Start container with pids.max=32
	// TODO(M9.3): Run :(){ :|:& };: inside container
	// TODO(M9.3): Wait for container to stabilize or die
	// TODO(M9.3): Verify host PID count unchanged
	t.Skip("not yet implemented")
}

// TestDiskFull fills up a container's writable layer and verifies
// the host filesystem is unaffected.
func TestDiskFull(t *testing.T) {
	// TODO(M9.3): Start container
	// TODO(M9.3): Write data until writes fail inside container
	// TODO(M9.3): Verify host disk space unchanged (overlay upper has a limit or fills up)
	t.Skip("not yet implemented")
}

// TestRapidCreateDelete rapidly creates and deletes containers to check for
// resource leaks (file descriptors, mounts, cgroups).
func TestRapidCreateDelete(t *testing.T) {
	// TODO(M9.3): Create and delete 100 containers in rapid succession
	// TODO(M9.3): Verify no leftover mounts (cat /proc/mounts)
	// TODO(M9.3): Verify no leftover cgroups
	// TODO(M9.3): Verify no leaked file descriptors
	t.Skip("not yet implemented")
}
