// Package network — veth.go creates veth pairs and connects containers to the bridge.
// A veth pair is a virtual ethernet cable: one end in the container's network namespace,
// the other end attached to the host bridge. Data sent into one end comes out the other.
//
// Milestones: M6.2 (veth pair setup), M6.6 (DNS)
package network

// CreateVethPair creates a virtual ethernet pair and attaches the host end to the bridge.
//  1. Create veth pair (host end + container end)
//  2. Move container end into the container's network namespace
//  3. Attach host end to the bridge
//  4. Bring host end up
func CreateVethPair(hostName string, containerName string, containerPid int, bridgeName string) error {
	// TODO(M6.2): netlink.LinkAdd veth pair
	// TODO(M6.2): netlink.LinkSetNsPid to move peer into container
	// TODO(M6.2): netlink.LinkSetMaster to attach host end to bridge
	// TODO(M6.2): netlink.LinkSetUp host end
	return nil
}

// ConfigureContainerNetwork sets up networking inside the container's namespace:
//  1. Assign IP address to eth0
//  2. Bring eth0 up
//  3. Add default route via bridge gateway
func ConfigureContainerNetwork(containerPid int, ip string, gateway string, iface string) error {
	// TODO(M6.2): Enter netns, assign IP, bring up interface, add default route
	return nil
}

// SetupDNS writes /etc/resolv.conf inside the container.
// Without this, the container can't resolve domain names.
func SetupDNS(rootfs string, nameservers []string) error {
	// TODO(M6.6): Write "nameserver X.X.X.X\n" for each nameserver to <rootfs>/etc/resolv.conf
	return nil
}

// CleanupVeth removes a veth pair. Deleting one end automatically deletes the other.
func CleanupVeth(hostName string) error {
	// TODO(M6.2): netlink.LinkDel
	return nil
}
