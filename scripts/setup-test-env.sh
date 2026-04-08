#!/bin/bash
# Setup script for the myruntime test environment.
# Run this once before running integration tests.
#
# What this does:
# 1. Downloads an Alpine Linux minimal rootfs for testing
# 2. Creates runtime data directories
# 3. Sets up the cgroup v2 hierarchy
# 4. Checks for required tools
#
# Requirements:
# - Linux kernel 5.8+ (for full cgroup v2 support)
# - cgroup v2 unified hierarchy mounted at /sys/fs/cgroup
# - Root privileges (for cgroup and mount operations)
# - curl, tar, iptables, iproute2

set -euo pipefail

RUNTIME_ROOT="/var/lib/myruntime"
TESTDATA_DIR="$(dirname "$0")/../testdata"

echo "=== myruntime test environment setup ==="

# Check kernel version
KVER=$(uname -r | cut -d. -f1-2)
echo "Kernel version: $(uname -r)"

# Check cgroup v2
if [ ! -f /sys/fs/cgroup/cgroup.controllers ]; then
    echo "ERROR: cgroup v2 not available."
    echo "  Ensure your kernel is booted with systemd.unified_cgroup_hierarchy=1"
    echo "  or cgroup_no_v1=all on the kernel command line."
    exit 1
fi
echo "Cgroup v2: OK (controllers: $(cat /sys/fs/cgroup/cgroup.controllers))"

# Check required tools
for tool in iptables ip curl tar; do
    if ! command -v "$tool" &>/dev/null; then
        echo "ERROR: $tool not found. Install it first."
        exit 1
    fi
done
echo "Required tools: OK"

# Download Alpine minirootfs for testing
mkdir -p "$TESTDATA_DIR/rootfs"
ALPINE_VERSION="3.19.0"
ALPINE_TAR="$TESTDATA_DIR/alpine-minirootfs-${ALPINE_VERSION}-x86_64.tar.gz"

if [ ! -f "$ALPINE_TAR" ]; then
    echo "Downloading Alpine Linux ${ALPINE_VERSION} minirootfs..."
    curl -L "https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-${ALPINE_VERSION}-x86_64.tar.gz" \
        -o "$ALPINE_TAR"
fi

if [ ! -d "$TESTDATA_DIR/rootfs/bin" ]; then
    echo "Extracting rootfs..."
    tar -xzf "$ALPINE_TAR" -C "$TESTDATA_DIR/rootfs"
fi
echo "Test rootfs: OK"

# Create runtime directories
sudo mkdir -p "${RUNTIME_ROOT}"/{images,layers,containers}
sudo chown -R "$(id -u):$(id -g)" "$RUNTIME_ROOT"
echo "Runtime directories: OK"

# Setup cgroup hierarchy
if [ ! -d /sys/fs/cgroup/myruntime ]; then
    sudo mkdir -p /sys/fs/cgroup/myruntime
    echo "+cpu +memory +io +pids" | sudo tee /sys/fs/cgroup/cgroup.subtree_control >/dev/null
    echo "Cgroup hierarchy: created"
else
    echo "Cgroup hierarchy: already exists"
fi

echo ""
echo "=== Setup complete ==="
echo "You can now run: make test-integration"
