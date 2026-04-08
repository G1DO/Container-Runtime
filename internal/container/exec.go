// Package container — exec.go runs a new command inside an existing container's namespaces.
// This is how "docker exec" works: it doesn't start a new container, it joins the
// namespaces of a running one using setns(2) on /proc/<pid>/ns/* files.
//
// Milestones: M5.4 (exec into running container)
package container

// Exec runs a command inside a running container by joining its namespaces.
//
// Steps:
//  1. Load container state, verify it's running
//  2. Open /proc/<pid>/ns/{pid,net,mnt,uts,ipc,user}
//  3. Call setns(2) for each namespace
//  4. Exec the command — now running inside the container
func (rt *Runtime) Exec(id string, command []string) error {
	// TODO(M5.4): Load container state, check StateRunning
	// TODO(M5.4): Call namespace.JoinNamespaces(container.Pid)
	// TODO(M5.4): syscall.Exec(command[0], command, env)
	return nil
}

// ReconcileOnStartup checks for containers marked as "running" whose processes
// have actually died (e.g., after a runtime crash). Cleans up orphaned state.
func (rt *Runtime) ReconcileOnStartup() error {
	// TODO(M8.3): List all containers
	// TODO(M8.3): For each "running" container, check if PID is alive (kill -0)
	// TODO(M8.3): If dead, update state to stopped and cleanup cgroup
	return nil
}
