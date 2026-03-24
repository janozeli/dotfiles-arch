#!/usr/bin/env bash
set -euo pipefail

sudo pacman -S --noconfirm --needed git base-devel
rm -rf /tmp/yay
git clone https://aur.archlinux.org/yay.git /tmp/yay
(cd /tmp/yay && makepkg -si --noconfirm)
rm -rf /tmp/yay
yay -Y --gendb
yay -Syu --noconfirm --devel
yay -Y --devel --save
