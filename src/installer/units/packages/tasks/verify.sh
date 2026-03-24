#!/usr/bin/env bash
packages=(
    stow zsh git less wget curl unzip
    fzf zoxide eza bat tldr
    kitty github-cli wl-clipboard xclip
    mpv yt-dlp
    zed firefox zen-browser-bin google-chrome
    ttf-firacode-nerd ttf-jetbrains-mono-nerd
    docker docker-compose teams-for-linux-bin
    flatpak
    visual-studio-code-bin obsidian-bin ufw
    snap-pac
)

for pkg in "${packages[@]}"; do
    yay -Q "$pkg" >/dev/null 2>&1 || exit 1
done
