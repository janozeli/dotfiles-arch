#!/usr/bin/env bash
# test-runner.sh — Full cycle: reset VM → boot → run installer → collect result.
#
# Usage:
#   ./test-runner.sh                  # run all units
#   ./test-runner.sh -unit packages   # run a single unit
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SSH_PORT="${SSH_PORT:-2222}"
RUNNER_ARGS="${*:---verbose}"

ssh_cmd() {
  ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    -p "$SSH_PORT" arch@localhost "$@"
}

wait_for_ssh() {
  echo ">>> Waiting for SSH..."
  local attempts=0
  while ! ssh_cmd true 2>/dev/null; do
    attempts=$((attempts + 1))
    if [[ $attempts -ge 60 ]]; then
      echo "ERROR: SSH not available after 60 attempts." >&2
      exit 1
    fi
    sleep 2
  done
  echo ">>> SSH ready."
}

# ── 1. Reset to clean state ────────────────────────────────────────────
"$SCRIPT_DIR/reset.sh"

# ── 2. Boot VM in background ───────────────────────────────────────────
"$SCRIPT_DIR/start.sh" --daemon

# ── 3. Wait for cloud-init + SSH ───────────────────────────────────────
wait_for_ssh

# ── 4. Mount shared repo inside guest ──────────────────────────────────
echo ">>> Mounting dotfiles repo in guest..."
ssh_cmd "sudo mkdir -p /mnt/dotfiles && sudo mount -t 9p -o trans=virtio dotfiles /mnt/dotfiles"

# ── 5. Run the installer ───────────────────────────────────────────────
echo ">>> Running installer with args: ${RUNNER_ARGS}"
echo "─────────────────────────────────────────────────"
ssh_cmd "cd /mnt/dotfiles && sudo bash install.sh ${RUNNER_ARGS}"
EXIT_CODE=$?
echo "─────────────────────────────────────────────────"

# ── 6. Report ──────────────────────────────────────────────────────────
if [[ $EXIT_CODE -eq 0 ]]; then
  echo ">>> Runner finished successfully."
else
  echo ">>> Runner exited with code ${EXIT_CODE}."
fi

echo ""
echo "VM still running. Use ./ssh.sh to inspect, or ./reset.sh to start over."
exit $EXIT_CODE
