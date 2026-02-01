// Package specs defines OCI-compatible configuration types shared across
// all internal packages. This is the contract layer — every package talks
// through these types.
//
// Milestones: M4.5 (image config), M5.1 (container config), M2.1 (resource config)
package specs

import (
	"net"
	"time"
)

// ContainerState represents the lifecycle state of a container.
type ContainerState string

const (
	StateCreated ContainerState = "created"
	StateRunning ContainerState = "running"
	StateStopped ContainerState = "stopped"
)

// ContainerConfig holds the full configuration for creating a container.
// Merges image defaults with user overrides.
type ContainerConfig struct {
	// Image is the image reference (e.g., "nginx:latest").
	Image string

	// Command overrides the image's CMD.
	Command []string

	// Entrypoint overrides the image's ENTRYPOINT.
	Entrypoint []string

	// Env is a list of "KEY=VALUE" environment variables.
	Env []string

	// WorkingDir sets the initial working directory inside the container.
	WorkingDir string

	// Hostname for the UTS namespace.
	Hostname string

	// User to run as inside the container (e.g., "nginx" or "1000:1000").
	User string

	// Resources defines cgroup resource limits.
	Resources ResourceConfig

	// Network configuration.
	Network NetworkConfig

	// PortMappings maps host ports to container ports.
	PortMappings []PortMapping
}

// ResourceConfig defines cgroup v2 resource limits for a container.
type ResourceConfig struct {
	// CPUQuota in microseconds per CPUPeriod (e.g., 50000 = 50% of one core).
	CPUQuota int64

	// CPUPeriod in microseconds (default: 100000 = 100ms).
	CPUPeriod int64

	// MemoryMax is the hard memory limit in bytes.
	MemoryMax int64

	// PidsMax is the maximum number of processes (including threads).
	PidsMax int64

	// IOMax is the IO limit string: "MAJOR:MINOR rbps=BYTES wbps=BYTES".
	IOMax string
}

// NetworkConfig defines container networking options.
type NetworkConfig struct {
	// BridgeName is the host bridge to connect to (e.g., "myruntime0").
	BridgeName string

	// Subnet for IP allocation (e.g., "172.20.0.0/24").
	Subnet string

	// DNS nameservers to write into /etc/resolv.conf.
	DNS []string
}

// PortMapping maps a host port to a container port.
type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string // "tcp" or "udp"
}

// Container represents a container's full runtime state.
type Container struct {
	ID         string
	State      ContainerState
	Pid        int // Host PID of the container init process (0 if not running).
	Config     ContainerConfig
	CreatedAt  time.Time
	StartedAt  time.Time
	FinishedAt time.Time
	ExitCode   int
	RootFS     string // Path to the merged overlay mount.
	CgroupPath string // Path to the container's cgroup directory.
}

// ImageReference identifies an image in a registry.
type ImageReference struct {
	Registry   string // e.g., "registry-1.docker.io"
	Repository string // e.g., "library/nginx"
	Tag        string // e.g., "latest"
}

// ImageManifest describes the layers and config of an OCI image.
type ImageManifest struct {
	SchemaVersion int
	MediaType     string
	Config        Descriptor
	Layers        []Descriptor
}

// Descriptor is a content-addressable reference to a blob.
type Descriptor struct {
	MediaType string
	Digest    string
	Size      int64
}

// ImageConfig holds runtime defaults parsed from an OCI image config blob.
type ImageConfig struct {
	User         string
	Env          []string
	Entrypoint   []string
	Cmd          []string
	WorkingDir   string
	ExposedPorts map[string]struct{}
	Volumes      map[string]struct{}
}

// Image is a fully resolved local image with manifest, config, and layer paths.
type Image struct {
	Reference  ImageReference
	Manifest   ImageManifest
	Config     ImageConfig
	LayerPaths []string // Ordered bottom to top.
}

// CgroupStats holds live resource usage metrics read from cgroup files.
type CgroupStats struct {
	CPUUsageUsec   uint64  // Total CPU time consumed in microseconds.
	CPUPercent     float64 // Calculated CPU usage percentage.
	MemoryCurrent  int64   // Current memory usage in bytes.
	MemoryLimit    int64   // Configured memory limit in bytes.
	PidsCurrent    int64   // Current number of processes.
	PidsLimit      int64   // Configured PID limit.
	IOReadBytes    uint64
	IOWriteBytes   uint64
	OOMKillCount   int64
}

// IPAllocation tracks an IP address assigned to a container.
type IPAllocation struct {
	ContainerID string
	IP          net.IP
}
