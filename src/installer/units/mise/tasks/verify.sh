#!/usr/bin/env bash
command -v mise >/dev/null \
    && mise which node >/dev/null 2>&1 \
    && mise which uv >/dev/null 2>&1 \
    && mise which pnpm >/dev/null 2>&1
