# ADR-005 — Per-container JSON state files on disk over an embedded database

**Status:** Accepted
**Date:** 2026-06-01

## Context

A container runtime that is a CLI (not a daemon) has no in-memory process to
hold state between invocations. `myruntime ps`, `inspect`, `stop`, and
crash-reconciliation all need to know what containers exist, their state
(`created`/`running`/`stopped`), host PID, rootfs, and cgroup path **after the
creating process has exited**. Options:

1. **Per-container JSON files** on disk (`<root>/<id>/state.json`).
2. **An embedded database** — SQLite or bolt.
3. **etcd** or another external store.
4. **No persistence** — reconstruct state on demand by scanning `/proc` and the
   cgroup tree.

## Decision

Persist each container as **JSON at `<root>/<id>/state.json`**, one directory
per container. The `Container` struct (`ID`, `State`, host `Pid`, `Config`,
`CreatedAt`/`StartedAt`/`FinishedAt`, `ExitCode`, `RootFS`, `CgroupPath`) is
directly serializable and is the contract in
[pkg/specs/config.go](../../pkg/specs/config.go). The store API
(`Save`/`Load`/`List`/`Delete`) lives in
[internal/store/state.go](../../internal/store/state.go).

> **Status:** this is a **contract, not yet an implementation** — every store
> method currently returns `nil`/`nil,nil`
> ([state.go:19-47](../../internal/store/state.go#L19-L47)). The decision is
> recorded now because it shapes the lifecycle work in M5.

## Rejected alternative

**An embedded database (SQLite/bolt).** It gives transactions, indexed queries,
and a single file to manage. Rejected because: it is a third-party dependency
(stdlib-only constraint), it is overkill for a flat list of containers a human
will `cat`/`jq`, and it hides state that should be trivially inspectable in a
learning project. **etcd** is daemon-scale infrastructure for a single-host CLI.
**Memory-only** can't survive the creating process exiting, which defeats `ps`
of stopped containers and crash reconciliation (M8.3).

Per-container files also give **atomic single-file replacement** (write-temp +
rename) and independent updates without lock contention between containers.

## Consequences

**What we commit to:**
- **No cross-file transactions.** Create/Delete are multi-step (cgroup +
  overlay + network + state); a crash mid-sequence can leave orphaned state
  vs. resources, which is exactly what crash reconciliation (M8.3) must repair.
- Per-container atomicity relies on a **write-then-rename + `fsync`** discipline
  that is documented but **not yet coded** — and the store is not symlink-safe
  (no `O_NOFOLLOW`), so a tampered `state.json` is a real concern
  (threat T-4).
- Absolute `RootFS`/`CgroupPath` strings in the persisted struct **couple state
  to the on-disk layout** — moving the root invalidates saved state.

**What becomes harder:**
- Querying "all running containers" is an O(n) directory walk + parse rather
  than an indexed lookup — fine at CLI scale, but it is a scan.

## Notes for reviewers

The decision follows the Unix philosophy deliberately: state you can read with
`cat <root>/<id>/state.json` is a feature for a system whose purpose is to be
understood. The cost — no transactions, manual atomicity — is acceptable for a
single-host CLI and is called out so the M5 implementation gets the
write-then-rename discipline right rather than discovering the need after a
torn write.
