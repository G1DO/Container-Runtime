package cgroup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/G1DO/Container-Runtime/pkg/specs"
)

func TestNewManagerCreatesBasePath(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "myruntime")

	manager := NewManager(basePath)
	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
	if manager.BasePath != basePath {
		t.Fatalf("NewManager().BasePath = %q, want %q", manager.BasePath, basePath)
	}
	assertDirExists(t, basePath)
}

func TestNewManagerEnablesAvailableControllers(t *testing.T) {
	basePath := t.TempDir()
	writeFile(t, filepath.Join(basePath, "cgroup.controllers"), "cpu io memory pids hugetlb\n")
	writeFile(t, filepath.Join(basePath, "cgroup.subtree_control"), "")

	manager := NewManager(basePath)
	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	got := readFile(t, filepath.Join(basePath, "cgroup.subtree_control"))
	for _, controller := range []string{"+cpu", "+io", "+memory", "+pids"} {
		if !strings.Contains(got, controller) {
			t.Fatalf("cgroup.subtree_control = %q, want controller %q enabled", got, controller)
		}
	}
	if strings.Contains(got, "+hugetlb") {
		t.Fatalf("cgroup.subtree_control = %q, should only enable controllers used by the runtime", got)
	}
}

func TestCreateMakesContainerCgroupDirectory(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}

	if err := manager.Create("abc123", specs.ResourceConfig{}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	assertDirExists(t, filepath.Join(manager.BasePath, "abc123"))
}

func TestCreateRejectsUnsafeContainerID(t *testing.T) {
	basePath := t.TempDir()
	manager := &Manager{BasePath: filepath.Join(basePath, "myruntime")}

	err := manager.Create("../escape", specs.ResourceConfig{})
	if err == nil {
		t.Fatal("Create() with path traversal ID succeeded, want error")
	}
	if _, statErr := os.Stat(filepath.Join(basePath, "escape")); !os.IsNotExist(statErr) {
		t.Fatalf("Create() wrote outside BasePath, stat error = %v", statErr)
	}
}

func TestAddProcessWritesPIDToCgroupProcs(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	containerID := "abc123"
	cgroupPath := filepath.Join(manager.BasePath, containerID)
	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		t.Fatalf("mkdir cgroup path: %v", err)
	}

	if err := manager.AddProcess(containerID, 4242); err != nil {
		t.Fatalf("AddProcess() error = %v", err)
	}

	assertFileContent(t, filepath.Join(cgroupPath, "cgroup.procs"), "4242\n")
}

func TestAddProcessRejectsInvalidPID(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}

	err := manager.AddProcess("abc123", 0)
	if err == nil {
		t.Fatal("AddProcess() with pid 0 succeeded, want error")
	}
}

func TestDestroyRemovesEmptyContainerCgroup(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	containerID := "abc123"
	cgroupPath := filepath.Join(manager.BasePath, containerID)
	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		t.Fatalf("mkdir cgroup path: %v", err)
	}
	writeFile(t, filepath.Join(cgroupPath, "cgroup.procs"), "")

	if err := manager.Destroy(containerID); err != nil {
		t.Fatalf("Destroy() error = %v", err)
	}

	if _, err := os.Stat(cgroupPath); !os.IsNotExist(err) {
		t.Fatalf("cgroup path still exists after Destroy(), stat error = %v", err)
	}
}

func TestDestroyRejectsCgroupWithLiveProcesses(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	containerID := "abc123"
	cgroupPath := filepath.Join(manager.BasePath, containerID)
	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		t.Fatalf("mkdir cgroup path: %v", err)
	}
	writeFile(t, filepath.Join(cgroupPath, "cgroup.procs"), "4242\n")

	err := manager.Destroy(containerID)
	if err == nil {
		t.Fatal("Destroy() with live processes succeeded, want error")
	}
	assertDirExists(t, cgroupPath)
}

func TestCreateWritesCPULimitWithQuota(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	config := specs.ResourceConfig{
		CPUQuota:  50000,
		CPUPeriod: 100000,
	}

	if err := manager.Create("abc123", config); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	cpuMaxPath := filepath.Join(manager.BasePath, "abc123", "cpu.max")
	assertFileContent(t, cpuMaxPath, "50000 100000\n")
}

