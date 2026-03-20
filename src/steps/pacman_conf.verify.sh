#!/usr/bin/env bash
grep -q "^Color" /etc/pacman.conf && grep -q "^ParallelDownloads" /etc/pacman.conf
