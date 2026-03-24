#!/usr/bin/env bash
flatpak info com.spotify.Client >/dev/null 2>&1 || exit 1
