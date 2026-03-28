# Architecture

## Component Dependency Graph

```
cmd/myruntime/main.go
    │
    └──► internal/container/
            │
            ├──► internal/namespace/      (no deps)
            ├──► internal/cgroup/         (depends on: pkg/specs)
            ├──► internal/filesystem/     (no deps)
            ├──► internal/image/          (depends on: pkg/specs)
            ├──► internal/network/        (depends on: pkg/specs)
            └──► internal/store/          (depends on: pkg/specs)

    All packages import pkg/specs/ for shared types.
    No circular dependencies. Dependency flows downward only.
```

## Directory Structure

```
myruntime/
├── cmd/
│   └── myruntime/
│       └── main.go                 # CLI entry point — parses commands, delegates to Runtime
│
├── internal/
│   ├── container/
│   │   ├── container.go            # Runtime struct — ties all subsystems together
│   │   ├── lifecycle.go            # Create, Start, Stop, Kill, Delete operations
│   │   ├── init.go                 # Container init process — runs INSIDE the namespaces
│   │   └── exec.go                 # Exec into running container + crash recovery
│   │
│   ├── namespace/
│   │   ├── namespace.go            # Clone flags, hostname setup, loopback, namespace joining
│   │   └── userns.go               # User namespace UID/GID mapping
│   │
│   ├── cgroup/
│   │   ├── manager.go              # Create/destroy cgroups, apply CPU/mem/PID/IO limits
│   │   └── stats.go                # Read cgroup metrics (cpu.stat, memory.current, etc.)
│   │
│   ├── filesystem/
│   │   ├── overlay.go              # OverlayFS mount/unmount, layer store
│   │   ├── devices.go              # /dev setup — device nodes, /dev/pts, /dev/shm
│   │   └── mounts.go               # pivot_root, /proc, /sys, /tmp mounts
│   │
│   ├── image/
│   │   ├── registry.go             # OCI registry client — auth, manifest, blob download
│   │   ├── store.go                # Local image storage and config parsing
│   │   └── unpack.go               # Layer extraction (tar.gz → directory)
│   │
│   ├── network/
│   │   ├── bridge.go               # Virtual bridge creation, NAT setup
│   │   ├── veth.go                 # veth pair creation, container network config, DNS
│   │   ├── ipam.go                 # IP address allocation from subnet
│   │   └── iptables.go             # Port mapping via DNAT rules
│   │
│   └── store/
│       └── state.go                # Container state persistence (JSON on disk)
│
├── pkg/
│   └── specs/
│       └── config.go               # Shared types: Container, Config, Image, CgroupStats
│
├── tests/
│   ├── isolation_test.go           # Namespace isolation verification
│   ├── cgroup_test.go              # Resource limit enforcement
│   ├── network_test.go             # Networking (bridge, veth, NAT, ports)
│   └── stress_test.go              # Many containers, fork bombs, disk fill
│
├── scripts/
│   └── setup-test-env.sh           # Test environment setup (rootfs, cgroups, dirs)
│
├── go.mod
├── Makefile
├── Vagrantfile                     # Linux dev VM (for macOS/Windows)
├── Dockerfile.dev                  # Linux dev container
├── README.md
├── ARCHITECTURE.md                 # This file
├── MILESTONES.md                   # Learning roadmap
└── JOURNAL.md                      # Learning journal
```

## Interface Boundaries

### What each package exposes vs. hides

| Package | Exports (public API) | Hides (internal) |
|---------|---------------------|-------------------|
| `pkg/specs` | All types: Container, ContainerConfig, ResourceConfig, ImageReference, Image, CgroupStats | Nothing — this is the shared contract |
| `internal/container` | `Runtime`, `NewRuntime`, lifecycle methods (Create/Start/Stop/Kill/Delete/Exec/List) | Init process internals, signal forwarding, zombie reaping |
| `internal/namespace` | `CloneFlags`, `SetupHostname`, `SetupLoopback`, `JoinNamespaces` | Raw syscall details |
| `internal/cgroup` | `Manager`, `NewManager`, Create/Destroy/AddProcess/Stats | File path construction, stat file parsing |
| `internal/filesystem` | `OverlayFS`, `LayerStore`, `CreateDevices`, `PivotRoot`, `SetupContainerMounts` | Mount option string building, device major/minor numbers |
| `internal/image` | `RegistryClient`, `Store`, `ParseImageConfig` | HTTP auth flow, tar extraction details |
| `internal/network` | `CreateBridge`, `CreateVethPair`, `IPAM`, `AddPortMapping`, `SetupDNS` | iptables command construction, netlink details |
| `internal/store` | `ContainerStore`, Save/Load/List/Delete | JSON serialization, file locking |

## Data Flow Diagrams

### `myruntime run nginx:latest`

