#!/usr/bin/env bash
set -euo pipefail

sudo systemctl enable --now fstrim.timer
sudo systemctl enable --now docker.service
sudo usermod -aG docker "$USER"
sudo systemctl enable --now ufw.service
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw enable
systemctl --user enable --now spotifyd.service
