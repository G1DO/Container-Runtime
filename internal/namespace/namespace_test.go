package namespace

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestCloneFlagsPhase1(t *testing.T) {
	got := CloneFlags(nil)
	want := uintptr(syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET)
	if got != want {
		t.Fatalf("CloneFlags(nil) = %#x, want %#x", got, want)
	}
}

func TestSetupLoopbackInFreshNetworkNamespace(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires linux")
	}

	probe, err := runLoopbackProbe()
	if err != nil {
		if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) {
			t.Skipf("requires unprivileged user and network namespaces: %v", err)
		}
		t.Fatalf("run loopback probe: %v", err)
	}

	if len(probe.Before) != 1 || probe.Before[0].Name != "lo" {
		t.Fatalf("fresh network namespace interfaces = %#v, want only lo", probe.Before)
	}
	if probe.Before[0].Up {
		t.Fatalf("expected loopback to start down in fresh netns, got %#v", probe.Before[0])
	}
	if len(probe.After) != 1 || probe.After[0].Name != "lo" {
		t.Fatalf("after SetupLoopback interfaces = %#v, want only lo", probe.After)
	}
	if !probe.After[0].Up {
		t.Fatalf("expected loopback to be up after SetupLoopback, got %#v", probe.After[0])
	}
	if !probe.LoopbackReachable {
		t.Fatal("expected 127.0.0.1 to be reachable after SetupLoopback")
	}

	hostIfaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("list host interfaces: %v", err)
	}
	for _, hostIface := range hostIfaces {
		if hostIface.Name == "lo" {
			continue
		}
		for _, iface := range probe.After {
			if iface.Name == hostIface.Name {
				t.Fatalf("host interface %q leaked into new netns", hostIface.Name)
			}
		}
	}
}

func TestSetupLoopbackHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_LOOPBACK_HELPER") != "1" {
		return
	}

	probe, err := collectLoopbackProbe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(probe); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

type interfaceProbe struct {
	Name string `json:"name"`
	Up   bool   `json:"up"`
}

type loopbackProbe struct {
	Before            []interfaceProbe `json:"before"`
	After             []interfaceProbe `json:"after"`
	LoopbackReachable bool             `json:"loopbackReachable"`
}

func runLoopbackProbe() (loopbackProbe, error) {
	cmd := exec.Command(os.Args[0], "-test.run=TestSetupLoopbackHelperProcess$")
	cmd.Env = append(os.Environ(), "GO_WANT_LOOPBACK_HELPER=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID:      os.Getuid(),
			Size:        1,
		}},
		GidMappings: []syscall.SysProcIDMap{{
			ContainerID: 0,
			HostID:      os.Getgid(),
			Size:        1,
		}},
		GidMappingsEnableSetgroups: false,
	}

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return loopbackProbe{}, fmt.Errorf("%w: %s", err, exitErr.Stderr)
		}
		return loopbackProbe{}, err
	}

	var probe loopbackProbe
	if err := json.Unmarshal(out, &probe); err != nil {
		return loopbackProbe{}, fmt.Errorf("decode helper output %q: %w", string(out), err)
	}
	return probe, nil
}

func collectLoopbackProbe() (loopbackProbe, error) {
	before, err := snapshotInterfaces()
	if err != nil {
		return loopbackProbe{}, err
	}
	if err := SetupLoopback(); err != nil {
		return loopbackProbe{}, err
	}
	after, err := snapshotInterfaces()
	if err != nil {
		return loopbackProbe{}, err
	}

	return loopbackProbe{
		Before:            before,
		After:             after,
		LoopbackReachable: loopbackReachable(),
	}, nil
}

func snapshotInterfaces() ([]interfaceProbe, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	snapshots := make([]interfaceProbe, 0, len(ifaces))
	for _, iface := range ifaces {
		snapshots = append(snapshots, interfaceProbe{
			Name: iface.Name,
			Up:   iface.Flags&net.FlagUp != 0,
		})
	}
	return snapshots, nil
}

func loopbackReachable() bool {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return false
	}
	defer ln.Close()

	accepted := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			accepted <- err
			return
		}
		_ = conn.Close()
		accepted <- nil
	}()

	conn, err := net.DialTimeout("tcp4", ln.Addr().String(), time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()

	select {
	case err := <-accepted:
		return err == nil
	case <-time.After(time.Second):
		return false
	}
}