```
User runs: myruntime run nginx:latest
    │
    ▼
CLI parses flags, calls Runtime.Create(config)
    │
    ├─► ImageStore.Resolve("nginx:latest")
    │       → returns layer paths + image config
    │
    ├─► LayerStore.PrepareContainer(id, layers)
    │       → creates overlay dirs
    │   OverlayFS.Mount()
    │       → mount -t overlay ... mergedDir
    │
    ├─► CgroupManager.Create(id, resources)
    │       → mkdir /sys/fs/cgroup/myruntime/<id>
    │       → write cpu.max, memory.max, pids.max
    │
    ├─► ContainerStore.Save(container{state: created})
    │
    ▼
CLI calls Runtime.Start(id)
    │
    ├─► Fork child process via re-exec
    │       cmd = /proc/self/exe "container-init" <id>
    │       SysProcAttr.Cloneflags = CLONE_NEWPID | CLONE_NEWNS | ...
    │       cmd.Start()
    │
    ├─► CgroupManager.AddProcess(id, child.Pid)
    │       → write PID to cgroup.procs
    │
    ├─► Network: CreateVethPair, IPAM.Allocate, ConfigureContainerNetwork
    │
    ├─► Signal child to proceed (via pipe)
    │
    │   ┌─── Child process (new namespaces) ───┐
    │   │                                        │
    │   │  WaitForReady() ◄── pipe signal        │
    │   │  PivotRoot(rootfs)                     │
    │   │  MountProc, MountSys                   │
    │   │  CreateDevices                         │
    │   │  SetHostname                           │
    │   │  SetEnv                                │
    │   │  Chdir(workingDir)                     │
    │   │  syscall.Exec(entrypoint) ←── PID 1   │
    │   │                                        │
    │   └────────────────────────────────────────┘
    │
    └─► ContainerStore.Save(container{state: running, pid: child.Pid})
```

### `myruntime exec <id> /bin/sh`

```
User runs: myruntime exec abc123 /bin/sh
    │
    ▼
CLI calls Runtime.Exec(id, ["/bin/sh"])
    │
    ├─► ContainerStore.Load(id) → get container.Pid
    │
    ├─► For each ns in {pid, net, mnt, uts, ipc, user}:
    │       fd = open("/proc/<pid>/ns/<ns>")
    │       setns(fd, 0)    ← enters the existing namespace
    │       close(fd)
    │
    └─► syscall.Exec("/bin/sh", [...], env)
            → now running INSIDE the container's namespaces
            → shares PID space, network, filesystem with the container
```

### `myruntime stop <id>`

```
User runs: myruntime stop abc123
    │
    ▼
CLI calls Runtime.Stop(id, 10s)
    │
    ├─► ContainerStore.Load(id) → get container.Pid
    │
    ├─► syscall.Kill(pid, SIGTERM)     ← "please shut down"
    │
    ├─► Wait up to 10 seconds
    │       ├── Process exits → capture exit code
    │       └── Timeout → syscall.Kill(pid, SIGKILL)  ← force kill
    │
    ├─► ContainerStore.Save(container{state: stopped, exitCode: N})
    │
    └─► (resources NOT cleaned up yet — user must call "rm" to delete)
```

### Container networking setup

```
During Runtime.Start():
    │
    ├─► CreateBridge("myruntime0", "172.20.0.1/24")     [once, on first container]
    │       ├── netlink: create bridge interface
    │       ├── netlink: assign 172.20.0.1/24 to bridge
    │       ├── netlink: bring bridge up
    │       ├── echo 1 > /proc/sys/net/ipv4/ip_forward
    │       └── iptables -t nat -A POSTROUTING -s 172.20.0.0/24 -j MASQUERADE
    │
    ├─► IPAM.Allocate(containerID) → 172.20.0.2
    │
    ├─► CreateVethPair("veth_abc", "eth0", containerPid, "myruntime0")
    │       ├── netlink: create veth pair (veth_abc ↔ eth0)
    │       ├── netlink: move eth0 into container's net namespace
    │       ├── netlink: attach veth_abc to bridge
    │       └── netlink: bring veth_abc up
    │
    ├─► ConfigureContainerNetwork(pid, "172.20.0.2", "172.20.0.1", "eth0")
    │       ├── [inside netns] ip addr add 172.20.0.2/24 dev eth0
    │       ├── [inside netns] ip link set eth0 up
    │       └── [inside netns] ip route add default via 172.20.0.1
    │
    ├─► SetupDNS(rootfs, ["8.8.8.8", "8.8.4.4"])
    │       └── write /etc/resolv.conf
    │
    └─► AddPortMappings([{8080, 80, "tcp"}], "172.20.0.2")
            └── iptables -t nat -A PREROUTING -p tcp --dport 8080
                    -j DNAT --to-destination 172.20.0.2:80
```

### Crash recovery (on startup)

```
Runtime.NewRuntime() calls ReconcileOnStartup():
    │
    ├─► ContainerStore.List() → all containers
    │
    ├─► For each container where state == "running":
    │       │
    │       ├── syscall.Kill(container.Pid, 0)   ← signal 0 = "is this process alive?"
    │       │
    │       ├── If alive: no action needed
    │       │
    │       └── If dead (ESRCH):
    │               ├── CgroupManager.Destroy(id)  ← clean up orphaned cgroup
    │               ├── container.State = stopped
    │               ├── container.ExitCode = -1     ← unknown exit
    │               └── ContainerStore.Save(container)
    │
    └─► Done — state is now consistent
```
