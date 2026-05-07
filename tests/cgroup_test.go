// Package tests — cgroup_test.go verifies that resource limits are actually enforced.
// These tests deliberately try to exceed limits and verify the kernel stops them.
//
// Milestones: M2.2 (CPU limits), M9.2 (cgroup enforcement tests)
package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestMemoryLimit starts a container with a 64MB memory limit, then tries
// to allocate 128MB. The kernel should OOM-kill the process (exit code 137).
func TestMemoryLimit(t *testing.T) {
	// TODO(M9.2): Start container with 64MB memory limit
	// TODO(M9.2): Run "stress --vm 1 --vm-bytes 128M" inside
	// TODO(M9.2): Assert exit code == 137 (SIGKILL from OOM)
	t.Skip("not yet implemented")
}

// TestCPULimit starts a container with 50% CPU, runs a CPU-bound task,
// and verifies throttling happens by checking cpu.stat counters.
func TestCPULimit(t *testing.T) {
	needsRoot(t)

	// Create a test cgroup with 50% CPU limit (50000 µs per 100000 µs period)
	basePath := filepath.Join(t.TempDir(), "myruntime")
	containerID := "test-cpu-limit"
	cgroupPath := filepath.Join(basePath, containerID)

	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		t.Fatalf("mkdir cgroup: %v", err)
	}

	// Write cpu.max: 50% of one core
	cpuMaxContent := "50000 100000\n"
	if err := os.WriteFile(filepath.Join(cgroupPath, "cpu.max"), []byte(cpuMaxContent), 0o644); err != nil {
		t.Fatalf("write cpu.max: %v", err)
	}

	// Create a CPU-bound process and put it in the cgroup
	cmd := exec.Command("bash", "-c", "while true; do :; done")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start CPU process: %v", err)
	}
	defer cmd.Process.Kill()

	// Add process to cgroup
	procsPath := filepath.Join(cgroupPath, "cgroup.procs")
	if err := os.WriteFile(procsPath, []byte(strconv.Itoa(cmd.Process.Pid)+"\n"), 0o644); err != nil {
		t.Fatalf("add process to cgroup: %v", err)
	}

	// Let it run for 2 seconds
	time.Sleep(2 * time.Second)

	// Read cpu.stat and verify throttling occurred
	cpuStatPath := filepath.Join(cgroupPath, "cpu.stat")
	statBytes, err := os.ReadFile(cpuStatPath)
	if err != nil {
		t.Fatalf("read cpu.stat: %v", err)
	}

	statContent := string(statBytes)

	// Parse nr_throttled from cpu.stat
	// Format: "nr_periods 123\nnr_throttled 45\nthrottled_usec 789\n..."
	var nrThrottled int64
	for _, line := range strings.Split(statContent, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nr_throttled" {
			if n, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
				nrThrottled = n
			}
		}
	}

	if nrThrottled == 0 {
		t.Fatalf("cpu.stat nr_throttled = 0, want > 0 (50%% CPU should be throttled)")
	}

	t.Logf("CPU throttled %d times as expected", nrThrottled)
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
