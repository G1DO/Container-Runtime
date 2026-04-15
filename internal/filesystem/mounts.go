// Package filesystem — mounts.go handles pivot_root and essential filesystem mounts.
// pivot_root swaps the root filesystem — unlike chroot, the old root is fully detached,
// preventing escape via open file descriptors.
//
// Milestones: M1.2 (mount namespace + pivot_root), M3.3 (essential mounts)
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

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
	if newRoot == "" {
		return fmt.Errorf("pivot_root: newRoot is empty")
	}
	if !filepath.IsAbs(newRoot) {
		return fmt.Errorf("pivot_root: newRoot must be absolute: %q", newRoot)
	}
	st, err := os.Stat(newRoot)
	if err != nil {
		return fmt.Errorf("pivot_root: stat newRoot %q: %w", newRoot, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("pivot_root: newRoot is not a directory: %q", newRoot)
	}

	// Stop mount propagation back to the host.
	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("pivot_root: make / private: %w", err)
	}

	// pivot_root requires newRoot to be a mountpoint; bind-mount it onto itself.
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("pivot_root: bind-mount newRoot: %w", err)
	}

	putOld := filepath.Join(newRoot, ".pivot_root")
	if err := os.MkdirAll(putOld, 0o700); err != nil {
		return fmt.Errorf("pivot_root: mkdir putOld: %w", err)
	}

	if err := syscall.PivotRoot(newRoot, putOld); err != nil {
		return fmt.Errorf("pivot_root: pivot_root(%q, %q): %w", newRoot, putOld, err)
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("pivot_root: chdir(/): %w", err)
	}

	// After pivot_root, the old root is mounted at "/.pivot_root".
	if err := syscall.Unmount("/.pivot_root", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("pivot_root: unmount old root: %w", err)
	}
	if err := os.RemoveAll("/.pivot_root"); err != nil {
		return fmt.Errorf("pivot_root: remove old root dir: %w", err)
	}

	return nil
}

// MountProc mounts a new /proc inside the container.
// This is critical: without it, ps/top show host processes instead of container processes.
// Requires PID namespace to be active.
func MountProc(rootfs string) error {
	_ = rootfs // rootfs is only relevant before pivot_root; after pivot_root we mount at /proc.

	if err := os.MkdirAll("/proc", 0o555); err != nil {
		return fmt.Errorf("mount proc: mkdir /proc: %w", err)
	}
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mount proc: %w", err)
	}
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
	if err := PivotRoot(rootfs); err != nil {
		return err
	}
	if err := MountProc(""); err != nil {
		return err
	}
	// TODO(M3.3): Call MountProc, MountSys, MountTmpfs, CreateDevices
	return nil
}
