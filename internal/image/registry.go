// Package image handles OCI image operations: pulling from registries,
// local storage, and layer extraction.
//
// registry.go implements the OCI Distribution Spec client for pulling images
// from Docker Hub and other OCI-compliant registries.
//
// The pull flow:
//  1. Authenticate — GET token from auth.docker.io
//  2. Fetch manifest — GET /v2/<repo>/manifests/<tag>
//  3. Fetch config blob — GET /v2/<repo>/blobs/<config-digest>
//  4. Fetch layer blobs — GET /v2/<repo>/blobs/<layer-digest> (parallel)
//
// Milestones: M4.1 (auth), M4.2 (manifest), M4.3 (layer download)
package image

import (
	"net/http"

	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// RegistryClient talks to OCI-compliant container registries.
type RegistryClient struct {
	HTTPClient *http.Client
}

// NewRegistryClient creates a registry client with default HTTP settings.
func NewRegistryClient() *RegistryClient {
	// TODO(M4.1): Create HTTP client with reasonable timeouts
	return nil
}

// Pull downloads an image from a registry: manifest, config, and all layers.
func (rc *RegistryClient) Pull(ref specs.ImageReference) (*specs.Image, error) {
	// TODO(M4.1): Authenticate with registry
	// TODO(M4.2): Fetch and parse manifest
	// TODO(M4.3): Fetch config blob
	// TODO(M4.3): Fetch layer blobs in parallel
	return nil, nil
}

// authenticate gets a bearer token for the given image reference.
func (rc *RegistryClient) authenticate(ref specs.ImageReference) (string, error) {
	// TODO(M4.1): GET https://auth.docker.io/token?service=...&scope=repository:...:pull
	return "", nil
}

// getManifest fetches and parses the OCI image manifest.
func (rc *RegistryClient) getManifest(ref specs.ImageReference, token string) (*specs.ImageManifest, error) {
	// TODO(M4.2): GET /v2/<repo>/manifests/<tag> with Accept header
	return nil, nil
}

// getBlob downloads a blob (config or layer) by its digest.
func (rc *RegistryClient) getBlob(ref specs.ImageReference, digest string, token string) ([]byte, error) {
	// TODO(M4.3): GET /v2/<repo>/blobs/<digest>
	return nil, nil
}
