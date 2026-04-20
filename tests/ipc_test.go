// Package tests — ipc_test.go verifies IPC namespace isolation.
// The container should be able to create SysV IPC objects without exposing them
// to the host IPC namespace.
//
// IMPORTANT: Requires root privileges and a Linux kernel with IPC namespace support.
// Run with: sudo go test -v ./tests/ -tags=integration
//
// Milestones: M1.4 (IPC namespace)
package tests

import (
	"bufio"
	"bytes"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func ipcHelperBinary(t *testing.T) string {
	t.Helper()

	bin := t.TempDir() + "/ipc-helper"
	cmd := exec.Command("go", "build", "-o", bin, "./tests/ipchelper")
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
		"GOCACHE="+cacheDir,
		"GOMODCACHE="+modCacheDir,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build ipc helper failed: %v\n%s", err, out)
	}
	return bin
}

func copyFile(t *testing.T, src, dst string, mode os.FileMode) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read helper binary: %v", err)
	}
	if err := os.WriteFile(dst, data, mode); err != nil {
		t.Fatalf("write helper binary: %v", err)
	}
}

func uniqueIPCKey(t *testing.T) int {
	t.Helper()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 64; i++ {
		// Stay in the positive 31-bit range to avoid sign/format surprises.
		key := int(r.Int31n(1<<30) + 1)
		if !hostSharedMemoryHasKey(t, key) {
			return key
		}
	}
	t.Fatal("could not find a free SysV IPC key")
	return 0
}

func hostSharedMemoryHasKey(t *testing.T, key int) bool {
	t.Helper()

	data, err := os.ReadFile("/proc/sysvipc/shm")
	if err != nil {
		t.Fatalf("read /proc/sysvipc/shm: %v", err)
	}

	want := strconv.Itoa(key)
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "key" {
			continue
		}
		if fields[0] == want {
			return true
		}
	}
	return false
}

func waitForLine(t *testing.T, r io.Reader) string {
	t.Helper()

	lineCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(r)
		if scanner.Scan() {
			lineCh <- scanner.Text()
			return
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
			return
		}
		errCh <- io.EOF
	}()

	select {
	case line := <-lineCh:
		return line
	case err := <-errCh:
		t.Fatalf("waiting for helper output: %v", err)
		return ""
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for helper output")
		return ""
	}
}

func startIPCContainer(t *testing.T, bin, rootfs string, key int) (*exec.Cmd, io.ReadCloser, *bytes.Buffer) {
	t.Helper()

	cmd := exec.Command(bin, "run", "--rootfs", rootfs, "/tmp/ipc-helper", "create", strconv.Itoa(key))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start container: %v\n%s", err, stderr.String())
	}
	return cmd, stdout, &stderr
}

// TestIPCSharedMemoryIsolation verifies that a SysV shared memory segment created
// inside the container is not visible from the host IPC namespace.
func TestIPCSharedMemoryIsolation(t *testing.T) {
	needsRoot(t)
	bin := runtimeBinary(t)
	rootfs := testRootFS(t)
	helper := ipcHelperBinary(t)

	tmpDir := filepath.Join(rootfs, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp in rootfs: %v", err)
	}
	helperPath := filepath.Join(tmpDir, "ipc-helper")
	copyFile(t, helper, helperPath, 0o755)
	t.Cleanup(func() { _ = os.Remove(helperPath) })

	key := uniqueIPCKey(t)
	if hostSharedMemoryHasKey(t, key) {
		t.Fatalf("test key unexpectedly already present on host: %d", key)
	}

	cmd, stdout, stderr := startIPCContainer(t, bin, rootfs, key)
	line := waitForLine(t, stdout)
	if !strings.Contains(line, "shm-created") {
		t.Fatalf("unexpected helper output: %q", line)
	}

	if hostSharedMemoryHasKey(t, key) {
		t.Fatalf("SysV shared memory key %d leaked to host IPC namespace", key)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("container run failed: %v\n%s", err, stderr.String())
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for IPC helper to exit")
	}

	if hostSharedMemoryHasKey(t, key) {
		t.Fatalf("SysV shared memory key %d still visible on host after container exit", key)
	}
}
