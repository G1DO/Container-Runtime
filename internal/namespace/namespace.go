// Package namespace configures Linux namespace isolation for containers.
// Namespaces are the core mechanism that makes a process "not see" the host —
// its own PID space, network stack, mount tree, hostname, IPC, and user IDs.
//
// Milestones: M1.1 (PID), M1.2 (MNT), M1.3 (UTS), M1.4 (IPC), M1.5 (NET)
package namespace

import (
	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// CloneFlags returns the combined CLONE_NEW* flags for all requested namespaces.
// These flags are passed to clone(2) / unshare(2) when forking the container process.
func CloneFlags(config *specs.ContainerConfig) uintptr {
	// TODO(M1.1): Return CLONE_NEWPID | CLONE_NEWNS | CLONE_NEWUTS | CLONE_NEWIPC | CLONE_NEWNET | CLONE_NEWUSER
	return 0
}

// SetupHostname sets the container's hostname via the UTS namespace.
// Requires CLONE_NEWUTS to have been set when creating the process.
func SetupHostname(hostname string) error {
	// TODO(M1.3): syscall.Sethostname([]byte(hostname))
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
