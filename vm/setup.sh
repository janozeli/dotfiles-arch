#!/usr/bin/env bash
# setup.sh — Download Arch Linux cloud image and prepare the base qcow2 disk.
#
# This creates a "golden" base image that is never booted directly.
# All boots go through an overlay (see overlay.sh).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
VM_DIR="${SCRIPT_DIR}/.state"
BASE_IMG="${VM_DIR}/base.qcow2"
SEED_IMG="${VM_DIR}/seed.img"
ARCH_IMG_URL="https://geo.mirror.pkgbuild.com/images/latest/Arch-Linux-x86_64-cloudimg.qcow2"

DISK_SIZE="${DISK_SIZE:-20G}"

mkdir -p "$VM_DIR"

# ── 1. Download cloud image ────────────────────────────────────────────
if [[ ! -f "$BASE_IMG" ]]; then
  echo ">>> Downloading Arch Linux cloud image..."
  curl -fL -o "$BASE_IMG.tmp" "$ARCH_IMG_URL"
  mv "$BASE_IMG.tmp" "$BASE_IMG"
  echo ">>> Resizing disk to ${DISK_SIZE}..."
  qemu-img resize "$BASE_IMG" "$DISK_SIZE"
else
  echo ">>> Base image already exists, skipping download."
fi

# ── 2. Create cloud-init seed ISO ──────────────────────────────────────
if [[ ! -f "$SEED_IMG" ]]; then
  echo ">>> Creating cloud-init seed..."

  CIDATA_DIR=$(mktemp -d)
  trap 'rm -rf "$CIDATA_DIR"' EXIT

  cat > "$CIDATA_DIR/meta-data" <<'YAML'
instance-id: dotfiles-test-vm
local-hostname: archvm
YAML

  cat > "$CIDATA_DIR/user-data" <<'YAML'
#cloud-config
users:
  - name: arch
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    lock_passwd: false
    plain_text_passwd: arch
    ssh_authorized_keys: []

package_update: true
packages:
  - git
  - base-devel
  - openssh

runcmd:
  - systemctl enable --now sshd
YAML

  # Try cloud-localds (from cloud-image-utils) or fall back to genisoimage/mkisofs
  if command -v cloud-localds &>/dev/null; then
    cloud-localds "$SEED_IMG" "$CIDATA_DIR/user-data" "$CIDATA_DIR/meta-data"
  elif command -v genisoimage &>/dev/null; then
    genisoimage -output "$SEED_IMG" -volid cidata -joliet -rock \
      "$CIDATA_DIR/user-data" "$CIDATA_DIR/meta-data"
  elif command -v mkisofs &>/dev/null; then
    mkisofs -output "$SEED_IMG" -volid cidata -joliet -rock \
      "$CIDATA_DIR/user-data" "$CIDATA_DIR/meta-data"
  elif command -v xorriso &>/dev/null; then
    xorriso -as mkisofs -o "$SEED_IMG" -volid cidata -joliet -rock \
      "$CIDATA_DIR/user-data" "$CIDATA_DIR/meta-data"
  else
    echo "ERROR: Need cloud-localds, genisoimage, mkisofs, or xorriso to create seed ISO." >&2
    exit 1
  fi

  echo ">>> Seed ISO created."
else
  echo ">>> Seed ISO already exists, skipping."
fi

echo ""
echo "=== Base setup complete ==="
echo "  Base image : $BASE_IMG"
echo "  Seed ISO   : $SEED_IMG"
echo ""
echo "Next: run ./overlay.sh to create a disposable overlay."
