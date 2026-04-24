package namespace

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCloneFlagsPhase1(t *testing.T) {
	got := CloneFlags(nil)
	want := uintptr(syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET)
	if got != want {
		t.Fatalf("CloneFlags(nil) = %#x, want %#x", got, want)
	}
}

func TestDefaultIDMappings(t *testing.T) {
	want := syscall.SysProcIDMap{
		ContainerID: 0,
		HostID:      100000,
		Size:        65536,
	}

	uidMappings := DefaultUIDMappings()
	if len(uidMappings) != 1 || uidMappings[0] != want {
		t.Fatalf("DefaultUIDMappings() = %#v, want %#v", uidMappings, []syscall.SysProcIDMap{want})
	}

	gidMappings := DefaultGIDMappings()
	if len(gidMappings) != 1 || gidMappings[0] != want {
		t.Fatalf("DefaultGIDMappings() = %#v, want %#v", gidMappings, []syscall.SysProcIDMap{want})
	}
}

func TestWriteIDMappingsFormatsProcFiles(t *testing.T) {
	procPidDir := t.TempDir()
	uidMappings := []syscall.SysProcIDMap{
		{ContainerID: 0, HostID: 100000, Size: 65536},
		{ContainerID: 65536, HostID: 200000, Size: 1},
	}
	gidMappings := []syscall.SysProcIDMap{
		{ContainerID: 0, HostID: 100000, Size: 65536},
	}

	if err := writeIDMappings(procPidDir, uidMappings, gidMappings); err != nil {
		t.Fatalf("writeIDMappings() error = %v", err)
	}

	assertFileContent(t, filepath.Join(procPidDir, "uid_map"), "0 100000 65536\n65536 200000 1\n")
	assertFileContent(t, filepath.Join(procPidDir, "setgroups"), "deny\n")
	assertFileContent(t, filepath.Join(procPidDir, "gid_map"), "0 100000 65536\n")
}

func TestWriteIDMappingsRejectsInvalidMappings(t *testing.T) {
	err := WriteIDMappings(0, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid pid") {
		t.Fatalf("WriteIDMappings(0, nil, nil) error = %v, want invalid pid", err)
	}

	err = writeIDMappings(t.TempDir(), []syscall.SysProcIDMap{{
		ContainerID: 0,
		HostID:      100000,
		Size:        0,
	}}, nil)
	if err == nil || !strings.Contains(err.Error(), "non-positive size") {
		t.Fatalf("writeIDMappings() error = %v, want non-positive size", err)
	}
}

func TestUserNamespaceMapsContainerRootToHostUser(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires linux")
	}
	if os.Getuid() == 0 {
		t.Skip("requires an unprivileged host user so container root maps away from host root")
	}

	probe, err := runUserNamespaceProbe(os.Getuid(), os.Getgid())
	if err != nil {
		if isNamespacePermissionError(err) {
			t.Skipf("requires unprivileged user namespaces: %v", err)
		}
		t.Fatalf("run user namespace probe: %v", err)
	}

	if probe.InsideUID != 0 {
		t.Fatalf("inside user namespace uid = %d, want 0", probe.InsideUID)
	}
	if probe.InsideGID != 0 {
		t.Fatalf("inside user namespace gid = %d, want 0", probe.InsideGID)
	}
	assertAllIDs(t, "host Uid", probe.HostUIDs, os.Getuid())
	assertAllIDs(t, "host Gid", probe.HostGIDs, os.Getgid())
}

func TestSetupLoopbackInFreshNetworkNamespace(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("requires linux")
	}

	probe, err := runLoopbackProbe()
	if err != nil {
		if isNamespacePermissionError(err) {
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

func TestUserNamespaceHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_USERNS_HELPER") != "1" {
		return
	}

	reader := bufio.NewReader(os.Stdin)
	if _, err := reader.ReadString('\n'); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	probe := userNamespaceProbe{
		InsideUID: os.Getuid(),
		InsideGID: os.Getgid(),
	}
	if err := json.NewEncoder(os.Stdout).Encode(probe); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, _ = io.Copy(io.Discard, reader)
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

type userNamespaceProbe struct {
	InsideUID int   `json:"insideUid"`
	InsideGID int   `json:"insideGid"`
	HostUIDs  []int `json:"hostUids,omitempty"`
	HostGIDs  []int `json:"hostGids,omitempty"`
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

func runUserNamespaceProbe(hostUID, hostGID int) (userNamespaceProbe, error) {
	cmd := exec.Command(os.Args[0], "-test.run=TestUserNamespaceHelperProcess$")
	cmd.Env = append(os.Environ(), "GO_WANT_USERNS_HELPER=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER,
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return userNamespaceProbe{}, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return userNamespaceProbe{}, err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return userNamespaceProbe{}, err
	}

	cleanup := func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}

	uidMappings := []syscall.SysProcIDMap{{
		ContainerID: 0,
		HostID:      hostUID,
		Size:        1,
	}}
	gidMappings := []syscall.SysProcIDMap{{
		ContainerID: 0,
		HostID:      hostGID,
		Size:        1,
	}}
	if err := WriteIDMappings(cmd.Process.Pid, uidMappings, gidMappings); err != nil {
		cleanup()
		return userNamespaceProbe{}, err
	}

	if _, err := io.WriteString(stdin, "ready\n"); err != nil {
		cleanup()
		return userNamespaceProbe{}, err
	}

	var probe userNamespaceProbe
	if err := json.NewDecoder(stdout).Decode(&probe); err != nil {
		cleanup()
		return userNamespaceProbe{}, fmt.Errorf("decode helper output: %w: %s", err, stderr.String())
	}

	hostUIDs, err := readProcStatusIDs(cmd.Process.Pid, "Uid")
	if err != nil {
		cleanup()
		return userNamespaceProbe{}, err
	}
	hostGIDs, err := readProcStatusIDs(cmd.Process.Pid, "Gid")
	if err != nil {
		cleanup()
		return userNamespaceProbe{}, err
	}
	probe.HostUIDs = hostUIDs
	probe.HostGIDs = hostGIDs

	if err := stdin.Close(); err != nil {
		cleanup()
		return userNamespaceProbe{}, err
	}
	if err := cmd.Wait(); err != nil {
		return userNamespaceProbe{}, fmt.Errorf("%w: %s", err, stderr.String())
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

func readProcStatusIDs(pid int, name string) ([]int, error) {
	status, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return nil, err
	}

	prefix := name + ":"
	for _, line := range strings.Split(string(status), "\n") {
		if !strings.HasPrefix(line, prefix) {
			continue
		}

		fields := strings.Fields(strings.TrimPrefix(line, prefix))
		ids := make([]int, 0, len(fields))
		for _, field := range fields {
			id, err := strconv.Atoi(field)
			if err != nil {
				return nil, fmt.Errorf("parse %s field %q: %w", name, field, err)
			}
			ids = append(ids, id)
		}
		return ids, nil
	}
	return nil, fmt.Errorf("%s line not found in /proc/%d/status", name, pid)
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("%s = %q, want %q", path, got, want)
	}
}

func assertAllIDs(t *testing.T, name string, got []int, want int) {
	t.Helper()
	if len(got) == 0 {
		t.Fatalf("%s = %v, want at least one %d", name, got, want)
	}
	for _, id := range got {
		if id != want {
			t.Fatalf("%s = %v, want all IDs to be %d", name, got, want)
		}
	}
}

func isNamespacePermissionError(err error) bool {
	return errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES)
}
