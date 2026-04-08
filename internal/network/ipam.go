// Package network — ipam.go manages IP address allocation for containers.
// Each container gets a unique IP from the bridge subnet (e.g., 172.20.0.0/24).
// Gateway (.1) is reserved for the bridge. Containers start at .2.
//
// Milestones: M6.3 (IP address management)
package network

import (
	"net"

	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// IPAM (IP Address Management) tracks which IPs are allocated to which containers.
type IPAM struct {
	Subnet    net.IPNet
	Gateway   net.IP
	Allocated map[string]net.IP // containerID → assigned IP
}

// NewIPAM creates an IPAM for the given subnet.
// Gateway is the first usable IP (e.g., 172.20.0.1 for 172.20.0.0/24).
func NewIPAM(subnet string) (*IPAM, error) {
	// TODO(M6.3): Parse CIDR, compute gateway, initialize allocation map
	return nil, nil
}

// Allocate assigns the next available IP to a container.
// Returns error if the subnet is exhausted.
func (ipam *IPAM) Allocate(containerID string) (net.IP, error) {
	// TODO(M6.3): Find next unallocated IP in subnet, starting from .2
	return nil, nil
}

// Release frees an IP address when a container is deleted.
func (ipam *IPAM) Release(containerID string) {
	// TODO(M6.3): Remove from allocated map
}

// GetAllocation returns the current IP for a container, if any.
func (ipam *IPAM) GetAllocation(containerID string) *specs.IPAllocation {
	// TODO(M6.3): Look up in allocated map
	return nil
}
