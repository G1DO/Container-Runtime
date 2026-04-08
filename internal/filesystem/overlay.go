// Package filesystem handles overlay filesystem setup for containers.
// OverlayFS merges multiple read-only layers with one writable layer,
// enabling copy-on-write semantics and efficient image sharing.
//
// Milestones: M3.1 (basic overlay mount), M3.2 (layer management)
package filesystem

// OverlayFS represents an overlay mount with lower (read-only) layers,
// an upper (writable) layer, a work directory (kernel internal), and
// the merged view that the container sees as its root filesystem.
type OverlayFS struct {
	LowerDirs []string // Read-only layers, bottom to top.
	UpperDir  string   // Writable layer (container-specific).
	WorkDir   string   // Internal work directory for atomic operations.
	MergedDir string   // The combined view — this becomes the container's rootfs.
}

// Mount creates the overlay mount.
// Kernel mount options: lowerdir=L1:L2:...,upperdir=U,workdir=W
func (o *OverlayFS) Mount() error {
	// TODO(M3.1): Create upper, work, merged dirs
	// TODO(M3.1): syscall.Mount("overlay", mergedDir, "overlay", 0, opts)
	return nil
}

// Unmount detaches the overlay mount.
func (o *OverlayFS) Unmount() error {
	// TODO(M3.1): syscall.Unmount(mergedDir, 0)
	return nil
}

// Cleanup unmounts and removes the container-specific upper and work directories.
// Lower layers are shared and must NOT be removed here.
func (o *OverlayFS) Cleanup() error {
	// TODO(M3.1): Unmount, then remove upper and work dirs
	return nil
}

// LayerStore manages content-addressable layer storage.
// Layers are identified by SHA256 digest and shared between containers.
type LayerStore struct {
	// Root is the base directory for layers (e.g., /var/lib/myruntime/layers/).
	Root string
}

// NewLayerStore creates a layer store at the given root.
func NewLayerStore(root string) *LayerStore {
	// TODO(M3.2): Create root directory
	return nil
}

// Extract unpacks a tar archive into a content-addressed layer directory.
// Returns the layer path. If the layer already exists (same SHA256), skips extraction.
func (ls *LayerStore) Extract(tarPath string) (string, error) {
	// TODO(M3.2): SHA256 of tar → layer ID
	// TODO(M3.2): Extract tar to <root>/<sha256>/
	// TODO(M3.2): Skip if already exists (content-addressable dedup)
	return "", nil
}

// PrepareContainer creates an OverlayFS for a specific container,
// using the given layers as read-only lower dirs.
func (ls *LayerStore) PrepareContainer(containerID string, layers []string) *OverlayFS {
	// TODO(M3.2): Create upper/work/merged paths under containers/<id>/
	return nil
}

// CleanupContainer removes a container's writable layer and work directory.
func (ls *LayerStore) CleanupContainer(containerID string) error {
	// TODO(M5.6): Remove containers/<id>/{upper,work,merged}
	return nil
}
