// Package filesystem — devices.go creates essential device nodes inside the container.
// Without /dev/null most programs crash. Without /dev/urandom TLS hangs.
// Without /dev/pts interactive shells don't work.
//
// Milestones: M3.3 (device nodes and special mounts)
package filesystem

// DeviceNode describes a device file to create inside the container.
type DeviceNode struct {
	Path  string // e.g., "/dev/null"
	Mode  uint32 // e.g., 0666
	Major uint32 // Device major number
	Minor uint32 // Device minor number
}

// DefaultDevices returns the list of device nodes every container needs.
//
//	/dev/null    (1, 3) — discard everything written to it
//	/dev/zero    (1, 5) — infinite stream of zero bytes
//	/dev/full    (1, 7) — always returns "disk full" on write
//	/dev/random  (1, 8) — blocking random bytes
//	/dev/urandom (1, 9) — non-blocking random bytes
//	/dev/tty     (5, 0) — controlling terminal
func DefaultDevices() []DeviceNode {
	// TODO(M3.3): Return the device list above
	return nil
}

// CreateDevices sets up /dev inside the container rootfs:
// 1. Mount tmpfs on /dev (so we control the contents)
// 2. Create device nodes via mknod(2)
// 3. Create symlinks: /dev/fd → /proc/self/fd, /dev/stdin → fd/0, etc.
// 4. Mount devpts on /dev/pts (pseudoterminals)
// 5. Mount tmpfs on /dev/shm (POSIX shared memory)
func CreateDevices(rootfs string) error {
	// TODO(M3.3): Mount tmpfs on <rootfs>/dev
	// TODO(M3.3): mknod each device from DefaultDevices()
	// TODO(M3.3): Create /dev/fd, /dev/stdin, /dev/stdout, /dev/stderr symlinks
	// TODO(M3.3): Mount devpts on /dev/pts
	// TODO(M3.3): Mount tmpfs on /dev/shm
	return nil
}
