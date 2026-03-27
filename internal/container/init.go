// Package container — init.go is the code that runs INSIDE the container's namespaces.
// After the parent forks with CLONE_NEW* flags, the child process executes this code
// to set up the container environment before exec'ing the user's entrypoint.
//
// This is the most sensitive code path: it runs as PID 1 inside the container.
// PID 1 has special responsibilities in Linux — it must reap orphaned children
// and handle signals properly.
//
// Milestones: M5.3 (init process), M8.1 (signal forwarding), M8.2 (zombie reaping)
package container

// ContainerInit is the entry point for the container init process.
// It runs inside the new namespaces after the parent has set up cgroups.
//
// Sequence:
//  1. Wait for parent to signal "ready" (cgroup setup complete)
//  2. Setup mounts (pivot_root, /proc, /sys, /dev)
//  3. Set hostname
//  4. Configure environment variables
//  5. Change to working directory
//  6. Start zombie reaper goroutine
//  7. Setup signal forwarding
//  8. Exec the entrypoint (replaces this process)
func ContainerInit(containerID string) error {
	// TODO(M5.3): Wait for parent signal via pipe
	// TODO(M1.2): Call filesystem.SetupContainerMounts(rootfs)
	// TODO(M1.3): Call namespace.SetupHostname(hostname)
	// TODO(M5.3): Set environment variables from config
	// TODO(M5.3): Chdir to working directory
	// TODO(M8.2): Start zombie reaper
	// TODO(M8.1): Setup signal forwarding
	// TODO(M5.3): syscall.Exec(entrypoint, args, env)
	return nil
}

// StartReaper launches a background goroutine that reaps zombie processes.
// PID 1 must call wait(2) on dead children, otherwise they accumulate as zombies.
// Most application processes don't do this — that's why container runtimes
// either provide a built-in init or use a tiny init like tini.
func StartReaper() {
	// TODO(M8.2): Loop calling Wait4(-1, WNOHANG) to reap orphaned children
}

// ForwardSignals sets up signal forwarding from the runtime to the container.
// When the runtime receives SIGTERM/SIGINT/SIGQUIT, it forwards them to the
// container's PID 1 so the application can shut down gracefully.
func ForwardSignals(containerPid int) {
	// TODO(M8.1): signal.Notify for SIGTERM, SIGINT, SIGQUIT
	// TODO(M8.1): Forward each received signal to containerPid via syscall.Kill
}
