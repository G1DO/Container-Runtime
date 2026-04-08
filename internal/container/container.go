// Package container is the top-level orchestrator that ties together
// namespaces, cgroups, filesystem, image, and networking into a working container.
//
// container.go defines the Runtime and its configuration.
//
// Milestones: M5.1–M5.6 (full container lifecycle)
package container

import (
	"github.com/G1DO/Container-Runtime/internal/cgroup"
	"github.com/G1DO/Container-Runtime/internal/filesystem"
	imagestore "github.com/G1DO/Container-Runtime/internal/image"
	"github.com/G1DO/Container-Runtime/internal/network"
	"github.com/G1DO/Container-Runtime/internal/store"
)

// Runtime is the top-level container runtime. It holds references to all subsystems
// and coordinates container lifecycle operations.
type Runtime struct {
	Store        *store.ContainerStore
	CgroupMgr   *cgroup.Manager
	LayerStore   *filesystem.LayerStore
	ImageStore   *imagestore.Store
	IPAM         *network.IPAM
	BridgeConfig network.BridgeConfig
	RuntimeRoot  string // Base directory for all runtime data.
}

// NewRuntime initializes the runtime and all its subsystems.
func NewRuntime(root string) (*Runtime, error) {
	// TODO(M5.1): Initialize store, cgroup manager, layer store, image store, IPAM
	// TODO(M6.1): Create network bridge
	// TODO(M8.3): Run reconcileOnStartup to clean up after any previous crash
	return nil, nil
}

// GenerateID creates a random 12-character hex container ID.
func GenerateID() string {
	// TODO(M5.2): crypto/rand → hex encode → truncate to 12 chars
	return ""
}
