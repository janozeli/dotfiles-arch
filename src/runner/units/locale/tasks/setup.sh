#!/usr/bin/env bash
set -euo pipefail

sudo sed -i 's/^#pt_BR.UTF-8/pt_BR.UTF-8/' /etc/locale.gen
sudo locale-gen
sudo localectl set-locale LANG=pt_BR.UTF-8
sudo localectl set-x11-keymap us,br pc105 altgr-intl,
