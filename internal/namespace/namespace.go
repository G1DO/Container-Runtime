// Package namespace configures Linux namespace isolation for containers.
// Namespaces are the core mechanism that makes a process "not see" the host —
// its own PID space, network stack, mount tree, hostname, IPC, and user IDs.
//
// Milestones: M1.1 (PID), M1.2 (MNT), M1.3 (UTS), M1.4 (IPC), M1.5 (NET), M1.6 (USER)
package namespace

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	"github.com/G1DO/Container-Runtime/pkg/specs"
)

// CloneFlags returns the namespace flags used in phase 1.
// Phase 1 enables PID, mount, UTS, IPC, network, and user namespaces.
func CloneFlags(config *specs.ContainerConfig) uintptr {
	_ = config
	return syscall.CLONE_NEWUSER | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET
}

// SetupHostname sets the container's hostname via the UTS namespace.
// Requires CLONE_NEWUTS to have been set when creating the process.
func SetupHostname(hostname string) error {
	if hostname == "" {
		return nil
	}
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return fmt.Errorf("set hostname %q: %w", hostname, err)
	}
	return nil
}

// SetupLoopback brings up the loopback (lo) interface inside a new network namespace.
// A new NET namespace starts with zero interfaces — not even loopback.
func SetupLoopback() error {
	lo, err := net.InterfaceByName("lo")
	if err != nil {
		return fmt.Errorf("find loopback interface: %w", err)
	}
	if lo.Flags&net.FlagUp != 0 {
		return nil
	}
	if err := linkSetUp(lo.Index); err != nil {
		return fmt.Errorf("bring up loopback interface: %w", err)
	}
	return nil
}

// JoinNamespaces enters the namespaces of an existing container process.
// Used by "exec" to run a new command inside a running container.
// Opens /proc/<pid>/ns/<type> and calls setns(2) for each namespace.
func JoinNamespaces(pid int) error {
	// TODO(M5.4): Open /proc/<pid>/ns/{pid,net,mnt,uts,ipc,user} and setns into each
	return nil
}

// linkSetUp sends a minimal RTM_NEWLINK request that marks one interface up.
// Later networking milestones will build on the same rtnetlink machinery for
// bridges, veth pairs, routes, and addresses.
func linkSetUp(index int) error {
	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_ROUTE)
	if err != nil {
		return fmt.Errorf("open rtnetlink socket: %w", err)
	}
	defer syscall.Close(fd)

	if err := syscall.Bind(fd, &syscall.SockaddrNetlink{Family: syscall.AF_NETLINK}); err != nil {
		return fmt.Errorf("bind rtnetlink socket: %w", err)
	}

	req := struct {
		Header syscall.NlMsghdr
		Info   syscall.IfInfomsg
	}{
		Header: syscall.NlMsghdr{
			Len:   uint32(syscall.SizeofNlMsghdr + syscall.SizeofIfInfomsg),
			Type:  syscall.RTM_NEWLINK,
			Flags: syscall.NLM_F_REQUEST | syscall.NLM_F_ACK,
			Seq:   1,
		},
		Info: syscall.IfInfomsg{
			Family: syscall.AF_UNSPEC,
			Index:  int32(index),
			Flags:  syscall.IFF_UP,
			Change: syscall.IFF_UP,
		},
	}

	msg := make([]byte, 0, req.Header.Len)
	msg = appendStructBytes(msg, &req.Header)
	msg = appendStructBytes(msg, &req.Info)

	if err := syscall.Sendto(fd, msg, 0, &syscall.SockaddrNetlink{Family: syscall.AF_NETLINK}); err != nil {
		return fmt.Errorf("send RTM_NEWLINK: %w", err)
	}

	if err := readNetlinkAck(fd, req.Header.Seq); err != nil {
		return fmt.Errorf("read RTM_NEWLINK ack: %w", err)
	}

	return nil
}

func appendStructBytes[T any](dst []byte, value *T) []byte {
	size := int(unsafe.Sizeof(*value))
	return append(dst, unsafe.Slice((*byte)(unsafe.Pointer(value)), size)...)
}

func readNetlinkAck(fd int, seq uint32) error {
	buf := make([]byte, os.Getpagesize())

	for {
		n, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return err
		}

		msgs, err := syscall.ParseNetlinkMessage(buf[:n])
		if err != nil {
			return fmt.Errorf("parse netlink message: %w", err)
		}

		for _, msg := range msgs {
			if msg.Header.Seq != seq {
				continue
			}

			switch msg.Header.Type {
			case syscall.NLMSG_DONE:
				return nil
			case syscall.NLMSG_ERROR:
				if len(msg.Data) < 4 {
					return fmt.Errorf("short netlink error payload")
				}

				errno := *(*int32)(unsafe.Pointer(&msg.Data[0]))
				if errno == 0 {
					return nil
				}

				return syscall.Errno(-errno)
			}
		}
	}
}
