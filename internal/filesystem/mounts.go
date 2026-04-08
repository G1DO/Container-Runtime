// Package filesystem — mounts.go handles pivot_root and essential filesystem mounts.
// pivot_root swaps the root filesystem — unlike chroot, the old root is fully detached,
// preventing escape via open file descriptors.
//
// Milestones: M1.2 (mount namespace + pivot_root), M3.3 (essential mounts)
package filesystem

// PivotRoot performs the pivot_root(2) syscall to change the container's root filesystem.
//
// Steps:
//  1. Make all mounts private (MS_REC | MS_PRIVATE) — stop propagation to host
//  2. Bind-mount newRoot onto itself (pivot_root requires this)
//  3. Create a temporary directory for the old root
//  4. pivot_root(newRoot, putOld) — swap roots
//  5. chdir("/") — we're now inside the new root
//  6. Unmount and remove the old root
//
// Security: pivot_root > chroot because the old root is completely detached.
// With chroot, an attacker can escape via fchdir to an open fd outside the chroot.
func PivotRoot(newRoot string) error {
	// TODO(M1.2): Implement pivot_root sequence described above
	return nil
}

// MountProc mounts a new /proc inside the container.
// This is critical: without it, ps/top show host processes instead of container processes.
// Requires PID namespace to be active.
func MountProc(rootfs string) error {
	// TODO(M1.2): syscall.Mount("proc", <rootfs>/proc, "proc", 0, "")
	return nil
}

// MountSys mounts a read-only /sys inside the container.
// Prevents the container from modifying kernel parameters via sysfs.
func MountSys(rootfs string) error {
	// TODO(M3.3): syscall.Mount("sysfs", <rootfs>/sys, "sysfs", MS_RDONLY, "")
	return nil
}

// MountTmpfs mounts a tmpfs on /tmp inside the container.
func MountTmpfs(rootfs string) error {
	// TODO(M3.3): syscall.Mount("tmpfs", <rootfs>/tmp, "tmpfs", 0, "")
	return nil
}

// SetupContainerMounts performs all mount operations for a container in order:
// 1. PivotRoot (change filesystem root)
// 2. MountProc (new /proc for PID namespace)
// 3. MountSys (read-only /sys)
// 4. MountTmpfs (/tmp)
// 5. CreateDevices (/dev, device nodes, /dev/pts, /dev/shm)
func SetupContainerMounts(rootfs string) error {
	// TODO(M1.2): Call PivotRoot
	// TODO(M3.3): Call MountProc, MountSys, MountTmpfs, CreateDevices
	return nil
}
