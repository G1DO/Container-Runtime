# ADR-003 — `pivot_root(2)` over `chroot` for root-filesystem isolation

**Status:** Accepted
**Date:** 2026-06-01

## Context

Once a container is in its own mount namespace, its root filesystem must be
swapped from the host's to the container's `rootfs`. Two syscalls can do this:

1. **`chroot(2)`** — change the apparent root directory. Simple, one call.
2. **`pivot_root(2)`** — move the current root mount to a new location and make
   a new filesystem the root, then unmount the old root entirely.

The security difference is decisive. `chroot` only changes the *path
resolution* root; it does not sever file descriptors or mounts that were open
to the old root before the call. A process that holds such an fd (or that
regains one, e.g. via a leaked descriptor) can `fchdir` out of the chroot and
walk the host filesystem — the classic chroot escape.

## Decision

Use **`pivot_root`**. `PivotRoot`
([internal/filesystem/mounts.go:27-74](../../internal/filesystem/mounts.go#L27-L74)):

1. Makes `/` recursively private (`MS_REC | MS_PRIVATE`) so nothing propagates
   to the host mount table ([mounts.go:43](../../internal/filesystem/mounts.go#L43)).
2. Bind-mounts `newRoot` onto itself so it is a mountpoint (a `pivot_root`
   requirement) ([mounts.go:48](../../internal/filesystem/mounts.go#L48)).
3. Calls `pivot_root(newRoot, putOld)`, `chdir("/")`, then
   `MNT_DETACH`-unmounts and removes the old root at `/.pivot_root`.

A new `/proc` is mounted afterward inside the PID namespace
([MountProc, mounts.go:79-89](../../internal/filesystem/mounts.go#L79-L89)). The
code carries an explicit security comment to this effect.

## Rejected alternative

**`chroot()`.** One line instead of a brittle multi-step sequence, and no
self-bind-mount requirement. Rejected because it is **not an isolation
boundary**: it is escapable via retained file descriptors to the old root, so
shipping it would teach the wrong lesson and present a security guarantee the
syscall does not provide. `pivot_root` fully detaches the old root, making those
descriptors unreachable.

## Consequences

**What we commit to:**
- A more complex, multi-step sequence with **no rollback**: if an intermediate
  step fails, the process is left partially pivoted and is expected to simply
  exit, tearing down its throwaway namespaces
  ([mounts.go:43-71](../../internal/filesystem/mounts.go#L43-L71)). Acceptable
  for single-process container init; it would leak mount state in a long-lived
  `Runtime` path.
- The new root must already be a mountpoint, hence the self-bind-mount step.

**What becomes harder:**
- The container filesystem today is **`rootfs` + `/proc` only**. `MountSys`
  (read-only `/sys`), `MountTmpfs` (`/tmp`), and device-node creation
  (`/dev/null`, `/dev/urandom`, …) are `TODO(M3.3)` stubs
  ([mounts.go:93-102](../../internal/filesystem/mounts.go#L93-L102),
  [internal/filesystem/devices.go](../../internal/filesystem/devices.go)), so
  most real userland will not run yet.

## Notes for reviewers

The mount-propagation step is not incidental: without `MS_REC | MS_PRIVATE`,
mounts performed inside the container would propagate back to the host mount
namespace. `TestHostMountsUnaffected` guards this (threat T-1).
