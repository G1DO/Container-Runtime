// Package main is the CLI entry point for myruntime.
// It parses commands and flags, then delegates to the container.Runtime.
//
// Usage:
//
//	myruntime pull <image>                  Pull an image from a registry
//	myruntime images                        List local images
//	myruntime create <image> [command]      Create a container
//	myruntime start <container-id>          Start a created container
//	myruntime run <image> [command]         Create + Start (shortcut)
//	myruntime exec <container-id> <cmd>     Run command in running container
//	myruntime stop <container-id>           Gracefully stop a container
//	myruntime kill <container-id> [signal]  Send signal to container
//	myruntime rm <container-id>             Delete a stopped container
//	myruntime ps                            List containers
//	myruntime logs <container-id>           Show container stdout/stderr
//	myruntime inspect <container-id>        Show detailed container info
//	myruntime stats <container-id>          Show live resource usage
//
// Milestones: M7.1 (core commands), M7.2 (logging), M7.3 (stats), M7.4 (inspect)
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/G1DO/Container-Runtime/internal/namespace"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "myruntime: a container runtime built from scratch")
		fmt.Fprintln(os.Stderr, "usage: myruntime <command> [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		// Child side: we're inside the new namespaces.
		// os.Args looks like: ["myruntime", "init", "/bin/sh", ...]
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "myruntime init: missing command")
			os.Exit(1)
		}
		cmd := os.Args[2]
		args := os.Args[2:]
		if err := syscall.Exec(cmd, args, os.Environ()); err != nil {
			fmt.Fprintf(os.Stderr, "myruntime init: exec %s: %v\n", cmd, err)
			os.Exit(1)
		}

	case "run":
		// Parent side: fork ourselves into new namespaces.
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "myruntime run: missing command")
			os.Exit(1)
		}
		cmd := exec.Command("/proc/self/exe", append([]string{"init"}, os.Args[2:]...)...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: namespace.CloneFlags(nil),
		}
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "myruntime run: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "myruntime: command %q not yet implemented\n", os.Args[1])
		os.Exit(1)
	}
}