func TestCreateWithZeroCPUQuotaSkipsLimit(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	config := specs.ResourceConfig{
		CPUQuota:  0,
		CPUPeriod: 100000,
	}

	if err := manager.Create("abc123", config); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	cpuMaxPath := filepath.Join(manager.BasePath, "abc123", "cpu.max")
	if _, err := os.Stat(cpuMaxPath); !os.IsNotExist(err) {
		t.Fatalf("cpu.max should not be written when CPUQuota is 0")
	}
}

func TestCreateDefaultsCPUPeriodToOneHundredMs(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	config := specs.ResourceConfig{
		CPUQuota:  50000,
		CPUPeriod: 0,
	}

	if err := manager.Create("abc123", config); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	cpuMaxPath := filepath.Join(manager.BasePath, "abc123", "cpu.max")
	assertFileContent(t, cpuMaxPath, "50000 100000\n")
}

func TestCreateWritesMemoryLimitWithMemoryMax(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	config := specs.ResourceConfig{
		MemoryMax: 64 << 20,
	}

	if err := manager.Create("abc123", config); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	memMaxPath := filepath.Join(manager.BasePath, "abc123", "memory.max")
	assertFileContent(t, memMaxPath, "67108864\n")
}

func TestCreateWithZeroMemoryMaxSkipsLimit(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	config := specs.ResourceConfig{
		MemoryMax: 0,
	}

	if err := manager.Create("abc123", config); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	memMaxPath := filepath.Join(manager.BasePath, "abc123", "memory.max")
	if _, err := os.Stat(memMaxPath); !os.IsNotExist(err) {
		t.Fatalf("memory.max should not be written when MemoryMax is 0")
	}
}

func TestStatsReadsMemoryCurrentAndLimit(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	containerID := "abc123"
	cgroupPath := filepath.Join(manager.BasePath, containerID)
	seedMemoryFiles(t, cgroupPath,
		"67108864\n",
		"4096\n",
		"low 0\nhigh 0\nmax 0\noom 0\noom_kill 0\n",
	)

	stats, err := manager.Stats(containerID)
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.MemoryCurrent != 4096 {
		t.Fatalf("Stats().MemoryCurrent = %d, want 4096", stats.MemoryCurrent)
	}
	if stats.MemoryLimit != 67108864 {
		t.Fatalf("Stats().MemoryLimit = %d, want 67108864", stats.MemoryLimit)
	}
}

func TestStatsReadsOOMKillCount(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	containerID := "abc123"
	cgroupPath := filepath.Join(manager.BasePath, containerID)
	seedMemoryFiles(t, cgroupPath,
		"67108864\n",
		"4096\n",
		"low 0\nhigh 0\nmax 3\noom 1\noom_kill 2\n",
	)

	stats, err := manager.Stats(containerID)
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.OOMKillCount != 2 {
		t.Fatalf("Stats().OOMKillCount = %d, want 2", stats.OOMKillCount)
	}
}

func TestStatsTreatsMemoryMaxAsUnlimited(t *testing.T) {
	manager := &Manager{BasePath: t.TempDir()}
	containerID := "abc123"
	cgroupPath := filepath.Join(manager.BasePath, containerID)
	seedMemoryFiles(t, cgroupPath,
		"max\n",
		"4096\n",
		"low 0\nhigh 0\nmax 0\noom 0\noom_kill 0\n",
	)

	stats, err := manager.Stats(containerID)
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.MemoryLimit != 0 {
		t.Fatalf("Stats().MemoryLimit = %d, want 0 (no limit)", stats.MemoryLimit)
	}
}

func seedMemoryFiles(t *testing.T, cgroupPath, max, current, events string) {
	t.Helper()
	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		t.Fatalf("mkdir cgroup path: %v", err)
	}
	writeFile(t, filepath.Join(cgroupPath, "memory.max"), max)
	writeFile(t, filepath.Join(cgroupPath, "memory.current"), current)
	writeFile(t, filepath.Join(cgroupPath, "memory.events"), events)
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s exists but is not a directory", path)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got := readFile(t, path)
	if got != want {
		t.Fatalf("%s = %q, want %q", path, got, want)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(got)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
