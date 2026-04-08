// Package network — iptables.go manages port mapping via iptables DNAT rules.
// Port mapping allows exposing a container's port on the host:
// host:8080 → container:80 (DNAT rewrites the destination address).
//
// Milestones: M6.5 (port mapping)
package network

import (
	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// AddPortMapping creates a DNAT iptables rule to forward traffic from
// a host port to a container's IP and port.
//
// Example: -p 8080:80 creates:
//
//	iptables -t nat -A PREROUTING -p tcp --dport 8080 -j DNAT --to-destination 172.20.0.2:80
func AddPortMapping(mapping specs.PortMapping, containerIP string) error {
	// TODO(M6.5): Execute iptables DNAT rule
	return nil
}

// RemovePortMapping removes a DNAT rule.
func RemovePortMapping(mapping specs.PortMapping, containerIP string) error {
	// TODO(M6.5): Execute iptables -D to remove the rule
	return nil
}

// AddPortMappings adds multiple port mapping rules.
func AddPortMappings(mappings []specs.PortMapping, containerIP string) error {
	// TODO(M6.5): Loop over mappings and call AddPortMapping
	return nil
}

// RemovePortMappings removes multiple port mapping rules.
func RemovePortMappings(mappings []specs.PortMapping, containerIP string) error {
	// TODO(M6.5): Loop over mappings and call RemovePortMapping
	return nil
}
