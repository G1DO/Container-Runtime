// Package tests — uts_test.go verifies UTS namespace hostname isolation.
// The container should be able to set its own hostname without affecting the host.
//
// IMPORTANT: Requires root privileges and a Linux kernel with UTS namespace support.
// Run with: sudo go test -v ./tests/ -tags=integration
//
// Milestones: M1.3 (UTS namespace — hostname isolation)
package tests

import "testing"

// TestContainerHostname verifies that the container can set its own hostname
// and that it takes effect inside the namespace.
func TestContainerHostname(t *testing.T) {
	// TODO(M1.3): Start container with Hostname set to "test-container"
	// TODO(M1.3): Run "hostname" inside container — should return "test-container"
	t.Skip("not yet implemented")
}

// TestHostHostnameUnchanged verifies that setting a hostname inside the container
// does not modify the host's hostname.
func TestHostHostnameUnchanged(t *testing.T) {
	// TODO(M1.3): Record host hostname before container start
	// TODO(M1.3): Start container that sets hostname to "container-test"
	// TODO(M1.3): After container exits, verify host hostname is unchanged
	t.Skip("not yet implemented")
}
