#!/usr/bin/env bash
set -euo pipefail

sudo sed -i 's/^#\?Color.*/Color/' /etc/pacman.conf
sudo sed -i 's/^#\?ParallelDownloads.*/ParallelDownloads = 15/' /etc/pacman.conf
if ! grep -q '^ILoveCandy' /etc/pacman.conf; then
    sudo sed -i '/^ParallelDownloads/a ILoveCandy' /etc/pacman.conf
fi
