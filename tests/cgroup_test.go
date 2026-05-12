// Package tests — cgroup_test.go verifies that resource limits are actually enforced.
// These tests deliberately try to exceed limits and verify the kernel stops them.
//
// Milestones: M2.2 (CPU limits), M9.2 (cgroup enforcement tests)
package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestMemoryLimit starts a container with a 64MB memory limit, then tries
// to allocate well past it. The cgroup OOM killer should SIGKILL the process
// (exit code 137 = 128 + 9), and memory.events should record the kill.
func TestMemoryLimit(t *testing.T) {
	needsRoot(t)

	basePath := filepath.Join("/sys/fs/cgroup", "myruntime-test-"+strconv.Itoa(os.Getpid()))
	containerID := "test-memory-limit"
	cgroupPath := filepath.Join(basePath, containerID)

	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		t.Fatalf("mkdir cgroup: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(cgroupPath)
		_ = os.Remove(basePath)
	})

	// Enable the memory controller in the parent's subtree_control so the
	// child cgroup actually has memory.max available.
	if err := os.WriteFile(filepath.Join(basePath, "cgroup.subtree_control"), []byte("+memory\n"), 0o644); err != nil {
		t.Fatalf("enable memory controller: %v", err)
	}

	// 64 MiB hard limit. Disable swap so the kernel can't dodge the OOM.
	if err := os.WriteFile(filepath.Join(cgroupPath, "memory.max"), []byte("67108864\n"), 0o644); err != nil {
		t.Fatalf("write memory.max: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cgroupPath, "memory.swap.max"), []byte("0\n"), 0o644); err != nil && !os.IsNotExist(err) {
		t.Fatalf("write memory.swap.max: %v", err)
	}

	// tail /dev/zero reads zeros into an ever-growing buffer, forcing the
	// kernel to back each page with real RAM. Cheap, reliable OOM trigger.
	cmd := exec.Command("tail", "/dev/zero")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start victim: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	procsPath := filepath.Join(cgroupPath, "cgroup.procs")
	if err := os.WriteFile(procsPath, []byte(strconv.Itoa(cmd.Process.Pid)+"\n"), 0o644); err != nil {
		t.Fatalf("add process to cgroup: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("victim still alive after 15s; OOM killer never fired")
	}

	// exit code should be 137 (128 + SIGKILL). Go reports this via
	// ProcessState.ExitCode() once the process has been signalled.
	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 137 {
		t.Fatalf("exit code = %d, want 137 (SIGKILL from OOM)", exitCode)
	}

	// memory.events should now show oom_kill >= 1.
	events, err := os.ReadFile(filepath.Join(cgroupPath, "memory.events"))
	if err != nil {
		t.Fatalf("read memory.events: %v", err)
	}
	var oomKill int64 = -1
	for _, line := range strings.Split(string(events), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "oom_kill" {
			oomKill, _ = strconv.ParseInt(fields[1], 10, 64)
		}
	}
	if oomKill < 1 {
		t.Fatalf("memory.events oom_kill = %d, want >= 1; events=%q", oomKill, string(events))
	}

	t.Logf("victim OOM-killed as expected, oom_kill=%d", oomKill)
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
