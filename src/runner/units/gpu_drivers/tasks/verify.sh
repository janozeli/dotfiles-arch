#!/usr/bin/env bash

gpu=$(lspci -nn | grep -i 'vga\|3d\|display')

if echo "$gpu" | grep -qi 'nvidia'; then
    pacman -Qi nvidia-open-dkms &>/dev/null && pacman -Qi nvidia-utils &>/dev/null
elif echo "$gpu" | grep -qi 'amd\|radeon\|intel'; then
    # open-source drivers handled by mesa, always OK
    true
else
    exit 1
fi
