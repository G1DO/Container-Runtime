// Package network handles container networking: bridge creation, veth pairs,
// IP allocation, NAT, and port mapping.
//
// bridge.go creates and manages the virtual network bridge that connects containers.
// The bridge acts as a virtual switch — containers plug into it via veth pairs.
//
// Milestones: M6.1 (bridge creation), M6.4 (NAT and IP forwarding)
package network

// BridgeConfig holds configuration for the container network bridge.
type BridgeConfig struct {
	Name    string // Bridge interface name (e.g., "myruntime0").
	Subnet  string // CIDR notation (e.g., "172.20.0.1/24").
	Gateway string // Bridge IP, also the default gateway for containers.
}

// CreateBridge sets up a virtual network bridge on the host.
//  1. Create bridge interface via netlink
//  2. Assign IP address (this becomes the gateway for containers)
//  3. Bring the bridge up
//  4. Enable IP forwarding (/proc/sys/net/ipv4/ip_forward = 1)
func CreateBridge(config BridgeConfig) error {
	// TODO(M6.1): netlink.LinkAdd bridge, AddrAdd, LinkSetUp
	// TODO(M6.1): Write "1" to /proc/sys/net/ipv4/ip_forward
	return nil
}

// SetupNAT adds iptables MASQUERADE rules so containers can reach the internet.
// Outbound packets from the container subnet get their source IP rewritten
// to the host's external IP — the internet sees the host, not the container.
func SetupNAT(subnet string, hostInterface string) error {
	// TODO(M6.4): iptables -t nat -A POSTROUTING -s <subnet> -o <iface> -j MASQUERADE
	return nil
}

// CleanupNAT removes the NAT rules.
func CleanupNAT(subnet string, hostInterface string) error {
	// TODO(M6.4): iptables -t nat -D POSTROUTING ...
	return nil
}

// DeleteBridge removes the bridge interface.
func DeleteBridge(name string) error {
	// TODO(M6.1): netlink.LinkDel bridge
	return nil
}
