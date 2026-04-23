// Package tests — network_test.go verifies container networking.
//
// Milestones: M9.1 (network isolation), M6.7 (container-to-container)
package tests

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestNetworkIsolation verifies the container has its own network namespace
// and cannot see host interfaces.
func TestNetworkIsolation(t *testing.T) {
	needsRoot(t)

	probe := runContainerNetworkProbe(t)
	if len(probe.Interfaces) != 1 || probe.Interfaces[0].Name != "lo" {
		t.Fatalf("container interfaces = %#v, want only lo", probe.Interfaces)
	}
	if !probe.Interfaces[0].Up {
		t.Fatalf("container loopback interface is down: %#v", probe.Interfaces[0])
	}

	hostIfaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("list host interfaces: %v", err)
	}
	for _, hostIface := range hostIfaces {
		if hostIface.Name == "lo" {
			continue
		}
		for _, iface := range probe.Interfaces {
			if iface.Name == hostIface.Name {
				t.Fatalf("host interface %q visible inside container", hostIface.Name)
			}
		}
	}
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

// TestContainerLoopbackReachability verifies that the runtime brings up lo
// inside the container's otherwise-empty network namespace.
func TestContainerLoopbackReachability(t *testing.T) {
	needsRoot(t)

	probe := runContainerNetworkProbe(t)
	if !probe.LoopbackReachable {
		t.Fatal("expected 127.0.0.1 to be reachable inside the container")
	}
}

type networkProbe struct {
	Interfaces        []networkInterface `json:"interfaces"`
	LoopbackReachable bool               `json:"loopbackReachable"`
}

type networkInterface struct {
	Name string `json:"name"`
	Up   bool   `json:"up"`
}

func networkHelperBinary(t *testing.T) string {
	t.Helper()

	bin := t.TempDir() + "/net-helper"
	cmd := exec.Command("go", "build", "-o", bin, "./tests/nethelper")
	cmd.Dir = ".."
	cacheDir := filepath.Join(t.TempDir(), "gocache")
	modCacheDir := filepath.Join(t.TempDir(), "gomodcache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("mkdir gocache: %v", err)
	}
	if err := os.MkdirAll(modCacheDir, 0o755); err != nil {
		t.Fatalf("mkdir gomodcache: %v", err)
	}
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOCACHE="+cacheDir,
		"GOMODCACHE="+modCacheDir,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build network helper failed: %v\n%s", err, out)
	}
	return bin
}

func runContainerNetworkProbe(t *testing.T) networkProbe {
	t.Helper()

	bin := runtimeBinary(t)
	rootfs := testRootFS(t)
	helper := networkHelperBinary(t)

	tmpDir := filepath.Join(rootfs, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp in rootfs: %v", err)
	}
	helperPath := filepath.Join(tmpDir, "net-helper")
	copyFile(t, helper, helperPath, 0o755)
	t.Cleanup(func() { _ = os.Remove(helperPath) })

	out, err := exec.Command(bin, "run", "--rootfs", rootfs, "/tmp/net-helper").CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}

	var probe networkProbe
	if err := json.Unmarshal(out, &probe); err != nil {
		t.Fatalf("decode network helper output %q: %v", string(out), err)
	}
	if len(probe.Interfaces) == 0 {
		t.Fatalf("network helper returned no interfaces: %s", fmt.Sprintf("%+v", probe))
	}
	return probe
}
