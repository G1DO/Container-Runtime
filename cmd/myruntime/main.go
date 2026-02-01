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
)

func main() {
	// TODO(M7.1): Initialize urfave/cli app with all commands and flags
	// TODO(M7.1): Initialize container.Runtime
	// TODO(M7.1): Run CLI app

	// Placeholder so the binary compiles and runs.
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "myruntime: a container runtime built from scratch")
		fmt.Fprintln(os.Stderr, "usage: myruntime <command> [args...]")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "myruntime: command %q not yet implemented\n", os.Args[1])
	os.Exit(1)
}
