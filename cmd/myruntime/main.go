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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/G1DO/Container-Runtime/internal/filesystem"
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
		// os.Args looks like:
		// ["myruntime", "init", "--rootfs", "/abs/rootfs", "--hostname", "name", "--", "/bin/sh", ...]
		initFlags := flag.NewFlagSet("init", flag.ContinueOnError)
		initFlags.SetOutput(os.Stderr)
		rootfs := initFlags.String("rootfs", "", "path to container root filesystem")
		hostname := initFlags.String("hostname", "", "container hostname")
		if err := initFlags.Parse(os.Args[2:]); err != nil {
			os.Exit(2)
		}
		args := initFlags.Args()
		if len(args) > 0 && args[0] == "--" {
			args = args[1:]
		}
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "myruntime init: missing command")
			os.Exit(1)
		}

		if *rootfs == "" {
			fmt.Fprintln(os.Stderr, "myruntime init: missing --rootfs")
			os.Exit(1)
		}
		absRootfs, err := filepath.Abs(*rootfs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "myruntime init: abs rootfs: %v\n", err)
			os.Exit(1)
		}
		if err := filesystem.SetupContainerMounts(absRootfs); err != nil {
			fmt.Fprintf(os.Stderr, "myruntime init: setup mounts: %v\n", err)
			os.Exit(1)
		}
		if *hostname != "" {
			if err := namespace.SetupHostname(*hostname); err != nil {
				fmt.Fprintf(os.Stderr, "myruntime init: setup hostname: %v\n", err)
				os.Exit(1)
			}
		}
		if err := namespace.SetupLoopback(); err != nil {
			fmt.Fprintf(os.Stderr, "myruntime init: setup loopback: %v\n", err)
			os.Exit(1)
		}

		cmd := args[0]
		if err := syscall.Exec(cmd, args, os.Environ()); err != nil {
			fmt.Fprintf(os.Stderr, "myruntime init: exec %s: %v\n", cmd, err)
			os.Exit(1)
		}

	case "run":
		// Parent side: fork ourselves into new namespaces.
		runFlags := flag.NewFlagSet("run", flag.ContinueOnError)
		runFlags.SetOutput(os.Stderr)
		rootfs := runFlags.String("rootfs", "", "path to container root filesystem")
		hostname := runFlags.String("hostname", "", "container hostname")
		if err := runFlags.Parse(os.Args[2:]); err != nil {
			os.Exit(2)
		}
		args := runFlags.Args()
		if len(args) > 0 && args[0] == "--" {
			args = args[1:]
		}
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "myruntime run: missing command")
			os.Exit(1)
		}
		if *rootfs == "" {
			fmt.Fprintln(os.Stderr, "myruntime run: missing --rootfs")
			os.Exit(1)
		}
		absRootfs, err := filepath.Abs(*rootfs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "myruntime run: abs rootfs: %v\n", err)
			os.Exit(1)
		}

		cloneFlags := namespace.CloneFlags(nil)

		initArgs := []string{"init", "--rootfs", absRootfs}
		if *hostname != "" {
			initArgs = append(initArgs, "--hostname", *hostname)
		}
		initArgs = append(initArgs, "--")
		initArgs = append(initArgs, args...)

		cmd := exec.Command("/proc/self/exe", initArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: cloneFlags,
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
