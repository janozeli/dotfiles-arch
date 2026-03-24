#!/usr/bin/env bash
systemctl is-active --quiet docker \
    && systemctl is-active --quiet fstrim.timer \
    && systemctl is-active --quiet ufw \
    && systemctl --user is-active --quiet spotifyd
