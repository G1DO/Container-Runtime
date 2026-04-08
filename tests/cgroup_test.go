// Package tests — cgroup_test.go verifies that resource limits are actually enforced.
// These tests deliberately try to exceed limits and verify the kernel stops them.
//
// Milestones: M9.2 (cgroup enforcement tests)
package tests

import "testing"

// TestMemoryLimit starts a container with a 64MB memory limit, then tries
// to allocate 128MB. The kernel should OOM-kill the process (exit code 137).
func TestMemoryLimit(t *testing.T) {
	// TODO(M9.2): Start container with 64MB memory limit
	// TODO(M9.2): Run "stress --vm 1 --vm-bytes 128M" inside
	// TODO(M9.2): Assert exit code == 137 (SIGKILL from OOM)
	t.Skip("not yet implemented")
}

// TestCPULimit starts a container with 50% CPU, runs a CPU-bound task,
// and verifies usage stays near the limit.
func TestCPULimit(t *testing.T) {
	// TODO(M9.2): Start container with 0.5 CPU limit
	// TODO(M9.2): Run "stress --cpu 4 --timeout 10" inside
	// TODO(M9.2): Read cpu.stat, verify usage < 60% (with margin)
	t.Skip("not yet implemented")
}

// TestPIDLimit starts a container with pids.max=32, then runs a fork bomb.
// The fork bomb should be contained — new forks fail with EAGAIN.
func TestPIDLimit(t *testing.T) {
	// TODO(M9.2): Start container with pids.max=32
	// TODO(M9.2): Attempt fork bomb inside container
	// TODO(M9.2): Verify host PID count is unaffected
	t.Skip("not yet implemented")
}

// TestIOLimit verifies that disk IO is throttled by io.max settings.
func TestIOLimit(t *testing.T) {
	// TODO(M9.2): Start container with IO limit (e.g., 1MB/s write)
	// TODO(M9.2): Write data with dd, measure throughput
	// TODO(M9.2): Verify throughput near the configured limit
	t.Skip("not yet implemented")
}
