# ADR-006 — All six namespaces created together in one clone, config ignored

**Status:** Accepted
**Date:** 2026-06-01

## Context

A container's isolation is the set of namespaces it gets. A runtime can:

1. **Always create a fixed full set** of namespaces.
2. **Let configuration select** which namespaces to create (opt in/out per
   namespace — e.g. host-network mode, shared IPC), and/or **create the user
   namespace first** in a separate clone, as some `runc` paths do, so UID/GID
   maps can be written before the other namespaces are entered.

## Decision

Create **all six namespaces together, always**: `CloneFlags` returns
`CLONE_NEWUSER | CLONE_NEWPID | CLONE_NEWNS | CLONE_NEWUTS | CLONE_NEWIPC |
CLONE_NEWNET` and explicitly discards its `*ContainerConfig` argument (`_ =
config`) ([internal/namespace/namespace.go:20-23](../../internal/namespace/namespace.go#L20-L23)).
Every container gets the same full-isolation surface, set in a single clone via
`SysProcAttr.Cloneflags`.

## Rejected alternative

**Phased or configurable namespace selection** (including a user-namespace-first
clone). It is strictly more flexible — host-network mode, shared-namespace
"sidecar" patterns, and clean ordering of UID-map writes all become possible.
Rejected for now because a single, monolithic full-isolation contract means the
(future) `Start` always reasons about the *same* six namespaces with **no
partial-isolation states** to get wrong — the simplest correct mental model for
a learning runtime, and it matches the default Docker/`runc` isolation surface.

## Consequences

**What we commit to:**
- **No knob** for host-network, shared-IPC, or any partial-isolation mode.
- The **user namespace is created in the same clone as the others** rather than
  first. This constrains the ordering of UID/GID-map writes: the maps must be
  written from the parent against the child's `/proc/<pid>` after the clone but
  before the entrypoint execs.

**What becomes harder — and the gap it produced:**
- Because there is no separate userns-first step and no parent↔child
  synchronization yet (see [ADR-001](ADR-001-reexec-init.md)), the live `run`
  path sets `CLONE_NEWUSER` but **never writes UID/GID maps**
  ([cmd/myruntime/main.go:93-139](../../cmd/myruntime/main.go#L93-L139) — no
  `WriteIDMappings` call), even though `WriteIDMappings`/`DefaultUIDMappings`
  are fully implemented and unit-tested
  ([internal/namespace/userns.go:28-89](../../internal/namespace/userns.go#L28-L89)).
  A container therefore runs with an **unmapped** user namespace (IDs appear as
  `nobody`), not the intended `0→100000` mapping. This is the project's most
  important known gap — see [threat model EP-1](../threat-model.md) and
  [design.md "Known limitations"](../design.md).
- Config-driven ID mappings are impossible without changing the `CloneFlags`
  signature contract.

## Notes for reviewers

This ADR and [ADR-001](ADR-001-reexec-init.md) interact: the "always six,
combined" choice plus the not-yet-built parent/child pipe are *together* why
rootless isolation is implemented in `userns.go` but not delivered on the
running path. The fix is mechanical (write the maps from the parent after fork,
before exec), not a redesign.
