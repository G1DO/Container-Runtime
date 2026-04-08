// Package container — lifecycle.go implements the container lifecycle operations:
// Create, Start, Stop, Kill, Delete. This is where all subsystems come together.
//
// Milestones: M5.2 (create), M5.3 (start), M5.5 (stop/kill), M5.6 (delete)
package container

import (
	"syscall"
	"time"

	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// Create prepares a container without starting it. Sets up:
//  1. Resolve image → get layers and config
//  2. Overlay filesystem → mount layers
//  3. Cgroup → create with resource limits
//  4. Persist state as "created"
//
// The container is ready to Start but no process is running yet.
func (rt *Runtime) Create(config specs.ContainerConfig) (*specs.Container, error) {
	// TODO(M5.2): Generate ID
	// TODO(M5.2): Resolve image layers and config
	// TODO(M5.2): Prepare and mount overlay filesystem
	// TODO(M5.2): Create cgroup with resource limits
	// TODO(M5.2): Save container state as StateCreated
	return nil, nil
}

// Start launches the container's init process in new namespaces.
//
// The sequence is:
//  1. Fork child via re-exec pattern with CLONE_NEW* flags
//  2. Add child PID to cgroup (from parent, before child execs)
//  3. Signal child to proceed (via pipe synchronization)
//  4. Child sets up mounts, hostname, env, then execs entrypoint
//  5. Update state to "running"
//
// The parent-child synchronization via pipe is critical:
// the child must wait for the parent to finish cgroup setup,
// otherwise the entrypoint runs without resource limits.
func (rt *Runtime) Start(id string) error {
	// TODO(M5.3): Load container state
	// TODO(M5.3): Fork with clone flags (re-exec pattern)
	// TODO(M5.3): Add child to cgroup
	// TODO(M6.2): Setup container networking (veth, IP)
	// TODO(M5.3): Signal child to proceed
	// TODO(M5.3): Update state to StateRunning
	return nil
}

// Stop gracefully stops a container: SIGTERM first, then SIGKILL after timeout.
func (rt *Runtime) Stop(id string, timeout time.Duration) error {
	// TODO(M5.5): Load container, send SIGTERM
	// TODO(M5.5): Wait up to timeout for exit
	// TODO(M5.5): If still alive, send SIGKILL
	// TODO(M5.5): Update state to StateStopped, record exit code
	return nil
}

// Kill sends a signal directly to a container's init process.
func (rt *Runtime) Kill(id string, signal syscall.Signal) error {
	// TODO(M5.5): Load container, syscall.Kill(pid, signal)
	return nil
}

// Delete removes all resources for a stopped container:
//  1. Destroy cgroup
//  2. Unmount and cleanup overlay filesystem
//  3. Release IP address
//  4. Remove iptables port mapping rules
//  5. Remove state from disk
func (rt *Runtime) Delete(id string) error {
	// TODO(M5.6): Verify container is stopped (can't delete running container)
	// TODO(M5.6): Destroy cgroup
	// TODO(M5.6): Cleanup overlay filesystem
	// TODO(M6.3): Release IP address
	// TODO(M6.5): Remove port mappings
	// TODO(M5.6): Delete state from store
	return nil
}

// List returns all containers with their current state.
func (rt *Runtime) List() ([]*specs.Container, error) {
	// TODO(M5.1): Delegate to store.List()
	return nil, nil
}

// Inspect returns detailed information about a single container.
func (rt *Runtime) Inspect(id string) (*specs.Container, error) {
	// TODO(M7.4): Load container state with full details
	return nil, nil
}
