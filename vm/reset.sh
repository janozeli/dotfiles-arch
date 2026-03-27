#!/usr/bin/env bash
# reset.sh — Destroy the overlay and recreate it, resetting the VM to
# its initial state. This is the "repeat test" shortcut.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
VM_DIR="${SCRIPT_DIR}/.state"
PID_FILE="${VM_DIR}/vm.pid"

# Stop VM if running
if [[ -f "$PID_FILE" ]] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  echo ">>> Stopping running VM..."
  kill "$(cat "$PID_FILE")" || true
  sleep 1
  rm -f "$PID_FILE"
fi

echo ">>> Resetting overlay..."
"$SCRIPT_DIR/overlay.sh"

echo ""
echo "=== VM reset to initial state ==="
echo "Run ./start.sh to boot a fresh instance."
