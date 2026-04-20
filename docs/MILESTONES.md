# Development Milestones

29 milestones across 9 phases. Each builds on the previous.

Legend:
- ⚠️ = Conceptually hard — expect to spend extra time studying before coding
- 🎯 = Good demo point — something visibly cool you can show

---

## Phase 1: Process Isolation with Namespaces

### M1.1: Basic Process with PID Namespace (2 days)

**Goal:** Run a child process in a new PID namespace. It sees itself as PID 1.

**Concepts to study BEFORE coding:**
- `man 7 namespaces` — overview of all Linux namespace types
- `man 2 clone` — the syscall that creates processes in new namespaces
- `man 2 unshare` — alternative to clone for entering new namespaces
- Watch: Liz Rice "Containers From Scratch" (35 min) — live-codes this exact step
- Understand the **re-exec pattern**: why Go can't just call clone() directly (goroutines = threads, and namespace ops need to happen before threads exist)

**Deliverable:** A Go binary that forks a child in a new PID namespace. Running `ps` inside shows only the child process.

**Tests to write FIRST:**
- Run `ps aux` inside the container → process count <= 2
- Run `echo $$` inside → prints `1`
- Host `ps` still shows all host processes (not affected)

**Checkpoint questions:**
1. Why can't you just call `clone()` directly from Go? What does the re-exec pattern solve?
2. What happens to the child if PID 1 inside the namespace dies?
3. From the host, can you see the container's process? What PID does it have on the host?

---

### ⚠️ M1.2: Mount Namespace and pivot_root (3 days)

**Goal:** Give the container its own filesystem root. The host's mounts are invisible.

