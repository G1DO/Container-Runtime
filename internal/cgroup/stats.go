// Package cgroup — stats.go reads and parses cgroup v2 metric files.
// These files have simple text formats (key-value or single values)
// but parsing them correctly requires understanding what each metric means.
//
// Milestones: M2.6 (cgroup stats collection)
package cgroup

import (
	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// ReadCPUStat parses /sys/fs/cgroup/<path>/cpu.stat.
// Format:
//
//	usage_usec 123456
//	user_usec 100000
//	system_usec 23456
//	nr_periods 500
//	nr_throttled 12
//	throttled_usec 34567
func ReadCPUStat(cgroupPath string) (usageUsec uint64, err error) {
	// TODO(M2.6): Read and parse cpu.stat file
	return 0, nil
}

// ReadMemoryCurrent reads /sys/fs/cgroup/<path>/memory.current.
// Returns current memory usage in bytes.
func ReadMemoryCurrent(cgroupPath string) (int64, error) {
	// TODO(M2.6): Read single integer from memory.current
	return 0, nil
}

// ReadMemoryMax reads /sys/fs/cgroup/<path>/memory.max.
// Returns "max" as -1, or the byte limit.
func ReadMemoryMax(cgroupPath string) (int64, error) {
	// TODO(M2.6): Read memory.max, handle "max" string
	return 0, nil
}

// ReadPidsCurrent reads /sys/fs/cgroup/<path>/pids.current.
func ReadPidsCurrent(cgroupPath string) (int64, error) {
	// TODO(M2.6): Read single integer from pids.current
	return 0, nil
}

// ReadMemoryEvents parses /sys/fs/cgroup/<path>/memory.events for OOM counts.
// Format:
//
//	oom 3
//	oom_kill 2
func ReadMemoryEvents(cgroupPath string) (oomKillCount int64, err error) {
	// TODO(M2.6): Parse memory.events for oom_kill count
	return 0, nil
}

// CollectStats gathers all cgroup metrics into a single CgroupStats struct.
func CollectStats(cgroupPath string) (*specs.CgroupStats, error) {
	// TODO(M2.6): Call all Read* functions and assemble CgroupStats
	return nil, nil
}
