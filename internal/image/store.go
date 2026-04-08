// Package image — store.go manages local image storage.
// Images are stored on disk with their manifest, config, and layer references.
//
// Milestones: M4.5 (config parsing), M4.6 (local image storage)
package image

import (
	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// Store manages locally stored OCI images.
type Store struct {
	// Root is the base directory (e.g., /var/lib/myruntime/images/).
	Root string
}

// NewStore creates an image store at the given root.
func NewStore(root string) *Store {
	// TODO(M4.6): Create root directory
	return nil
}

// Save persists an image's manifest and config to disk.
func (s *Store) Save(image *specs.Image) error {
	// TODO(M4.6): Write manifest.json and config.json under <root>/<digest>/
	return nil
}

// Load reads an image by reference from disk.
func (s *Store) Load(ref specs.ImageReference) (*specs.Image, error) {
	// TODO(M4.6): Read manifest and config, resolve layer paths
	return nil, nil
}

// List returns all locally stored images.
func (s *Store) List() ([]*specs.Image, error) {
	// TODO(M4.6): Walk root directory and load each image
	return nil, nil
}

// Resolve returns the layer paths and image config for a given image reference.
// Used by container Create to set up the overlay filesystem.
func (s *Store) Resolve(ref specs.ImageReference) ([]string, *specs.ImageConfig, error) {
	// TODO(M4.6): Load image, return LayerPaths and Config
	return nil, nil, nil
}

// ParseImageConfig parses an OCI image config JSON blob into an ImageConfig.
func ParseImageConfig(data []byte) (*specs.ImageConfig, error) {
	// TODO(M4.5): JSON unmarshal into ImageConfig
	return nil, nil
}
