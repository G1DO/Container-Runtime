// Package namespace configures Linux namespace isolation for containers.
// Namespaces are the core mechanism that makes a process "not see" the host —
// its own PID space, network stack, mount tree, hostname, IPC, and user IDs.
//
// Milestones: M1.1 (PID), M1.2 (MNT), M1.3 (UTS), M1.4 (IPC), M1.5 (NET)
package namespace

import (
	"fmt"
	"syscall"

	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// CloneFlags returns the namespace flags used in phase 1.
// Phase 1 enables PID, mount, UTS, and IPC namespaces.
// Later milestones extend this when network and user namespace support lands.
func CloneFlags(config *specs.ContainerConfig) uintptr {
	return syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC
}

// SetupHostname sets the container's hostname via the UTS namespace.
// Requires CLONE_NEWUTS to have been set when creating the process.
func SetupHostname(hostname string) error {
	if hostname == "" {
		return nil
	}
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return fmt.Errorf("set hostname %q: %w", hostname, err)
	}
	return nil
}

// SetupLoopback brings up the loopback (lo) interface inside a new network namespace.
// A new NET namespace starts with zero interfaces — not even loopback.
func SetupLoopback() error {
	// TODO(M1.5): Use netlink to find "lo" and bring it up
	return nil
}

// JoinNamespaces enters the namespaces of an existing container process.
// Used by "exec" to run a new command inside a running container.
// Opens /proc/<pid>/ns/<type> and calls setns(2) for each namespace.
func JoinNamespaces(pid int) error {
	// TODO(M5.4): Open /proc/<pid>/ns/{pid,net,mnt,uts,ipc,user} and setns into each
	return nil
}
