#!/usr/bin/env bash
set -euo pipefail

gpu=$(lspci -nn | grep -i 'vga\|3d\|display')

if echo "$gpu" | grep -qi 'nvidia'; then
    echo "NVIDIA detectada, instalando drivers proprietários..."
    sudo pacman -S --noconfirm --needed nvidia-open-dkms nvidia-utils nvidia-settings
elif echo "$gpu" | grep -qi 'amd\|radeon'; then
    echo "AMD detectada, drivers open-source (mesa) já cobertos."
elif echo "$gpu" | grep -qi 'intel'; then
    echo "Intel detectada, drivers open-source (mesa) já cobertos."
else
    echo "GPU não identificada: $gpu"
    exit 1
fi
