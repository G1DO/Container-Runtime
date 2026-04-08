// Package image — unpack.go extracts OCI image layers (tar.gz) to disk.
// Each layer is extracted to a content-addressable directory (keyed by SHA256).
// Identical layers are shared across images — extract once, use many times.
//
// Milestones: M4.4 (layer unpacking)
package image

// UnpackLayers extracts all layers of an image to the layer store.
// Returns ordered layer paths (bottom to top) for use as overlay lower dirs.
func UnpackLayers(layerStore string, image []byte) ([]string, error) {
	// TODO(M4.4): For each layer: decompress gzip, extract tar, store by digest
	return nil, nil
}

// ExtractTarGz decompresses and extracts a .tar.gz archive to the destination directory.
// Must preserve file permissions, ownership, symlinks, and special files.
func ExtractTarGz(src string, dest string) error {
	// TODO(M4.4): Open gzip reader → tar reader → extract entries
	// TODO(M4.4): Handle whiteout files (.wh.*) for layer deletions
	return nil
}

// SHA256File computes the SHA256 digest of a file.
// Used for content-addressable layer identification.
func SHA256File(path string) (string, error) {
	// TODO(M4.4): Read file, compute sha256, return hex string
	return "", nil
}
