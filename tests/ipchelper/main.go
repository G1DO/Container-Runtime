// Package main provides a tiny helper used by the IPC integration tests.
// It creates a SysV shared memory segment inside the container and then sleeps
// briefly so the host test can inspect /proc/sysvipc/shm while the segment exists.
package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"
)

const (
	ipcCreat = 0o1000
	ipcExcl  = 0o2000
)

func shmget(key, size, flags int) (int, error) {
	r0, _, errno := syscall.Syscall(syscall.SYS_SHMGET, uintptr(key), uintptr(size), uintptr(flags))
	if errno != 0 {
		return 0, errno
	}
	return int(r0), nil
}

func main() {
	if len(os.Args) != 3 || os.Args[1] != "create" {
		fmt.Fprintln(os.Stderr, "usage: ipc-helper create <key>")
		os.Exit(2)
	}

	key64, err := strconv.ParseInt(os.Args[2], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse key: %v\n", err)
		os.Exit(2)
	}
	key := int(key64)

	shmid, err := shmget(key, 4096, ipcCreat|ipcExcl|0o600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "shmget: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("shm-created key=%d shmid=%d\n", key, shmid)
	time.Sleep(3 * time.Second)
}
