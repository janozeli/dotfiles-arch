#!/usr/bin/env bash
# ssh.sh — Quick SSH into the running VM.
set -euo pipefail

SSH_PORT="${SSH_PORT:-2222}"

exec ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  -p "$SSH_PORT" arch@localhost "$@"
