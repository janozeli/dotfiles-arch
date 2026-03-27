#!/usr/bin/env bash
# start.sh — Boot the Arch VM from the overlay image.
#
# The repo is shared into the guest via virtio-9p at /mnt/dotfiles.
# SSH is forwarded to host port 2222.
#
# Usage:
#   ./start.sh            # foreground with serial console
#   ./start.sh --daemon   # background (use ssh to connect)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
VM_DIR="${SCRIPT_DIR}/.state"
OVERLAY_IMG="${VM_DIR}/overlay.qcow2"
SEED_IMG="${VM_DIR}/seed.img"
PID_FILE="${VM_DIR}/vm.pid"

SSH_PORT="${SSH_PORT:-2222}"
RAM="${RAM:-2G}"
CPUS="${CPUS:-2}"

if [[ ! -f "$OVERLAY_IMG" ]]; then
  echo "ERROR: Overlay not found. Run ./overlay.sh first." >&2
  exit 1
fi

DAEMON=0
if [[ "${1:-}" == "--daemon" ]]; then
  DAEMON=1
fi

# Kill any existing VM on the same pidfile
if [[ -f "$PID_FILE" ]] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  echo ">>> Stopping existing VM (pid $(cat "$PID_FILE"))..."
  kill "$(cat "$PID_FILE")" || true
  sleep 1
fi

QEMU_ARGS=(
  -enable-kvm
  -m "$RAM"
  -smp "$CPUS"
  -cpu host

  # Disks
  -drive "file=${OVERLAY_IMG},format=qcow2,if=virtio"
  -drive "file=${SEED_IMG},format=raw,if=virtio,readonly=on"

  # Share the repo into the guest
  -virtfs "local,path=${REPO_DIR},mount_tag=dotfiles,security_model=mapped-xattr,id=dotfiles"

  # Network — forward SSH
  -nic "user,hostfwd=tcp::${SSH_PORT}-:22"

  # Misc
  -pidfile "$PID_FILE"
)

if [[ "$DAEMON" -eq 1 ]]; then
  QEMU_ARGS+=( -daemonize -display none )
  echo ">>> Starting VM in background (ssh port ${SSH_PORT})..."
  qemu-system-x86_64 "${QEMU_ARGS[@]}"
  echo ">>> VM started. Connect with:  ssh -p ${SSH_PORT} arch@localhost"
else
  QEMU_ARGS+=( -nographic -serial mon:stdio )
  echo ">>> Booting VM (ssh port ${SSH_PORT})..."
  echo ">>> Exit serial console with: Ctrl-A X"
  echo ""
  qemu-system-x86_64 "${QEMU_ARGS[@]}"
fi
