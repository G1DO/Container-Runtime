# myruntime — Container Runtime from Scratch

A container runtime built from scratch in Go. Not a Docker wrapper — this implements the core Linux primitives that make "containers" possible.

## What a Container Actually Is

A container is a regular Linux process with three restrictions:

```
1. It can't SEE things     →  Namespaces (own PID space, network, filesystem, hostname)
2. It can't USE too much   →  Cgroups (CPU, memory, IO, process limits)
3. It has its own files    →  OverlayFS (layered filesystem with copy-on-write)
```

There's no hypervisor. No hardware emulation. Just a process the kernel has isolated.

```
"Virtual Machine"                     "Container"

┌──────────────────┐                  ┌──────────────────┐
│   Guest App      │                  │   App Process     │
│   Guest Kernel   │                  │   (same kernel!)  │
│   Hypervisor     │                  │   Namespaces      │
│   Host Kernel    │                  │   Cgroups         │
│   Hardware       │                  │   Host Kernel     │
└──────────────────┘                  │   Hardware        │
                                      └──────────────────┘
VM: Emulates hardware.               Container: Restricts a process.
```

## Architecture

```
                    ┌─────────────────────────────────┐
                    │            CLI (cmd/)             │
                    │   run / exec / stop / ps / pull   │
                    └──────────────┬──────────────────┘
                                   │
                    ┌──────────────▼──────────────────┐
                    │     Container Runtime             │
                    │     (internal/container/)         │
                    │                                   │
                    │  Create → Start → Stop → Delete   │
                    └──────────────┬──────────────────┘
                                   │
         ┌─────────────┬───────────┼───────────┬──────────────┐
         ▼             ▼           ▼           ▼              ▼
  ┌────────────┐ ┌──────────┐ ┌────────┐ ┌─────────┐ ┌────────────┐
  │ Namespace  │ │  Cgroup  │ │Filesys │ │  Image  │ │  Network   │
  │            │ │          │ │        │ │         │ │            │
  │ PID        │ │ CPU      │ │Overlay │ │Registry │ │ Bridge     │
  │ NET        │ │ Memory   │ │pivot_  │ │Unpack   │ │ veth       │
  │ MNT        │ │ PIDs     │ │  root  │ │Store    │ │ IPAM       │
  │ UTS        │ │ IO       │ │/dev    │ │         │ │ NAT        │
  │ IPC        │ │          │ │        │ │         │ │ Ports      │
  │ USER       │ │          │ │        │ │         │ │            │
  └────────────┘ └──────────┘ └────────┘ └─────────┘ └────────────┘
                                   │
                    ┌──────────────▼──────────────────┐
                    │     Container Process (PID 1)    │
                    │     Isolated, resource-limited    │
                    └─────────────────────────────────┘
```

## Build

```bash
# Check your environment first
make check

# Build the binary
make build

# Binary is at bin/myruntime
./bin/myruntime
```

## Test

```bash
# Setup test environment (downloads Alpine rootfs, creates dirs)
./scripts/setup-test-env.sh

# Unit tests
make test

# Integration tests (requires root)
make test-integration
```

## Requirements

- **Linux** — kernel 5.8+ (namespaces, cgroups v2, overlayfs are Linux-only)
- **cgroup v2** unified hierarchy mounted at `/sys/fs/cgroup`
- **Go 1.21+**
- **Tools:** iptables, iproute2, bridge-utils, curl
- **Root privileges** for namespace, cgroup, and mount operations

If you're on macOS or Windows, use the provided Vagrantfile or Dockerfile.dev:

```bash
# Option 1: Vagrant VM
vagrant up && vagrant ssh

# Option 2: Docker dev container (requires --privileged)
docker build -f Dockerfile.dev -t myruntime-dev .
docker run --privileged -it -v $(pwd):/workspace myruntime-dev
```

## Docs

- [design.md](docs/design.md) — problem, goals & non-goals, an honest **implemented-vs-scaffolded status table**, failure modes, security stance, and trade-offs
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) — intended component breakdown, dependency graph, data flows
- [decisions/](docs/decisions/) — ADRs, each naming the rejected alternative:
  - [ADR-001](docs/decisions/ADR-001-reexec-init.md) — re-exec `/proc/self/exe` vs. direct `clone()`
  - [ADR-002](docs/decisions/ADR-002-cgroup-v2-only.md) — cgroup v2-only via direct sysfs writes
  - [ADR-003](docs/decisions/ADR-003-pivot-root-vs-chroot.md) — `pivot_root` vs. `chroot`
  - [ADR-004](docs/decisions/ADR-004-raw-rtnetlink.md) — hand-rolled rtnetlink vs. a netlink library / `ip(8)`
  - [ADR-005](docs/decisions/ADR-005-json-state-files.md) — JSON state files vs. an embedded database
  - [ADR-006](docs/decisions/ADR-006-all-namespaces-combined.md) — all six namespaces combined vs. configurable
- [threat-model.md](docs/threat-model.md) — STRIDE-lite analysis with the trust-boundary diagram
- [diagrams/](docs/diagrams/) — Mermaid sources + rendered SVGs (CI-gated for drift)
