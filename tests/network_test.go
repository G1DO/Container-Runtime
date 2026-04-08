// Package tests — network_test.go verifies container networking.
//
// Milestones: M9.1 (network isolation), M6.7 (container-to-container)
package tests

import "testing"

// TestNetworkIsolation verifies the container has its own network namespace
// and cannot see host interfaces.
func TestNetworkIsolation(t *testing.T) {
	// TODO(M9.1): Start container, run "ip link show"
	// TODO(M9.1): Should only see lo and eth0 (veth), not host interfaces
	t.Skip("not yet implemented")
}

// TestContainerToContainerPing verifies two containers on the same bridge
// can reach each other by IP.
func TestContainerToContainerPing(t *testing.T) {
	// TODO(M6.7): Start container A and container B
	// TODO(M6.7): From A, ping B's IP address
	// TODO(M6.7): Verify ping succeeds
	t.Skip("not yet implemented")
}

// TestContainerInternetAccess verifies a container can reach the internet via NAT.
func TestContainerInternetAccess(t *testing.T) {
	// TODO(M6.4): Start container
	// TODO(M6.4): Ping 8.8.8.8 from inside container
	// TODO(M6.4): Verify ping succeeds
	t.Skip("not yet implemented")
}

// TestPortMapping verifies that exposing a container port makes it
// reachable from the host.
func TestPortMapping(t *testing.T) {
	// TODO(M6.5): Start container with -p 8080:80 running a simple HTTP server
	// TODO(M6.5): From host, curl localhost:8080
	// TODO(M6.5): Verify response from container
	t.Skip("not yet implemented")
}

// TestDNSResolution verifies that the container can resolve domain names.
func TestDNSResolution(t *testing.T) {
	// TODO(M6.6): Start container
	// TODO(M6.6): Run "nslookup google.com" inside
	// TODO(M6.6): Verify resolution succeeds
	t.Skip("not yet implemented")
}
