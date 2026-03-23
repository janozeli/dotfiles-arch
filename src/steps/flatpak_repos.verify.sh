#!/usr/bin/env bash
command -v flatpak >/dev/null 2>&1 || exit 1
flatpak remote-list | grep -q flathub || exit 1
