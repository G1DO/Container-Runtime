// Package namespace — userns.go handles User namespace UID/GID mapping.
// The user namespace is the most security-critical namespace: without it,
// root inside the container IS root on the host.
//
// Milestones: M1.6 (user namespace and UID remapping)
package namespace

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	defaultHostIDMapStart = 100000
	defaultIDMapSize      = 65536
)

// DefaultUIDMappings returns UID mappings that map container root (UID 0)
// to an unprivileged host UID (e.g., 100000), with 65536 UIDs available.
//
// Inside container:  UID 0  → Host UID 100000
// Inside container:  UID 1  → Host UID 100001
// ...up to 65536 UIDs.
func DefaultUIDMappings() []syscall.SysProcIDMap {
	return []syscall.SysProcIDMap{{
		ContainerID: 0,
		HostID:      defaultHostIDMapStart,
		Size:        defaultIDMapSize,
	}}
}

// DefaultGIDMappings returns GID mappings analogous to DefaultUIDMappings.
func DefaultGIDMappings() []syscall.SysProcIDMap {
	return []syscall.SysProcIDMap{{
		ContainerID: 0,
		HostID:      defaultHostIDMapStart,
		Size:        defaultIDMapSize,
	}}
}

// WriteIDMappings writes UID and GID mappings to /proc/<pid>/uid_map and gid_map.
// This must be done from the parent process after clone but before the child execs.
func WriteIDMappings(pid int, uidMappings, gidMappings []syscall.SysProcIDMap) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid %d", pid)
	}
	return writeIDMappings(filepath.Join("/proc", strconv.Itoa(pid)), uidMappings, gidMappings)
}

func writeIDMappings(procPidDir string, uidMappings, gidMappings []syscall.SysProcIDMap) error {
	var uidMap string
	if len(uidMappings) > 0 {
		var err error
		uidMap, err = formatIDMappings(uidMappings)
		if err != nil {
			return fmt.Errorf("format uid mappings: %w", err)
		}
	}

	var gidMap string
	if len(gidMappings) > 0 {
		var err error
		gidMap, err = formatIDMappings(gidMappings)
		if err != nil {
			return fmt.Errorf("format gid mappings: %w", err)
		}
	}

	if len(uidMappings) > 0 {
		if err := os.WriteFile(filepath.Join(procPidDir, "uid_map"), []byte(uidMap), 0o644); err != nil {
			return fmt.Errorf("write uid_map: %w", err)
		}
	}

	if len(gidMappings) > 0 {
		if err := os.WriteFile(filepath.Join(procPidDir, "setgroups"), []byte("deny\n"), 0o644); err != nil {
			return fmt.Errorf("write setgroups deny: %w", err)
		}
		if err := os.WriteFile(filepath.Join(procPidDir, "gid_map"), []byte(gidMap), 0o644); err != nil {
			return fmt.Errorf("write gid_map: %w", err)
		}
	}

	return nil
}

func formatIDMappings(mappings []syscall.SysProcIDMap) (string, error) {
	var b strings.Builder
	for i, mapping := range mappings {
		if mapping.ContainerID < 0 {
			return "", fmt.Errorf("mapping %d has negative container id %d", i, mapping.ContainerID)
		}
		if mapping.HostID < 0 {
			return "", fmt.Errorf("mapping %d has negative host id %d", i, mapping.HostID)
		}
		if mapping.Size <= 0 {
			return "", fmt.Errorf("mapping %d has non-positive size %d", i, mapping.Size)
		}
		fmt.Fprintf(&b, "%d %d %d\n", mapping.ContainerID, mapping.HostID, mapping.Size)
	}
	return b.String(), nil
}