**Concepts to study BEFORE coding:**
- `man 2 pivot_root` — swaps the root filesystem
- `man 2 chroot` — the weaker alternative (understand why it's insufficient)
- `man 7 mount_namespaces` — mount propagation: shared, private, slave, unbindable
- Read: runc's `rootfs_linux.go` — how the reference implementation does pivot_root
- Understand: `MS_REC | MS_PRIVATE` — why you must make mounts private before pivot

**Deliverable:** Container has its own root filesystem. Writes inside don't appear on host.

**Tests to write FIRST:**
- Create a file inside the container at `/tmp/test` → does NOT exist on host `/tmp/test`
- `ls /` inside container shows only the container's rootfs
- After container exits, host mounts are unaffected

**Checkpoint questions:**
1. What's the security difference between `chroot` and `pivot_root`? How can you escape chroot?
2. Why must you make the root mount `MS_PRIVATE` before `pivot_root`?
3. What happens if you forget to unmount the old root after pivot?

**Suggested todo order:**
- [ ] Study: `man 2 pivot_root` — what it does, its preconditions, and how it differs from `chroot`
- [ ] Study: `man 2 chroot` escape techniques — understand why `chroot` is insufficient for container security
- [ ] Study: `man 7 mount_namespaces` — mount propagation types (`shared`, `private`, `slave`, `unbindable`)
- [ ] Study: `MS_REC | MS_PRIVATE` — why the root mount must be made private before `pivot_root`
- [ ] Study: read runc's `rootfs_linux.go` pivot_root implementation in the reference runtime
- [ ] Answer the 3 checkpoint questions above before writing any code
- [ ] Write integration tests first in `tests/mount_test.go`: writes to `/tmp/test` stay inside the container, `ls /` shows only the container rootfs, and host mounts are unchanged after exit
- [ ] Implement `internal/filesystem/mounts.go`: make mounts private, bind-mount the new root, `pivot_root`, `chdir("/")`, unmount the old root, then wire `SetupContainerMounts` into `internal/container/init.go`

---

### M1.3: UTS Namespace — Hostname Isolation (1 day)

**Goal:** Container has its own hostname that doesn't affect the host.

**Concepts to study BEFORE coding:**
- `man 2 sethostname`
- `man 7 uts_namespaces`

**Deliverable:** Container hostname is independently set. Host hostname unchanged.

**Tests to write FIRST:**
- Set hostname to "container-test" inside → `hostname` returns "container-test"
- Host `hostname` returns the original host hostname

**Checkpoint questions:**
1. Why is hostname isolation needed? What breaks if containers share hostname?
2. What does UTS stand for?

---

### M1.4: IPC Namespace (1 day)

**Goal:** Container gets its own System V IPC objects (shared memory, semaphores, message queues).

**Concepts to study BEFORE coding:**
- `man 7 svipc` — System V IPC overview
- `man 7 ipc_namespaces`
- `man 1 ipcs` — how to list IPC resources

**Deliverable:** IPC resources created inside the container are invisible on the host.

**Tests to write FIRST:**
- Create a shared memory segment inside container → `ipcs` on host doesn't show it
- Two containers can't see each other's IPC objects

**Checkpoint questions:**
1. What attack is possible if containers share an IPC namespace?
2. What IPC mechanisms does this namespace isolate? What does it NOT isolate?

**Suggested todo order:**
- [ ] Study: `man 7 ipc_namespaces` — what an IPC namespace isolates and how it is created
- [ ] Study: `man 7 svipc` — shared memory, semaphores, message queues, and the permission model
- [ ] Study: `man 1 ipcs`, `man 1 ipcmk`, and `man 1 ipcrm` — inspect, create, and remove IPC objects from the shell
- [ ] Study: `man 2 shmget`, `man 2 semget`, and `man 2 msgget` — the syscalls behind the userland tools
- [ ] Understand: which IPC resources are namespace-scoped, which are not, and what can still leak between containers
- [ ] Understand: why shared IPC state is a security and stability risk for containers
- [ ] Answer the 2 checkpoint questions above before writing any code

---

### M1.5: Network Namespace (Basic — Loopback Only) (2 days)

**Goal:** Container has its own empty network namespace with only loopback.

**Concepts to study BEFORE coding:**
- `man 7 network_namespaces`
- `man 7 netlink` — the kernel API for network configuration
- `man 8 ip-netns` — the `ip` tool's namespace support
- Try manually: `ip netns add test && ip netns exec test ip link show`

**Deliverable:** Container's `ip link show` shows only `lo`. No host interfaces visible.

**Tests to write FIRST:**
- `ip link show` inside container → only `lo` interface
- Host interfaces (eth0, etc.) NOT visible inside container
- After `lo` is brought up: `ping 127.0.0.1` works inside container

**Checkpoint questions:**
1. Why does a new network namespace start completely empty — not even loopback?
2. How is this different from just not having network access?
3. What syscall brings up the loopback interface?

---

### ⚠️ M1.6: User Namespace — UID Remapping (3 days)

**Goal:** Root (UID 0) inside the container maps to an unprivileged UID on the host.

**Concepts to study BEFORE coding:**
- `man 7 user_namespaces` — the most complex namespace, read carefully
- `/proc/[pid]/uid_map` and `gid_map` format
- `/proc/[pid]/setgroups` — must write "deny" before writing gid_map
- Read: runc's userns implementation
- Understand: namespace creation ordering — user namespace typically first, then others

**Deliverable:** Root inside container = unprivileged user on host. Container root cannot modify host files.

**Tests to write FIRST:**
- `id` inside container → uid=0(root)
- From host: `cat /proc/<pid>/status` → Uid shows unprivileged UID (e.g., 100000)
- Container cannot write to host paths outside its rootfs

**Checkpoint questions:**
1. If you skip the user namespace, what can root inside the container do to the host?
2. Why must user namespace be created before other namespaces in some configurations?
3. What is the `setgroups` deny requirement and why does it exist?

---

## Phase 2: Resource Limits with Cgroups v2

### M2.1: Cgroup v2 Foundation (2 days)

**Goal:** Create a cgroup hierarchy and enable controllers.

**Concepts to study BEFORE coding:**
- `man 7 cgroups` — cgroup overview
- https://docs.kernel.org/admin-guide/cgroup-v2.html — the definitive reference
- Understand: unified hierarchy (v2) vs per-controller hierarchies (v1)
- Understand: `cgroup.subtree_control` — how to enable controllers for child cgroups
- Explore: `ls /sys/fs/cgroup/` on your system — read the files

**Deliverable:** Runtime creates `/sys/fs/cgroup/myruntime/<id>/` with controllers enabled.

**Tests to write FIRST:**
- After Create: cgroup directory exists at expected path
- Controllers (cpu, memory, pids) are listed in cgroup.controllers
- After Destroy: cgroup directory is removed

**Checkpoint questions:**
1. What's the fundamental difference between cgroup v1 and v2?
2. Why can't you enable a controller in a cgroup that has processes AND child cgroups?
3. What happens if you try to rmdir a cgroup that still has processes?

---

### M2.2: CPU Limits (2 days)

**Goal:** Cap a container's CPU usage using cpu.max.

**Concepts to study BEFORE coding:**
- `cpu.max` format: `"$QUOTA $PERIOD"` — microseconds per period
- `cpu.stat` — throttle counters (nr_throttled, throttled_usec)
- Understand: `50000 100000` = 50ms every 100ms = 50% of one core
- Understand: `200000 100000` = 200ms every 100ms = 2 full cores

**Deliverable:** Container with 50% CPU limit gets throttled when exceeding it.

**Tests to write FIRST:**
- Set cpu.max to "50000 100000", run CPU-bound task → cpu.stat shows throttling
- Container CPU usage stays near 50% (check via stats)

**Checkpoint questions:**
1. What does the kernel do when a container exceeds its CPU quota in a period?
2. How do you limit a container to 2 CPU cores? What about 0.1 cores?
3. What's the difference between CPU throttling and CPU shares/weight?

---

### ⚠️ M2.3: Memory Limits and OOM (2 days)

**Goal:** Enforce memory limits. Container gets OOM-killed when exceeding the limit.

**Concepts to study BEFORE coding:**
- `memory.max` — hard limit
- `memory.current` — current usage
- `memory.events` — oom, oom_kill counts
- `man 5 proc` — `/proc/[pid]/oom_score_adj`
- Understand: kernel page reclaim → OOM killer → cgroup-scoped OOM

**Deliverable:** Container with 64MB limit gets OOM-killed when allocating 128MB.

**Tests to write FIRST:**
- Container with 64MB limit: allocate 128MB → exit code 137 (SIGKILL)
- `memory.events` shows oom_kill count incremented
- Host memory is not affected

**Checkpoint questions:**
1. Walk through what happens kernel-side when a container hits memory.max.
2. Why is exit code 137 specifically? (128 + signal number)
3. What's the difference between memory.max and memory.high?

---

### M2.4: PID Limits (1 day)

**Goal:** Prevent fork bombs by limiting the number of processes in a container.

**Concepts to study BEFORE coding:**
- `pids.max` — maximum PIDs (includes threads)
- `pids.current` — current count
- Try a fork bomb without limits first (in a VM!): `:(){ :|:& };:`

**Deliverable:** Fork bomb inside container is contained. Host is unaffected.

**Tests to write FIRST:**
- Container with pids.max=32: fork bomb fails with EAGAIN
- Host PID count is unchanged during container fork bomb
- `pids.current` never exceeds `pids.max`

**Checkpoint questions:**
1. Why is PID limiting critical for host protection?
2. Does pids.max count threads or only processes?
3. What error does fork() return when the PID limit is hit?

---

### M2.5: IO Limits (2 days)

**Goal:** Throttle container disk IO.

**Concepts to study BEFORE coding:**
- `io.max` format: `"MAJOR:MINOR rbps=X wbps=X riops=X wiops=X"`
- `io.stat` — per-device IO statistics
- Find your disk's major:minor: `cat /proc/partitions` or `lsblk`
- Understand: IO limits only apply to direct IO, not buffered IO in page cache

**Deliverable:** Container writes are throttled to configured limit.

**Tests to write FIRST:**
- Set wbps=1048576 (1MB/s), run `dd` with direct IO → throughput ~1MB/s
- Without limit, `dd` runs at full disk speed

**Checkpoint questions:**
1. Why doesn't io.max affect buffered writes? What's the page cache?
2. What's the difference between rbps (rate bytes per second) and riops (rate IO operations per second)?
3. How do you find the major:minor number for a device?

---

### M2.6: Cgroup Stats Collection (1 day)

**Goal:** Read and return live resource usage from cgroup files.

**Concepts to study BEFORE coding:**
- All the stat file formats from M2.2–M2.5
- `cpu.stat` key-value format
- How to calculate CPU percentage from cumulative usage_usec

**Deliverable:** `myruntime stats <id>` shows CPU%, memory, PIDs in real time.

**Tests to write FIRST:**
- Stats for running container return non-zero CPU and memory values
- Memory current <= memory max
- PID count matches running processes

**Checkpoint questions:**
1. How do you compute CPU percentage from two readings of usage_usec?
2. What does `nr_throttled` in cpu.stat tell you about your CPU limit configuration?

---

## Phase 3: Overlay Filesystem

### M3.1: Basic Overlay Mount (2 days)

**Goal:** Mount an overlay filesystem with lower + upper + work directories.

**Concepts to study BEFORE coding:**
- `man overlayfs` (or kernel docs: Documentation/filesystems/overlayfs.rst)
- Understand: lowerdir (read-only), upperdir (writable), workdir (kernel scratch)
- Understand: copy-on-write — what happens when you modify a file from a lower layer
- Understand: whiteout files — how deletes work in overlay

**Deliverable:** Files from lower layers visible in merged view. Writes go to upper only.

**Tests to write FIRST:**
- Create file in lowerdir → visible in merged
- Write file in merged → appears in upperdir, NOT in lowerdir
- Delete file from lower via merged → whiteout in upperdir, lower untouched

**Checkpoint questions:**
1. What happens at the filesystem level when a container modifies a file from a lower layer?
2. What is a whiteout file and why is it needed?
3. Why does overlay need a workdir? What is it used for?

---

### M3.2: Layer Management (2 days)

**Goal:** Content-addressable layer storage with deduplication.

**Concepts to study BEFORE coding:**
- SHA256 content addressing — how Docker identifies layers
- Tar format basics
- Why layer dedup matters: 50 containers on same base = 1 copy on disk

**Deliverable:** Layers stored by SHA256 digest. Same layer extracted once, shared by many containers.

**Tests to write FIRST:**
- Extract same tar twice → only one directory on disk
- Two containers using same base layer → both reference same directory
- Disk usage for N containers with same base ≈ disk usage for 1

**Checkpoint questions:**
1. Why use SHA256 for layer identification instead of a random ID?
2. What happens if two different layers produce the same SHA256? (collision)
3. Why must lower layers be read-only?

---

### ⚠️ M3.3: Device Nodes and Special Mounts (3 days)

**Goal:** Create /dev with essential devices, /dev/pts, /dev/shm inside container.

**Concepts to study BEFORE coding:**
- `man 2 mknod` — creating device files
- Device major/minor numbers: `man 4 null`, `man 4 zero`, `man 4 random`
- devpts: `man 4 pts` — pseudoterminals
- Why each device matters: /dev/null (discard), /dev/urandom (TLS), /dev/pts (shells)
- Read: runc's `rootfs_linux.go` → `createDevices` function

**Deliverable:** Interactive shell works inside container. TLS/random works. /dev/null works.

**Tests to write FIRST:**
- `echo hello > /dev/null` → no error
- `head -c 16 /dev/urandom | xxd` → produces random bytes
- Interactive shell (requires /dev/pts) → can type commands
- `ls /dev/` shows null, zero, full, random, urandom, tty, fd, stdin, stdout, stderr

**Checkpoint questions:**
1. What breaks if /dev/urandom doesn't exist? Why?
2. What's the difference between device major and minor numbers?
3. Why mount tmpfs on /dev instead of using the host's /dev?

---

## Phase 4: OCI Image Support

### M4.1: Registry Authentication (2 days)

**Goal:** Get bearer tokens from Docker Hub's auth service.

**Concepts to study BEFORE coding:**
- OCI Distribution Spec: https://github.com/opencontainers/distribution-spec
- Docker registry HTTP API v2
- Token-based auth flow: request → 401 → get token from auth endpoint → retry with token
- `curl` the auth endpoint manually first

**Deliverable:** Successfully authenticate with registry-1.docker.io.

**Tests to write FIRST:**
- Get token for library/alpine:latest → non-empty bearer token
- Use token to make authenticated request → 200 response

**Checkpoint questions:**
1. Walk through the auth flow step by step. What happens on a 401?
2. What's the difference between anonymous and authenticated pulls?

---

### M4.2: Manifest Fetching (2 days)

**Goal:** Download and parse OCI image manifests.

**Concepts to study BEFORE coding:**
- OCI Image Manifest Spec
- Manifest media types: `application/vnd.oci.image.manifest.v1+json`
- Manifest list / index (multi-arch images)
- `curl` a manifest manually to see the structure

**Deliverable:** Parse manifest for nginx:latest — extract layer digests and config digest.

**Tests to write FIRST:**
- Fetch alpine:latest manifest → valid JSON with layers array
- Layer digests are valid SHA256 strings
- Config digest points to a valid config blob

**Checkpoint questions:**
1. What information does a manifest contain?
2. How do multi-architecture images work? What's a manifest list?

---

### M4.3: Layer Download (3 days)

**Goal:** Download layer blobs from the registry, with progress tracking.

**Concepts to study BEFORE coding:**
- HTTP blob download with Content-Length for progress
- Parallel downloads using goroutines
- Digest verification after download (SHA256 must match)

**Deliverable:** Download all layers for ubuntu:22.04 with progress bars.

**Tests to write FIRST:**
- Download alpine:latest layers → files exist on disk
- SHA256 of downloaded blob matches digest from manifest
- Parallel download of 3+ layers completes faster than sequential

**Checkpoint questions:**
1. Why verify the SHA256 after download? What attack does this prevent?
2. How would you resume a partially downloaded layer?

---

### M4.4: Layer Unpacking (2 days)

**Goal:** Extract tar.gz layers to the layer store.

**Concepts to study BEFORE coding:**
- Go's `archive/tar` and `compress/gzip` packages
- Tar entry types: regular files, directories, symlinks, hardlinks
- Whiteout files: `.wh.filename` means "delete filename from lower layer"
- Security: path traversal in tar archives (../../../etc/passwd)

**Deliverable:** Layers extracted with correct permissions, ownership, and symlinks.

**Tests to write FIRST:**
- Extract a layer → directory structure matches tar contents
- File permissions preserved correctly
- Symlinks created correctly
- Whiteout files detected (not extracted as regular files)

**Checkpoint questions:**
1. What are whiteout files and how does the overlay filesystem use them?
2. How can a malicious tar archive attack the extraction process? (path traversal)
3. Why must you handle hardlinks carefully during extraction?

---

### M4.5: Image Config Parsing (2 days)

**Goal:** Parse OCI image config to get runtime defaults (entrypoint, env, user).

**Concepts to study BEFORE coding:**
- OCI Image Configuration Spec
- How Cmd and Entrypoint interact (Entrypoint + Cmd = full command)
- Environment variable inheritance

**Deliverable:** nginx image runs with correct entrypoint, env vars, and working directory.

**Tests to write FIRST:**
- Parse nginx config → Entrypoint = ["/docker-entrypoint.sh"]
- Parse nginx config → Cmd = ["nginx", "-g", "daemon off;"]
- Env contains PATH and NGINX_VERSION

**Checkpoint questions:**
1. What command actually runs if Entrypoint is ["/bin/sh", "-c"] and Cmd is ["echo hello"]?
2. How does a user override the entrypoint vs the command?
3. What's the difference between Env in the image config and Env passed at runtime?

---

### 🎯 M4.6: Local Image Storage (1 day)

**Goal:** Store pulled images locally. List them with `myruntime images`.

**Deliverable:** Pull nginx → it's stored locally. `myruntime images` lists it. Running it doesn't re-download.

**Tests to write FIRST:**
- Pull alpine:latest → stored on disk
- `List()` returns alpine:latest
- Second pull of same image → skips download (already exists)

**Checkpoint questions:**
1. How does the image store know if an image is already downloaded?
2. What's the difference between an image ID and a tag?

**Demo:** `myruntime pull nginx:latest && myruntime images` — real image from Docker Hub, stored locally.

---

## Phase 5: Container Lifecycle

### M5.1: State Persistence (2 days)

**Goal:** Container state survives runtime restarts. Stored as JSON on disk.

**Concepts to study BEFORE coding:**
- JSON serialization in Go
- File atomicity — write to temp file then rename (prevents corruption)
- What state must persist: ID, state, PID, config, timestamps, exit code

**Deliverable:** Restart the runtime → it remembers all containers and their states.

**Tests to write FIRST:**
- Save container → Load container → fields match
- List returns all saved containers
- Delete removes state from disk

**Checkpoint questions:**
1. Why write-then-rename instead of writing directly to the state file?
2. What state do you NOT need to persist? (hint: anything reconstructable)

---

### M5.2: Container Create (3 days)

**Goal:** Set up everything for a container without starting a process.

**Deliverable:** `create` resolves image, mounts overlay, creates cgroup. State = "created".

**Tests to write FIRST:**
- After create: state is "created", PID is 0
- Overlay mount exists at expected path
- Cgroup directory exists with limits set
- Image layers are set up as overlay lower dirs

**Checkpoint questions:**
1. Why separate create and start? What's the use case?
2. In what order must resources be set up? What depends on what?

---

### ⚠️ M5.3: Container Start (3 days)

**Goal:** Fork the init process into new namespaces with proper synchronization.

**Concepts to study BEFORE coding:**
- Go's `os/exec` + `SysProcAttr` for clone flags
- The re-exec pattern: why `exec.Command("/proc/self/exe", "init")`
- Parent-child synchronization via pipes: parent sets up cgroup → signals child → child execs
- Race condition: if child execs before parent adds it to cgroup, limits don't apply

**Deliverable:** Container process runs isolated. Cgroup limits active from the start.

**Tests to write FIRST:**
- After start: state is "running", PID > 0
- Process exists on host (kill -0 succeeds)
- Process is in the correct cgroup (check /proc/<pid>/cgroup)
- All namespaces are different from host (compare /proc/<pid>/ns/*)

**Checkpoint questions:**
1. Why is the pipe synchronization between parent and child necessary? What goes wrong without it?
2. What happens if the container's entrypoint exits immediately? How do you detect it?
3. Why re-exec instead of just calling a Go function in the child?

---

### M5.4: Container Exec (2 days)

**Goal:** Run a command inside a running container's namespaces.

**Concepts to study BEFORE coding:**
- `man 2 setns` — enter an existing namespace
- `/proc/<pid>/ns/*` — namespace file descriptors
- This is how `docker exec` works — no new container, just join existing namespaces

**Deliverable:** `myruntime exec <id> /bin/sh` gives an interactive shell inside the container.

**Tests to write FIRST:**
- Exec `hostname` in container → returns container's hostname
- Exec `cat /proc/1/status` → shows container's PID 1, not host's
- Exec sees the same filesystem as the container's entrypoint

**Checkpoint questions:**
1. How is exec different from starting a new container?
2. What namespaces do you need to join? Does the order matter?
3. The exec'd process shares the container's cgroup — why is this important?

---

### M5.5: Container Stop and Kill (2 days)

**Goal:** Graceful shutdown: SIGTERM first, SIGKILL after timeout.

**Concepts to study BEFORE coding:**
- Signal handling in Linux: SIGTERM (catchable) vs SIGKILL (not catchable)
- `man 2 wait4` — waiting for child process exit
- Timeout pattern in Go: `select` with `time.After`

**Deliverable:** `stop` gives the process a chance to shutdown. `kill` is immediate.

**Tests to write FIRST:**
- Stop a container that handles SIGTERM → exits cleanly, exit code 0
- Stop a container that ignores SIGTERM → killed after timeout, exit code 137
- Kill sends signal immediately without waiting

**Checkpoint questions:**
1. Why send SIGTERM before SIGKILL? What's the practical difference?
2. What happens to child processes when PID 1 is killed?
3. What's a reasonable default timeout? What does Docker use?

---

### 🎯 M5.6: Container Delete (2 days)

**Goal:** Clean up all container resources: cgroup, overlay, network, state.

**Deliverable:** After delete, zero traces of the container remain.

**Tests to write FIRST:**
- Cannot delete a running container → error
- After delete: cgroup directory gone, overlay unmounted, state file removed
- No leftover mounts in /proc/mounts

**Checkpoint questions:**
1. In what order should resources be cleaned up? Why does order matter?
2. What happens if cleanup fails halfway? How do you handle partial cleanup?

**Demo:** Full lifecycle: `create → start → exec → stop → delete`. Show each state transition.

---

## Phase 6: Container Networking

### M6.1: Bridge Creation (2 days)

**Goal:** Create a virtual network bridge on the host.

**Concepts to study BEFORE coding:**
- What is a network bridge? (virtual L2 switch)
- `man 8 bridge` — bridge management
- `man 8 ip-link` — interface management
- Try manually: `ip link add myruntime0 type bridge && ip addr add 172.20.0.1/24 dev myruntime0 && ip link set myruntime0 up`
- `github.com/vishvananda/netlink` — the Go library for netlink operations

**Deliverable:** Bridge interface exists on host with an assigned IP address.

**Tests to write FIRST:**
- After setup: `ip link show myruntime0` → bridge exists and is UP
- Bridge has IP 172.20.0.1/24
- IP forwarding is enabled

**Checkpoint questions:**
1. What's the difference between a bridge and a router?
2. Why does the bridge need an IP address? What role does it play?
3. Why must IP forwarding be enabled on the host?

---

### ⚠️ M6.2: veth Pair and Container Connection (3 days)

**Goal:** Create veth pairs, move one end into container namespace, attach other to bridge.

**Concepts to study BEFORE coding:**
- `man 4 veth` — virtual ethernet device pairs
- Moving interfaces between network namespaces
- Try manually: `ip link add veth0 type veth peer name veth1 && ip link set veth1 netns <pid>`
- Timing: the veth must be created AFTER the container's net namespace exists

**Deliverable:** Container has eth0 connected to host bridge via veth pair.

**Tests to write FIRST:**
- Container `ip link show` → shows eth0 (the veth end)
- Host `ip link show` → shows veth_xxx attached to bridge
- Container can ping the bridge gateway (172.20.0.1)

**Checkpoint questions:**
1. What is a veth pair physically? Describe the data path.
2. Why must one end be moved into the container's net namespace? Why not both ends on the host?
3. What happens to the veth pair when the container is deleted?

---

### M6.3: IP Address Management (2 days)

**Goal:** Allocate unique IPs from the bridge subnet for each container.

**Concepts to study BEFORE coding:**
- CIDR notation and subnet math
- Why .1 is reserved for the gateway
- IP exhaustion — what happens when the subnet is full

**Deliverable:** Each container gets a unique IP. Released on delete.

**Tests to write FIRST:**
- First container gets 172.20.0.2
- Second container gets 172.20.0.3
- After first container deleted, IP 172.20.0.2 is available again
- 254th container → error (subnet full for /24)

**Checkpoint questions:**
1. How many usable IPs does a /24 subnet have? Why not 256?
2. How do you prevent two containers from getting the same IP?

---

### M6.4: NAT and Internet Access (2 days)

**Goal:** Containers can reach the internet via iptables MASQUERADE.

**Concepts to study BEFORE coding:**
- `man 8 iptables` — packet filtering and NAT
- MASQUERADE target: replaces source IP with host's outbound IP
- POSTROUTING chain: applies after routing decision, before packet leaves host
- Try manually: `iptables -t nat -A POSTROUTING -s 172.20.0.0/24 -o eth0 -j MASQUERADE`

**Deliverable:** Container can ping 8.8.8.8 and reach the internet.

**Tests to write FIRST:**
- Container `ping -c 1 8.8.8.8` → succeeds
- Container `curl -s ifconfig.me` → returns host's public IP (MASQUERADE working)

**Checkpoint questions:**
1. What does MASQUERADE do to the packet? Draw the source/dest IP at each hop.
2. How does the return traffic find its way back to the container?
3. What's the difference between MASQUERADE and SNAT?

---

### M6.5: Port Mapping (2 days)

**Goal:** Expose container ports on host via iptables DNAT.

**Concepts to study BEFORE coding:**
- DNAT: rewrite destination address
- PREROUTING chain: applies before routing decision
- Try manually: `iptables -t nat -A PREROUTING -p tcp --dport 8080 -j DNAT --to-destination 172.20.0.2:80`

**Deliverable:** `myruntime run -p 8080:80 nginx` makes nginx reachable at host:8080.

**Tests to write FIRST:**
- Start container with -p 8080:80 running a simple HTTP server
- `curl localhost:8080` from host → response from container
- After container deleted → port mapping rule removed

**Checkpoint questions:**
1. What does DNAT do to the destination address? Draw the packet flow.
2. Why is the rule in PREROUTING and not FORWARD?
3. What happens if two containers try to map the same host port?

---

### M6.6: DNS Configuration (1 day)

**Goal:** Containers can resolve domain names.

**Deliverable:** Container can `nslookup google.com`.

**Tests to write FIRST:**
- `/etc/resolv.conf` inside container contains nameserver entries
- `nslookup google.com` inside container → resolves successfully

**Checkpoint questions:**
1. Why not just bind-mount the host's /etc/resolv.conf?
2. What happens if the DNS server is unreachable?

---

### 🎯 M6.7: Container-to-Container Communication (2 days)

**Goal:** Multiple containers on the same bridge can communicate by IP.

**Deliverable:** Container A can ping Container B. Both can reach the internet.

**Tests to write FIRST:**
- Start two containers → they get different IPs
- From container A: `ping <container-B-IP>` → succeeds
- Both containers can independently reach the internet

**Checkpoint questions:**
1. How does the bridge know where to forward packets between containers?
2. Could you implement container-to-container DNS (resolve by name)? How?

**Demo:** Two containers pinging each other across the bridge. Show the veth pairs, bridge, and iptables rules.

---

## Phase 7: CLI Interface

### M7.1: Core Commands (3 days)

**Goal:** Implement all CLI commands: run, create, start, stop, kill, rm, ps, exec, pull, images.

**Concepts to study BEFORE coding:**
- `github.com/urfave/cli/v2` — CLI framework
- Flag parsing, subcommands, help text
- How Docker's CLI is organized (for UX reference)

**Deliverable:** All commands work end-to-end.

**Tests to write FIRST:**
- `myruntime run alpine echo hello` → prints "hello"
- `myruntime ps` → lists running containers
- `myruntime exec <id> cat /etc/hostname` → returns container hostname

**Checkpoint questions:**
1. What's the difference between `run` and `create + start`?
2. How should errors be reported to the user?

---

### M7.2: Container Logging (2 days)

**Goal:** Capture stdout/stderr to files. View with `myruntime logs`.

**Concepts to study BEFORE coding:**
- Redirecting child process stdout/stderr to files
- Log rotation (optional, but know why it matters)

**Deliverable:** `myruntime logs <id>` shows container output.

**Tests to write FIRST:**
- Container runs `echo hello` → logs show "hello"
- Container writes to stderr → captured separately
- Logs persist after container stops

**Checkpoint questions:**
1. Where should log files be stored?
2. What happens if a container generates gigabytes of log output?

---

### 🎯 M7.3: Stats Command (2 days)

**Goal:** Live resource usage display from cgroup stats.

**Deliverable:** `myruntime stats <id>` shows CPU%, memory, PIDs updating in real time.

**Tests to write FIRST:**
- Stats for running container show non-zero values
- CPU% increases during CPU-bound work
- Memory usage reflects actual allocation

**Checkpoint questions:**
1. How do you calculate CPU percentage between two readings?
2. What refresh rate makes sense? Why?

**Demo:** Run `stress` inside a container while watching `myruntime stats` — see CPU and memory climb.

---

### M7.4: Inspect Command (1 day)

**Goal:** Show detailed container info as JSON.

**Deliverable:** `myruntime inspect <id>` outputs full state, config, network, mounts.

**Tests to write FIRST:**
- Inspect returns valid JSON
- Contains: state, config, PID, IP, cgroup path, created/started timestamps

**Checkpoint questions:**
1. What information is most useful for debugging container issues?

---

## Phase 8: Signal Handling and Robustness

### M8.1: Signal Forwarding (2 days)

**Goal:** Runtime forwards SIGTERM/SIGINT to the container process.

**Concepts to study BEFORE coding:**
- Go's `os/signal` package
- Which signals to forward and which to handle ourselves
- What happens if the runtime is killed with SIGKILL (no chance to forward)

**Deliverable:** Ctrl-C on the runtime gracefully stops the container.

**Tests to write FIRST:**
- Send SIGTERM to runtime → container receives SIGTERM and exits
- Send SIGINT to runtime → container exits gracefully
- Container that traps SIGTERM can clean up before exiting

**Checkpoint questions:**
1. Why forward signals instead of just letting the kernel handle it?
2. What signals should NOT be forwarded?
3. What happens to the container if the runtime is SIGKILL'd?

---

### ⚠️ M8.2: Zombie Process Reaping (3 days)

**Goal:** Container PID 1 properly reaps orphaned child processes.

**Concepts to study BEFORE coding:**
- `man 2 wait` — the wait() family of syscalls
- What is a zombie process? (exited but parent hasn't called wait())
- Why PID 1 is special: it inherits orphaned processes
- What `tini` does and why Docker uses it
- Try: run a program that forks and doesn't wait() — watch zombies accumulate

**Deliverable:** No zombie processes accumulate inside containers.

**Tests to write FIRST:**
- Run a program that creates 10 child processes that exit → no zombies remain
- Container PID 1 reaps orphaned children automatically
- `ps` inside container shows no defunct/zombie processes

**Checkpoint questions:**
1. Why does PID 1 inherit orphaned processes? Where is this in the kernel?
2. What happens if PID 1 never calls wait()? What's the consequence?
3. Why is this different from a normal process?

---

### M8.3: Crash Recovery (2 days)

**Goal:** Runtime reconciles state on startup after a crash.

**Concepts to study BEFORE coding:**
- `kill(pid, 0)` — check if a process exists without sending a signal
- What resources leak after a crash: cgroups, mounts, state files
- Reconciliation pattern: compare expected state with actual state

**Deliverable:** After runtime crash and restart, orphaned containers are cleaned up.

**Tests to write FIRST:**
- Simulate crash: kill runtime while container is running
- Restart runtime → detects dead container, updates state to stopped
- Orphaned cgroup is cleaned up

**Checkpoint questions:**
1. What resources can leak if the runtime crashes? List all of them.
2. How do you detect that a "running" container is actually dead?
3. What should you NOT clean up automatically? (e.g., logs, image layers)

---

## Phase 9: Testing

### M9.1: Namespace Isolation Tests (2 days)

**Goal:** Prove that namespace isolation works correctly.

**Deliverable:** Test suite covering PID, hostname, filesystem, network, IPC, and user isolation.

**Tests:** See `tests/isolation_test.go` for the full list.

---

### M9.2: Cgroup Enforcement Tests (2 days)

**Goal:** Prove that resource limits are actually enforced by the kernel.

**Deliverable:** Tests that deliberately exceed limits and verify enforcement.

**Tests:** See `tests/cgroup_test.go` for the full list.

---

### 🎯 M9.3: Stress Tests (2 days)

**Goal:** Push the runtime to its limits. 50 containers, fork bombs, disk fill.

**Deliverable:** Runtime handles stress without leaking resources or crashing.

**Tests:** See `tests/stress_test.go` for the full list.

**Demo:** Start 50 containers, show stats, run a fork bomb in one — everything else is fine.

---

### M9.4: Integration Tests (1 day)

**Goal:** End-to-end workflows from pull to delete.

**Deliverable:** Full workflow tests that exercise the entire runtime.

**Tests:**
- Pull → create → start → exec → stop → delete → verify cleanup
- Pull same image twice → second time is cached
- Run two containers → they can communicate over bridge
