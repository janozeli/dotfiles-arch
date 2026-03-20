#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."
stow --dotfiles --no-folding -t "$HOME" .
