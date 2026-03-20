#!/usr/bin/env bash
getent passwd "$USER" | grep -q "/zsh$"
