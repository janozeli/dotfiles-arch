#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
stow --dotfiles --no-folding -t "$HOME" .
