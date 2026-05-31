# ADR-001 — Re-exec `/proc/self/exe` with an `init` subcommand vs. direct `clone()`

**Status:** Accepted
**Date:** 2026-06-01

## Context

To isolate a process, the runtime must create new namespaces (`CLONE_NEW*`)
and then run the container's entrypoint inside them. There are two ways to get
a process into fresh namespaces:

1. **Call `clone(2)`/`unshare(2)` directly** from the running Go program and
   continue executing in the child.
2. **Set the clone flags on a re-exec of our own binary** — spawn
   `/proc/self/exe` with a subcommand that runs the in-namespace setup, so the
   namespaces take effect on a brand-new process image.

The constraint that decides this: the Go runtime is **multi-threaded** —
goroutines are multiplexed onto OS threads, and the scheduler may have spawned
several before `main` runs. `CLONE_NEWUSER` (and namespace creation generally)
requires a single-threaded caller; doing it from inside the live Go runtime is
unsafe and, for the user namespace, rejected by the kernel.

## Decision

Use the **re-exec pattern**. The parent (`run`) builds an `init` argv and runs
`exec.Command("/proc/self/exe", initArgs...)` with `SysProcAttr.Cloneflags`
set; the freshly-`exec`'d child enters `main()`'s `init` branch as a fresh,
effectively single-threaded process, performs the in-namespace setup
(`pivot_root`, `/proc`, hostname, loopback), and then `syscall.Exec`s the
entrypoint.

Implemented at [cmd/myruntime/main.go:93-139](../../cmd/myruntime/main.go#L93-L139)
(parent `run`) and [cmd/myruntime/main.go:42-91](../../cmd/myruntime/main.go#L42-L91)
(child `init`). The privileged-operation tests use the same trick, re-execing
the test binary as a helper subprocess gated on `GO_WANT_*_HELPER`
([internal/namespace/namespace_test.go](../../internal/namespace/namespace_test.go)).

## Rejected alternative

**Direct `clone()`/`unshare()` from the live runtime.** It would avoid a second
process and the argv/pipe plumbing. But it fights the Go runtime: namespace
flags must take effect on a single-threaded process, which the Go scheduler
does not guarantee, and `runtime.LockOSThread` does not make the *process*
single-threaded. This is the same reason `runc` re-execs a C `nsenter`
constructor before the Go runtime starts. Re-exec also keeps CLI and
container-init logic in one binary, dispatched by subcommand.

## Consequences

**What we commit to:**
- The parent/child boundary is a **process boundary**: configuration crosses it
  via argv (and, in the designed lifecycle, a synchronization pipe).
- Cross-boundary error propagation is limited to exit codes plus stderr text —
  the parent cannot distinguish a mount failure from an exec failure beyond the
  message string ([main.go:136-139](../../cmd/myruntime/main.go#L136-L139)).

**What becomes harder:**
- Operations that must happen *after* the child exists but *before* it execs
  the entrypoint — writing UID/GID maps, adding the PID to a cgroup — require
  the parent to act on the child mid-flight, gated by a pipe. That
  synchronization is documented in [internal/container/lifecycle.go:39-41](../../internal/container/lifecycle.go#L39-L41)
  but is **not yet built**, which is why the live `run` path applies neither
  ID maps nor cgroup limits.

## Notes for reviewers

The tell that this is the right call: every production Go runtime (`runc`,
`containerd`'s shim) does the same re-exec dance, because there is no
thread-safe way to namespace a process from inside a started Go runtime. The
cost is the parent↔child plumbing, which is exactly where this project's
lifecycle work (M5) still has gaps.
