#!/usr/bin/env bash
set -euo pipefail

curl https://mise.jdx.dev/install.sh | sh
eval "$("$HOME/.local/bin/mise" activate bash)"
mise install node@lts uv@latest pnpm@latest
mise use --global node@lts uv@latest pnpm@latest
