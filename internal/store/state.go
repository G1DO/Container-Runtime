// Package store handles container state persistence to disk.
// State is stored as JSON files under /var/lib/myruntime/containers/<id>/state.json.
// This is how the runtime remembers containers across restarts.
//
// Milestones: M5.1 (state persistence), M8.3 (crash recovery reconciliation)
package store

import (
	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// ContainerStore manages reading and writing container state to the filesystem.
type ContainerStore struct {
	// Root is the base directory for all container state
	// (e.g., /var/lib/myruntime/containers/).
	Root string
}

// NewContainerStore creates a store rooted at the given directory.
func NewContainerStore(root string) *ContainerStore {
	// TODO(M5.1): Create root directory if it doesn't exist
	return nil
}

// Save persists a container's state to disk as JSON.
func (cs *ContainerStore) Save(container *specs.Container) error {
	// TODO(M5.1): Marshal container to JSON and write to <root>/<id>/state.json
	return nil
}

// Load reads a container's state from disk.
func (cs *ContainerStore) Load(id string) (*specs.Container, error) {
	// TODO(M5.1): Read JSON from <root>/<id>/state.json and unmarshal
	return nil, nil
}

// List returns all known containers.
func (cs *ContainerStore) List() ([]*specs.Container, error) {
	// TODO(M5.1): Walk root directory, load each container's state
	return nil, nil
}

// Delete removes a container's state directory from disk.
func (cs *ContainerStore) Delete(id string) error {
	// TODO(M5.6): Remove <root>/<id>/ directory entirely
	return nil
}
