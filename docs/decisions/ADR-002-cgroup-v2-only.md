# ADR-002 — cgroup v2 unified hierarchy only, via direct sysfs writes

**Status:** Accepted
**Date:** 2026-06-01

## Context

Containers need resource limits — CPU, memory, PID count — enforced by the
kernel's control groups. There are three layers of choice:

1. **cgroup v1** (per-controller hierarchies, each mounted separately) vs.
   **cgroup v2** (a single unified hierarchy with `cgroup.subtree_control`) vs.
   a **hybrid** v1/v2 compatibility layer.
2. **Talk to the kernel directly** (read/write the cgroup files under
   `/sys/fs/cgroup`) vs. **delegate to `systemd`** over D-Bus, or use a
   `libcgroup`-style library.

## Decision

Target **cgroup v2 only**, written **directly to sysfs**. The manager reads
`cgroup.controllers`, enables the supported controllers (`cpu`, `io`, `memory`,
`pids`) by writing `cgroup.subtree_control`, then writes `cpu.max`,
`memory.max`, and `pids.max` under `/sys/fs/cgroup/myruntime/<id>/`.

Implemented at [internal/cgroup/manager.go](../../internal/cgroup/manager.go):
controller enablement at lines 153-184, `cpu.max` at 186-197, `memory.max` at
199-206, `pids.max` at 257-263; `AddProcess` writes the PID to `cgroup.procs`
([manager.go:74-91](../../internal/cgroup/manager.go#L74-L91)). CPU throttling
is verified by `TestCPULimit`.

## Rejected alternative

**cgroup v1 / hybrid, or `systemd` delegation.** v1 is still widely deployed and
maximally compatible. `systemd` delegation is what production runtimes
(`runc` + `containerd`) use, and it gets cleanup and delegation for free.

Both were rejected for a learning runtime: v1's per-controller mount points
multiply the bookkeeping for no pedagogical gain, and a `systemd`/D-Bus
dependency would (a) break the stdlib-only constraint and (b) hide the very
mechanism the project exists to expose. Direct file I/O against v2 is
transparent, inspectable with `cat`, and mirrors exactly what `runc` writes.

## Consequences

**What we commit to:**
- The host **must** boot with the unified hierarchy
  (`systemd.unified_cgroup_hierarchy=1` or `cgroup_no_v1=all`). On a v1 or
  hybrid host, `cgroup.controllers` is absent and controller enablement
  silently no-ops, so limits never apply.
- No `systemd` integration means **no help with delegation or cleanup** —
  orphaned cgroup directories are ours to remove.

**What becomes harder:**
- `io.max` is not yet written (`TODO(M2.5)`, [manager.go:70](../../internal/cgroup/manager.go#L70)).
- A second, parallel stats API exists in
  [internal/cgroup/stats.go](../../internal/cgroup/stats.go) (`ReadCPUStat`,
  `CollectStats`, …) that is **entirely stubbed and unused** — `Manager.Stats`
  uses its own inline readers ([manager.go:122-151](../../internal/cgroup/manager.go#L122-L151)).
  It is a planned refactor target, not the live stats path; don't mistake it for one.

## Notes for reviewers

`validateContainerID` ([manager.go:265-286](../../internal/cgroup/manager.go#L265-L286))
rejects empty, absolute, `.`/`..`, and separator-bearing IDs before any path is
constructed, so a hostile container ID cannot redirect a sysfs write outside the
hierarchy (threat T-2). The conservative `Destroy` refuses to remove a cgroup
that still lists live PIDs ([manager.go:101-118](../../internal/cgroup/manager.go#L101-L118)).
