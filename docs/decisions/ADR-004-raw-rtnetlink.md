# ADR-004 — Hand-rolled rtnetlink over a netlink library or shelling out to `ip(8)`

**Status:** Accepted
**Date:** 2026-06-01

## Context

A fresh network namespace starts with **zero interfaces — not even loopback**.
Before the container's entrypoint runs, `lo` must be brought up, and later
milestones need veth pairs, a bridge, addresses, and routes. Configuring
interfaces means talking to the kernel's rtnetlink (`AF_NETLINK`,
`NETLINK_ROUTE`). Three ways to do it:

1. **A netlink library** — e.g. `github.com/vishvananda/netlink`, the
   de-facto Go choice (and a study reference for this project).
2. **Shell out to `ip(8)`** — `ip link set lo up`, etc.
3. **Construct the netlink messages by hand** over a raw socket.

## Decision

**Hand-roll rtnetlink.** `SetupLoopback` opens an `AF_NETLINK`/`SOCK_RAW`
socket and constructs an `RTM_NEWLINK` message (`NlMsghdr` + `IfInfomsg`,
serialized via `unsafe`), sends it, and parses the ACK by sequence number to
bring `lo` up ([internal/namespace/namespace.go:39-106](../../internal/namespace/namespace.go#L39-L106),
ACK loop at [113-152](../../internal/namespace/namespace.go#L113-L152)).

## Rejected alternative

**`vishvananda/netlink`** — the obvious, ergonomic choice — was rejected because
it violates the project's **stdlib-only constraint** (`go.mod` has no `require`
block, no `go.sum`), which is itself a goal: the point is to see the netlink
wire format, not to call a function that hides it.

**Shelling out to `ip(8)`** was rejected because it adds a *runtime* dependency
on iproute2 — and `ip` would not exist inside a minimal container's mount
namespace anyway. Raw netlink keeps everything in-process and dependency-free.

## Consequences

**What we commit to:**
- Manual message construction with `unsafe.Pointer` serialization that is
  **sensitive to syscall struct layout** across Go versions and architectures.
- The same hand-rolled pattern must be repeated for the (still-unimplemented)
  bridge/veth/address/route code in `internal/network` — significant surface
  area that a library would have absorbed.

**What becomes harder:**
- The ACK read loop has **no receive timeout**
  ([namespace.go:113-152](../../internal/namespace/namespace.go#L113-L152)), so
  `SetupLoopback` blocks indefinitely if the kernel never ACKs. A
  `SO_RCVTIMEO` deadline is the obvious hardening.
- All of `internal/network` (bridge, veth, IPAM, NAT) is currently a stub; the
  rtnetlink groundwork in `linkSetUp` is the template those milestones will
  extend, but the cost of "no library" is paid again there.

## Notes for reviewers

`linkSetUp` is deliberately written to generalize: the `NlMsghdr` +
`appendStructBytes` + `readNetlinkAck` plumbing is the reusable core, and
`RTM_NEWLINK` for `lo` is the smallest real use of it. The honest trade is
*pedagogical transparency and zero deps* in exchange for *more code and more
ways to get the wire format wrong* — a trade that only makes sense for a
learning project, not a production runtime.
