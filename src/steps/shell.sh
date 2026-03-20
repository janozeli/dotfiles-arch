#!/usr/bin/env bash
set -euo pipefail

sudo chsh -s "$(which zsh)" "$USER"
