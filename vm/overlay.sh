#!/usr/bin/env bash
# overlay.sh — Create (or recreate) a qcow2 overlay on top of the base image.
#
# The overlay captures all writes. Deleting it resets the VM to the
# pristine post-cloud-init state.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
VM_DIR="${SCRIPT_DIR}/.state"
BASE_IMG="${VM_DIR}/base.qcow2"
OVERLAY_IMG="${VM_DIR}/overlay.qcow2"

if [[ ! -f "$BASE_IMG" ]]; then
  echo "ERROR: Base image not found. Run ./setup.sh first." >&2
  exit 1
fi

if [[ -f "$OVERLAY_IMG" ]]; then
  echo ">>> Removing existing overlay..."
  rm -f "$OVERLAY_IMG"
fi

echo ">>> Creating new overlay backed by base.qcow2..."
qemu-img create -f qcow2 -b "$BASE_IMG" -F qcow2 "$OVERLAY_IMG"

echo ""
echo "=== Overlay ready ==="
echo "  Overlay : $OVERLAY_IMG"
echo "  Backing : $BASE_IMG"
echo ""
echo "Next: run ./start.sh to boot the VM."
