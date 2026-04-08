// Package cgroup manages cgroup v2 resource limits for containers.
// Cgroups cap how much CPU, memory, IO, and processes a container can use.
// Without cgroups, a container could eat all host resources despite namespace isolation.
//
// All operations work by reading/writing files under /sys/fs/cgroup/.
//
// Milestones: M2.1 (foundation), M2.2 (CPU), M2.3 (memory), M2.4 (PIDs), M2.5 (IO)
package cgroup

import (
	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// Manager creates, configures, and destroys cgroup v2 hierarchies for containers.
type Manager struct {
	// BasePath is the runtime's cgroup root (e.g., /sys/fs/cgroup/myruntime).
	BasePath string
}

// NewManager creates a cgroup manager rooted at the given path.
func NewManager(basePath string) *Manager {
	// TODO(M2.1): Create basePath directory, enable subtree controllers
	return nil
}

// Create sets up a per-container cgroup with the specified resource limits.
// Creates /sys/fs/cgroup/myruntime/<containerID>/ and writes limit files.
func (m *Manager) Create(containerID string, config specs.ResourceConfig) error {
	// TODO(M2.1): mkdir cgroup dir
	// TODO(M2.1): Enable controllers in parent subtree_control
	// TODO(M2.2): Write cpu.max
	// TODO(M2.3): Write memory.max
	// TODO(M2.4): Write pids.max
	// TODO(M2.5): Write io.max
	return nil
}

// AddProcess moves a process into a container's cgroup by writing its PID
// to cgroup.procs. Must be called from the parent before the child execs.
func (m *Manager) AddProcess(containerID string, pid int) error {
	// TODO(M2.1): Write pid to <cgroupPath>/cgroup.procs
	return nil
}

// Destroy removes a container's cgroup. All processes must be dead first —
// the kernel refuses to remove a cgroup with live processes.
func (m *Manager) Destroy(containerID string) error {
	// TODO(M2.1): Kill remaining procs, wait, rmdir cgroup path
	return nil
}

// Stats reads live resource usage from cgroup stat files.
func (m *Manager) Stats(containerID string) (*specs.CgroupStats, error) {
	// TODO(M2.6): Read cpu.stat, memory.current, memory.max, pids.current, io.stat
	return nil, nil
}
