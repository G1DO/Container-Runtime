// Package cgroup manages cgroup v2 resource limits for containers.
// Cgroups cap how much CPU, memory, IO, and processes a container can use.
// Without cgroups, a container could eat all host resources despite namespace isolation.
//
// All operations work by reading/writing files under /sys/fs/cgroup/.
//
// Milestones: M2.1 (foundation), M2.2 (CPU), M2.3 (memory), M2.4 (PIDs), M2.5 (IO)
package cgroup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/G1DO/Container-Runtime/pkg/specs"
)

var supportedControllers = []string{"cpu", "io", "memory", "pids"}

// Manager creates, configures, and destroys cgroup v2 hierarchies for containers.
type Manager struct {
	// BasePath is the runtime's cgroup root (e.g., /sys/fs/cgroup/myruntime).
	BasePath string
}

// NewManager creates a cgroup manager rooted at the given path.
func NewManager(basePath string) *Manager {
	_ = enableAvailableControllers(filepath.Dir(basePath))
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil
	}

	manager := &Manager{BasePath: basePath}
	_ = enableAvailableControllers(manager.BasePath)
	return manager
}

// Create sets up a per-container cgroup with the specified resource limits.
// Creates /sys/fs/cgroup/myruntime/<containerID>/ and writes limit files.
func (m *Manager) Create(containerID string, config specs.ResourceConfig) error {
	cgroupPath, err := m.containerPath(containerID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cgroupPath, 0o755); err != nil {
		return fmt.Errorf("create cgroup %q: %w", containerID, err)
	}

	if config.CPUQuota > 0 {
		if err := m.setCPULimit(cgroupPath, config.CPUQuota, config.CPUPeriod); err != nil {
			return err
		}
	}

	// TODO(M2.3): Write memory.max
	// TODO(M2.4): Write pids.max
	// TODO(M2.5): Write io.max
	return nil
}

// AddProcess moves a process into a container's cgroup by writing its PID
// to cgroup.procs. Must be called from the parent before the child execs.
func (m *Manager) AddProcess(containerID string, pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid %d", pid)
	}

	cgroupPath, err := m.containerPath(containerID)
	if err != nil {
		return err
	}

	procsPath := filepath.Join(cgroupPath, "cgroup.procs")
	if err := os.WriteFile(procsPath, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		return fmt.Errorf("add pid %d to cgroup %q: %w", pid, containerID, err)
	}
	return nil
}

// Destroy removes a container's cgroup. All processes must be dead first —
// the kernel refuses to remove a cgroup with live processes.
func (m *Manager) Destroy(containerID string) error {
	cgroupPath, err := m.containerPath(containerID)
	if err != nil {
		return err
	}

	procsPath := filepath.Join(cgroupPath, "cgroup.procs")
	procs, err := os.ReadFile(procsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read cgroup.procs for %q: %w", containerID, err)
	}
	if strings.TrimSpace(string(procs)) != "" {
		return fmt.Errorf("destroy cgroup %q: cgroup still has live processes", containerID)
	}

	if err := os.Remove(cgroupPath); err != nil {
		if isDirectoryNotEmpty(err) {
			if err := os.RemoveAll(cgroupPath); err != nil {
				return fmt.Errorf("remove cgroup %q: %w", containerID, err)
			}
			return nil
		}
		return fmt.Errorf("remove cgroup %q: %w", containerID, err)
	}
	return nil
}

// Stats reads live resource usage from cgroup stat files.
func (m *Manager) Stats(containerID string) (*specs.CgroupStats, error) {
	// TODO(M2.6): Read cpu.stat, memory.current, memory.max, pids.current, io.stat
	return nil, nil
}

func enableAvailableControllers(cgroupPath string) error {
	controllersPath := filepath.Join(cgroupPath, "cgroup.controllers")
	controllers, err := os.ReadFile(controllersPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read cgroup.controllers: %w", err)
	}

	available := strings.Fields(string(controllers))
	availableSet := make(map[string]struct{}, len(available))
	for _, controller := range available {
		availableSet[controller] = struct{}{}
	}

	enabled := make([]string, 0, len(supportedControllers))
	for _, controller := range supportedControllers {
		if _, ok := availableSet[controller]; ok {
			enabled = append(enabled, "+"+controller)
		}
	}
	if len(enabled) == 0 {
		return nil
	}

	subtreeControlPath := filepath.Join(cgroupPath, "cgroup.subtree_control")
	if err := os.WriteFile(subtreeControlPath, []byte(strings.Join(enabled, " ")+"\n"), 0o644); err != nil {
		return fmt.Errorf("enable cgroup controllers: %w", err)
	}
	return nil
}

// setCPULimit writes cpu.max and cpu.period_us to apply CPU throttling limits.
// quota: microseconds per period (e.g., 50000 = 50% of one core)
// period: microseconds (default 100000 = 100ms if 0)
func (m *Manager) setCPULimit(cgroupPath string, quota int64, period int64) error {
	if period == 0 {
		period = 100000
	}

	maxPath := filepath.Join(cgroupPath, "cpu.max")
	content := fmt.Sprintf("%d %d\n", quota, period)
	if err := os.WriteFile(maxPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("set cpu.max: %w", err)
	}

	return nil
}

func (m *Manager) containerPath(containerID string) (string, error) {
	if m == nil {
		return "", fmt.Errorf("nil cgroup manager")
	}
	if err := validateContainerID(containerID); err != nil {
		return "", err
	}
	return filepath.Join(m.BasePath, containerID), nil
}

func validateContainerID(containerID string) error {
	if containerID == "" {
		return fmt.Errorf("container id is empty")
	}
	if filepath.IsAbs(containerID) || filepath.Clean(containerID) != containerID || strings.ContainsRune(containerID, os.PathSeparator) {
		return fmt.Errorf("unsafe container id %q", containerID)
	}
	if containerID == "." || containerID == ".." {
		return fmt.Errorf("unsafe container id %q", containerID)
	}
	return nil
}

func isDirectoryNotEmpty(err error) bool {
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		return false
	}
	return errors.Is(pathErr.Err, syscall.ENOTEMPTY)
}
