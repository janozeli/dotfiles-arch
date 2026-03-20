#!/usr/bin/env bash
systemctl is-active --quiet docker \
    && systemctl is-active --quiet fstrim.timer
