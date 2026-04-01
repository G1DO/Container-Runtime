// Package namespace — userns.go handles User namespace UID/GID mapping.
// The user namespace is the most security-critical namespace: without it,
// root inside the container IS root on the host.
//
// Milestones: M1.6 (user namespace and UID remapping)
package namespace

import (
	"syscall"
)

// DefaultUIDMappings returns UID mappings that map container root (UID 0)
// to an unprivileged host UID (e.g., 100000), with 65536 UIDs available.
//
// Inside container:  UID 0  → Host UID 100000
// Inside container:  UID 1  → Host UID 100001
// ...up to 65536 UIDs.
func DefaultUIDMappings() []syscall.SysProcIDMap {
	// TODO(M1.6): Return mapping {ContainerID: 0, HostID: 100000, Size: 65536}
	return nil
}

// DefaultGIDMappings returns GID mappings analogous to DefaultUIDMappings.
func DefaultGIDMappings() []syscall.SysProcIDMap {
	// TODO(M1.6): Return mapping {ContainerID: 0, HostID: 100000, Size: 65536}
	return nil
}

// WriteIDMappings writes UID and GID mappings to /proc/<pid>/uid_map and gid_map.
// This must be done from the parent process after clone but before the child execs.
func WriteIDMappings(pid int, uidMappings, gidMappings []syscall.SysProcIDMap) error {
	// TODO(M1.6): Write mappings to /proc/<pid>/uid_map and /proc/<pid>/gid_map
	// Note: Must also write "deny" to /proc/<pid>/setgroups before writing gid_map
	return nil
}
