#!/usr/bin/env bash
claude mcp list 2>/dev/null | grep -q "exa" \
    && claude mcp list 2>/dev/null | grep -q "context7"
